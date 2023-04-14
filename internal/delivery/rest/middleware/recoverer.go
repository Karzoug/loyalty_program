package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func Recoverer(logger *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if err == http.ErrAbortHandler {
						logger.Debug("not recovered by middleware", zap.Error(http.ErrAbortHandler))
						panic(err)
					}

					logger.Error("recovering from panic", zap.Any("error", err), zap.Stack("stacktrace"))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
}
