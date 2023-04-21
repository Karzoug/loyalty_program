package rest

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Karzoug/loyalty_program/internal/delivery/rest/helper"
	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/service"
	"github.com/goccy/go-json"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var (
	ErrInvalidSum = errors.New("invalid sum")
)

type withdrawResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (s *server) listUserWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), handlerTimeout)
	defer cancel()

	var hErr *helper.HandlerError
	login, err := helper.GetLoginFromJWTInContext(ctx, s.logger)
	if err != nil {
		if errors.As(err, &hErr) {
			helper.WriteJSONError(w, hErr.Message, hErr.Code, s.logger)
		} else {
			s.logger.Error("List user withdrawals handler: get login from context error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	ws, err := s.service.ListUserWithdrawals(ctx, *login)
	if err != nil {
		s.logger.Error("List user withdrawals handler: list withdrawals service error", zap.Error(err))
		helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		return
	}

	if len(ws) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	withdrawalsResp := make([]withdrawResponse, 0, len(ws))
	for _, w := range ws {
		withdrawalsResp = append(withdrawalsResp, withdrawResponse{
			Order:       strconv.FormatInt(int64(w.OrderNumber), 10),
			Sum:         w.Sum.InexactFloat64(),
			ProcessedAt: w.ProcessedAt,
		})
	}
	sort.Slice(withdrawalsResp, func(i, j int) bool {
		return withdrawalsResp[i].ProcessedAt.Before(withdrawalsResp[j].ProcessedAt)
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(withdrawalsResp); err != nil {
		s.logger.Error("List user withdrawals handler: encode json response error", zap.Error(err))
		helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		return
	}
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (r withdrawRequest) validate() error {
	if r.Sum <= 0 {
		return ErrInvalidSum
	}
	return nil
}

func (s *server) createWithdrawHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), handlerTimeout)
	defer cancel()

	var hErr *helper.HandlerError
	login, err := helper.GetLoginFromJWTInContext(ctx, s.logger)
	if err != nil {
		if errors.As(err, &hErr) {
			helper.WriteJSONError(w, hErr.Message, hErr.Code, s.logger)
		} else {
			s.logger.Error("Create withdraw handler: get login from context error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	var withdrawReq withdrawRequest
	err = helper.DecodeJSON(r, &withdrawReq)
	if err != nil {
		if errors.As(err, &hErr) {
			helper.WriteJSONError(w, hErr.Message, hErr.Code, s.logger)
		} else {
			s.logger.Error("Create withdraw handler: decode withdraw request from JSON error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	if err := withdrawReq.validate(); err != nil {
		helper.WriteJSONError(w, err.Error(), http.StatusBadRequest, s.logger)
		return
	}
	number, err := strconv.ParseInt(withdrawReq.Order, 10, 64)
	if err != nil {
		helper.WriteJSONError(w, service.ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity, s.logger)
		return
	}
	orderNumber := order.Number(number)

	_, err = s.service.CreateWithdraw(ctx, *login, orderNumber, decimal.NewFromFloat(withdrawReq.Sum))
	if err != nil {
		switch err {
		case service.ErrInvalidAuthData:
			helper.WriteJSONError(w, err.Error(), http.StatusUnauthorized, s.logger)
		case service.ErrInvalidOrderNumber:
			helper.WriteJSONError(w, err.Error(), http.StatusUnprocessableEntity, s.logger)
		case service.ErrInsufficientBalance:
			helper.WriteJSONError(w, err.Error(), http.StatusPaymentRequired, s.logger)
		default:
			s.logger.Error("Create withdraw handler: create withdraw service error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
