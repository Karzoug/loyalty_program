package main

import (
	"log"

	"github.com/Karzoug/loyalty_program/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type buildLoggerConfig interface {
	IsDebugMode() bool
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Read config error: %s", err)
	}

	logger, err := buildLogger(cfg)
	if err != nil {
		log.Fatalf("Create logger error: %s", err)
	}
	defer logger.Sync()
}

func buildLogger(cfg buildLoggerConfig) (*zap.Logger, error) {
	if cfg.IsDebugMode() {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return config.Build()
	}
	return zap.NewProduction()
}
