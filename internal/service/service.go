package service

import (
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"go.uber.org/zap"
)

type Service struct {
	storages storage.TxStorages
	logger   *zap.Logger
}

func New(storages storage.TxStorages, logger *zap.Logger) *Service {
	return &Service{
		storages: storages,
		logger:   logger,
	}
}
