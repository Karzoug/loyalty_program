package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var (
	// sugaredLogFormat is the format the Chi logs will use when
	// a sugared Zap logger is passed.
	sugaredLogFormat = `[%s] "%s %s %s" from %s - %s %dB in %s`
)

// Logger is a middleware that logs each request recieved using the provided Zap logger.
func Logger(l interface{}) func(next http.Handler) http.Handler {
	switch logger := l.(type) {
	case *zap.Logger:
		return func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
				t1 := time.Now()
				defer func() {
					logger.Info("served",
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Int("status", ww.Status()),
						zap.String("remoteAddr", r.RemoteAddr),
						zap.Duration("latency", time.Since(t1)),
						zap.Int("size", ww.BytesWritten()))
				}()
				next.ServeHTTP(ww, r)
			}
			return http.HandlerFunc(fn)
		}

	case *zap.SugaredLogger:
		return func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
				t1 := time.Now()
				defer func() {
					logger.Infof(sugaredLogFormat,
						r.Method,
						r.URL.Path,
						r.RemoteAddr,
						statusLabel(ww.Status()),
						ww.BytesWritten(),
						time.Since(t1),
					)
				}()
				next.ServeHTTP(ww, r)
			}
			return http.HandlerFunc(fn)
		}
	default:
		log.Fatalf("Unknown logger passed in. Please provide *Zap.Logger or *Zap.SugaredLogger")
	}
	return nil
}

func statusLabel(status int) string {
	switch {
	case status >= 100 && status < 300:
		return fmt.Sprintf("%d OK", status)
	case status >= 300 && status < 400:
		return fmt.Sprintf("%d Redirect", status)
	case status >= 400 && status < 500:
		return fmt.Sprintf("%d Client Error", status)
	case status >= 500:
		return fmt.Sprintf("%d Server Error", status)
	default:
		return fmt.Sprintf("%d Unknown", status)
	}
}
