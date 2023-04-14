package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/Karzoug/loyalty_program/internal/delivery/rest/middleware"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type serverConfig interface {
	RunAddress() string
}

type server struct {
	cfg    serverConfig
	logger *zap.Logger

	server *http.Server
}

func New(cfg serverConfig, logger *zap.Logger) server {
	return server{
		cfg:    cfg,
		logger: logger,

		server: &http.Server{Addr: cfg.RunAddress()},
	}
}

func (s *server) Run(ctx context.Context) error {
	s.logger.Info("Running http server", zap.String("address", s.cfg.RunAddress()))

	s.server.Handler = s.newRouter()

	go func() {
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Fatal("Server listen and serve error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	s.logger.Info("Shutting down http server")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctxShutdown)
}

func (s *server) newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger(s.logger))
	r.Use(middleware.Recoverer(s.logger))

	r.Post("/api/user/register", s.registerUserHandler)
	r.Post("/api/user/login", s.loginUserHandler)

	r.Post("/api/user/orders", s.createOrderHandler)
	r.Get("/api/user/orders", s.listUserOrdersHandler)
	r.Get("/api/user/balance", s.getUserBalanceHandler)
	r.Post("/api/user/balance/withdraw", s.createWithdrawHandler)
	r.Get("/api/user/withdrawals", s.listUserWithdrawalsHandler)

	return r
}
