package service

import (
	"context"
	"errors"
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/repository/processor"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	processUnprocessedOrdersGoroutineLimit = 10
	processUnprocessedOrdersStorageLimit   = 100
	processMaxWaitingDuration              = 90 * time.Second
)

// processOrder calls order processor to update status and accrual (if possible).
func (s *Service) processOrder(ctx context.Context, o order.Order) {
	ctx, cancel := context.WithTimeout(ctx, processMaxWaitingDuration)
	defer cancel()

	t1 := time.Now()
	procOrder, err := s.orderProcessor.Process(ctx, o)

	// order not found: delete (?)
	if errors.Is(err, processor.ErrOrderNotRegistered) {
		s.logger.Warn("Process order: order not registered in accrual service", zap.Int64("order number", int64(o.Number)))
		//s.logger.Warn("Process order: order not registered in accrual service and will be deleted", zap.Int64("order number", int64(o.Number)))
		//err := s.storages.Order().Delete(ctx, o.Number)
		// if err != nil {
		// 	s.logger.Error("Process order: order storage: delete order error", zap.Error(err))
		// }
		return
	}

	// no result received: process later again
	if err != nil {
		s.logger.Warn("Process order: no result received",
			zap.Int64("order number", int64(o.Number)),
			zap.Duration("processing time", time.Since(t1)))
		return
	}

	// got the same result as before: process later again
	if o.Status == procOrder.Status {
		s.logger.Debug("Process order: no new result received, status not changed",
			zap.Int64("order number", int64(o.Number)),
			zap.Duration("processing time", time.Since(t1)))
		return
	}

	// order status not 'processed': update only status
	if procOrder.Status != order.StatusProcessed {
		err := s.storages.Order().Update(ctx, *procOrder)
		if err != nil {
			s.logger.Error("Process order: order storage: update order status error", zap.Error(err))
		}
		return
	}

	if procOrder.Accrual.LessThanOrEqual(decimal.Zero) {
		s.logger.Error("Process order: got order with negative accrual value", zap.Float64("accrual", procOrder.Accrual.InexactFloat64()))
		return
	}

	// order status 'processed': update order and user balance inside transaction
	tx, err := s.storages.BeginTx(ctx)
	if err != nil {
		s.logger.Error("Process order: storages: begin transaction error", zap.Error(err))
		return
	}
	defer tx.Rollback(ctx)

	err = tx.Order().Update(ctx, *procOrder)
	if err != nil {
		s.logger.Error("Process order: order storage: update order error", zap.Error(err))
		return
	}
	_, err = tx.User().UpdateBalance(ctx, procOrder.UserLogin, procOrder.Accrual)
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

// processUnprocessedOrders searches for unprocessed orders in the storage and
// calls order processor to update status and accrual (if possible).
func (s *Service) processUnprocessedOrders(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		orders, err := s.storages.Order().
			ListUnprocessed(ctx, processUnprocessedOrdersStorageLimit, 0, time.Now().UTC().Add(-processMaxWaitingDuration))
		if err != nil {
			s.logger.Error("Process unprocessed orders: order storage error", zap.Error(err))
			return
		}
		if len(orders) == 0 {
			return
		}

		g, _ := errgroup.WithContext(ctx)
		g.SetLimit(processUnprocessedOrdersGoroutineLimit)
		for _, o := range orders {
			o := o
			g.Go(func() error {
				s.processOrder(ctx, o)
				return nil
			})
		}
		g.Wait()
	}
}
