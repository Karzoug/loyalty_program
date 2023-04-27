package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	morder "github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/repository/processor"
	"github.com/Karzoug/loyalty_program/pkg/e"
	"github.com/shopspring/decimal"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

var _ processor.Order = (*orderProcessor)(nil)

const (
	accrualURLPathFmt = "/api/orders/%d"
	maxAttemptNumber  = 3 // number of attempts to get a response from the server

	rateLimit = 1000 // requests count per second
	rateBurst = 30   // max burst of requests

	// only work if response status code is not 429 or 503
	backoffMinDuration = time.Second
	backoffMaxDuration = 10 * time.Second
)

var (
	errRequestNotSucceeded = errors.New("request not succeeded")
)

// accrual service specific response regexp
var (
	rpmRegExp *regexp.Regexp = regexp.MustCompile("([0-9]+) (?:requests per minute allowed)")
	rpcRegExp *regexp.Regexp = regexp.MustCompile("([0-9]+) (?:requests per second allowed)")
)

type orderProcessorConfig interface {
	AccrualSystemAddress() url.URL
}

type orderProcessor struct {
	cfg    orderProcessorConfig
	logger *zap.Logger

	client       *http.Client
	limiter      *rate.Limiter
	backOffUntil *atomic.Time
}

func NewOrderProcessor(cfg orderProcessorConfig, logger *zap.Logger) *orderProcessor {
	return &orderProcessor{
		cfg:    cfg,
		logger: logger,

		client:       &http.Client{Timeout: 3 * time.Second},
		limiter:      rate.NewLimiter(rate.Limit(rateLimit), rateBurst),
		backOffUntil: atomic.NewTime(time.Now()),
	}
}

// Process returns order data from the server.
func (p *orderProcessor) Process(ctx context.Context, o morder.Order) (*order.Order, error) {
	p.logger.Debug("Order processor: start order processing", zap.Int64("order number", int64(o.Number)))

	accrual, err := p.getOrderAccrual(ctx, o.Number)
	if err != nil {
		p.logger.Debug("Order processor: accrual service returns error", zap.Int64("order number", int64(o.Number)), zap.Error(err))
		return nil, err
	}

	p.logger.Debug("Order processor: accrual service returns order status", zap.Int64("order number", int64(o.Number)), zap.String("status", accrual.Status))
	switch accrual.Status {
	case `REGISTERED`: // TODO: find out the details about this status
		o.Status = morder.StatusNew
	case `INVALID`:
		o.Status = morder.StatusInvalid
	case `PROCESSING`:
		o.Status = morder.StatusProcessing
	case `PROCESSED`:
		o.Status = morder.StatusProcessed
		o.Accrual = decimal.NewFromFloat(accrual.Accrual)
	}
	return &o, nil
}

type orderAccrual struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// getOrderAccrual makes attempts to get order data from accrual service.
func (p *orderProcessor) getOrderAccrual(ctx context.Context, number morder.Number) (*orderAccrual, error) {
	url := p.cfg.AccrualSystemAddress()
	url.Path = fmt.Sprintf(accrualURLPathFmt, number)

	var body []byte
	for i := 1; i <= maxAttemptNumber; i++ {
		var (
			wait time.Duration
			err  error
		)

		// wait if the number of requests to the service was exceeded (global)
		<-time.After(time.Until(p.backOffUntil.Load()))

		// wait if current request limit is exceeded (local)
		if err := p.limiter.Wait(ctx); err != nil {
			return nil, err
		}

		p.logger.Debug("Order processor: do request to accrual service", zap.Int("attempt number", i), zap.String("url", url.String()))
		body, wait, err = p.doAttemptRequest(ctx, i, url)
		if err == nil {
			break // the request is succeeded
		}
		switch err {
		case processor.ErrOrderNotRegistered:
			return nil, processor.ErrOrderNotRegistered
		case errRequestNotSucceeded:
			// got response but status code indicates that the request was not succeeded
		default:
			p.logger.Warn("Order processor: do request error", zap.Error(err))
		}

		// sleep until next attempt or context canceled
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	if len(body) == 0 {
		return nil, processor.ErrServerNotRespond
	}

	var acc orderAccrual
	err := json.Unmarshal(body, &acc)
	if err != nil {
		return nil, e.Wrap("unmarshal to orderAccrual struct", err)
	}

	return &acc, nil
}

// doAttemptRequest does a request to the server and returns the received data or the duration until the next request and an error.
func (p *orderProcessor) doAttemptRequest(ctx context.Context, attemptNum int, url url.URL) ([]byte, time.Duration, error) {
	wait := p.backoff(backoffMinDuration, backoffMaxDuration, attemptNum, nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, wait, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, wait, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil, 0, processor.ErrOrderNotRegistered
	case http.StatusOK:
		b, err := io.ReadAll(resp.Body)
		return b, 0, err
	default:
		wait = p.backoff(backoffMinDuration, backoffMaxDuration, attemptNum, resp)
		return nil, wait, errRequestNotSucceeded
	}
}

// backoff returns the waiting duration until the next request.
// In the case when the status code is 429/503 trying to get information from the response -
// requests per second and retry after time and set them to the order processor.
func (p *orderProcessor) backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			if s, ok := resp.Header["Retry-After"]; ok {
				if rps, ok := tryGetLimitFromResponse(resp); ok {
					p.limiter.SetLimit(rate.Limit(rps)) // set new requests per second if possible
				}

				if sleep, err := strconv.ParseInt(s[0], 10, 64); err == nil {
					sleepDuration := time.Second * time.Duration(sleep)
					// make all other requests attempts wait until the end of the time returned
					p.backOffUntil.Store(time.Now().Add(sleepDuration))
					return sleepDuration
				}
			}
		}
	}

	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > max {
		sleep = max
	}
	return sleep

}

// tryGetLimitFromResponse tries to find in the server response the allowed requests per second.
func tryGetLimitFromResponse(resp *http.Response) (rps int, ok bool) {
	if resp.Body == nil {
		return 0, false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return 0, false
	}

	rpmStrings := rpmRegExp.FindStringSubmatch(string(body))
	if rpmStrings != nil || len(rpmStrings) >= 2 {
		rate, err := strconv.Atoi(rpmStrings[1])
		if err != nil {
			return 0, false
		}
		return 60 * rate, true
	}

	rpcStrings := rpcRegExp.FindStringSubmatch(string(body))
	if rpcStrings != nil || len(rpcStrings) >= 2 {
		rate, err := strconv.Atoi(rpcStrings[1])
		if err != nil {
			return 0, false
		}
		return rate, true
	}

	return 0, false
}
