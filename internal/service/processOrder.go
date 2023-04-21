package service

import (
	"context"
	"errors"
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/repository/processor"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const processMaxWaitingDuration = 3 * time.Minute

func (s *Service) processOrder(o order.Order) {
	ctx, cancel := context.WithTimeout(context.Background(), processMaxWaitingDuration)
	defer cancel()

	t1 := time.Now()
	ch := s.orderProcessor.Process(ctx, o)

	var result processor.AcrualOrderResult
	select {
	case result = <-ch:
	case <-ctx.Done():
	}

	// order not found: delete
	if errors.Is(result.Err, processor.ErrOrderNotRegistered) {
		s.logger.Warn("Process order: order not registred in accrual service and will be deleted", zap.Int64("order number", int64(o.Number)))
		err := s.storages.Order().Delete(ctx, o.Number)
		if err != nil {
			s.logger.Error("Process order: order storage: delete order error", zap.Error(err))
		}
		return
	}

	// no result received: process later again
	if result.Err != nil {
		s.logger.Warn("Process order: no result received",
			zap.Int64("order number", int64(o.Number)),
			zap.Duration("processing time", time.Since(t1)))
		return
	}

	// got the same result as before: process later again
	if o.Status == result.Order.Status {
		s.logger.Debug("Process order: no new result received, status not changed",
			zap.Int64("order number", int64(o.Number)),
			zap.Duration("processing time", time.Since(t1)))
		return
	}

	// order status not 'processed': update only status
	if result.Order.Status != order.StatusProcessed {
		err := s.storages.Order().Update(ctx, *result.Order)
		if err != nil {
			s.logger.Error("Process order: order storage: update order status error", zap.Error(err))
		}
		return
	}

	if result.Order.Accrual.LessThanOrEqual(decimal.Zero) {
		s.logger.Error("Process order: got order with negative accrual value", zap.Float64("accrual", result.Order.Accrual.InexactFloat64()))
		return
	}

	// order status 'processed': update order and user balance inside transaction
	tx, err := s.storages.BeginTx(ctx)
	if err != nil {
		s.logger.Error("Process order: storages: begin transaction error", zap.Error(err))
		return
	}
	defer tx.Rollback(ctx)

	err = tx.Order().Update(ctx, *result.Order)
	if err != nil {
		s.logger.Error("Process order: order storage: update order error", zap.Error(err))
		return
	}
	_, err = tx.User().UpdateBalance(ctx, result.Order.UserLogin, result.Order.Accrual)
	if err != nil {
		s.logger.Error("Process order: user storage: update user balance error", zap.Error(err))
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("Process order: storages: commit transaction error", zap.Error(err))
		return
	}
}
