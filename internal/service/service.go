package service

import (
	"context"
	"time"

	"github.com/Karzoug/loyalty_program/internal/repository/processor"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"go.uber.org/zap"
)

const (
	processUnprocessedOrdersInterval = 10 * time.Minute
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

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(processUnprocessedOrdersInterval)

	for {
		select {
		case <-ticker.C:
			go func() {
				ctx, cancel := context.WithTimeout(ctx, processUnprocessedOrdersInterval)
				defer cancel()

				s.processUnprocessedOrders(ctx)
			}()
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}
