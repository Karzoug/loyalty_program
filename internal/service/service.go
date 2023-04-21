package service

import (
	"github.com/Karzoug/loyalty_program/internal/repository/processor"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"go.uber.org/zap"
)

type Service struct {
	storages       storage.TxStorages
	orderProcessor processor.Order
	logger         *zap.Logger
}

func New(storages storage.TxStorages, proc processor.Order, logger *zap.Logger) *Service {
	return &Service{
		storages:       storages,
		orderProcessor: proc,
		logger:         logger,
	}
}
