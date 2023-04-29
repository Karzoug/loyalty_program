package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Karzoug/loyalty_program/internal/config"
	"github.com/Karzoug/loyalty_program/internal/delivery/rest"
	"github.com/Karzoug/loyalty_program/internal/repository/processor/accrual"
	"github.com/Karzoug/loyalty_program/internal/repository/storage/postgresql"
	"github.com/Karzoug/loyalty_program/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

type buildLoggerConfig interface {
	IsDebugMode() bool
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Read config error: %s", err)
	}

	logger, err := buildLogger(cfg)
	if err != nil {
		log.Fatalf("Create logger error: %s", err)
	}
	defer logger.Sync()

	storages, err := postgresql.NewStorages(ctx, cfg)
	if err != nil {
		logger.Fatal("Database error", zap.Error(err))
	}

	proc := accrual.NewOrderProcessor(cfg, logger)

	service := service.New(storages, proc, logger)

	g, _ := errgroup.WithContext(ctx)

	server := rest.New(cfg, service, logger)
	g.Go(func() error {
		err := server.Run(ctx)
		if err != nil {
			logger.Error("Server shutdown failed", zap.Error(err))
		}
		return err
	})

	g.Go(func() error {
		err := service.Run(ctx)
		if err != nil {
			logger.Error("Service stop failed", zap.Error(err))
		}
		return err
	})

	g.Wait()
}

func buildLogger(cfg buildLoggerConfig) (*zap.Logger, error) {
	if cfg.IsDebugMode() {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return config.Build()
	}
	return zap.NewProduction()
}
