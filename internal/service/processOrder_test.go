package service

import (
	"context"
	"testing"
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/processor"
	pmock "github.com/Karzoug/loyalty_program/internal/repository/processor/mock"
	smock "github.com/Karzoug/loyalty_program/internal/repository/storage/mock"
	"github.com/pioz/faker"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_processOrder(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	config := zap.NewDevelopmentConfig()
	logger, _ := config.Build()

	storages, err := smock.NewStorages(ctx)
	require.NoError(t, err)

	ch := make(chan processor.AcrualOrderResult, 1)
	proc := pmock.NewOrder(ch)

	service := New(storages, proc, logger)

	login := user.Login(faker.Username())
	password := faker.StringWithSize(15)
	_, err = service.RegisterUser(ctx, login, password)
	require.NoError(t, err)

	accrual := decimal.NewFromFloat(120)

	t.Run("positive", func(t *testing.T) {
		orderNumber := generateOrderNumber(t)

		o, err := order.New(orderNumber, login)
		require.NoError(t, err)

		err = service.storages.Order().Create(ctx, *o)
		require.NoError(t, err)

		procOrder := *o
		procOrder.Status = order.StatusProcessed
		procOrder.Accrual = accrual
		ch <- processor.AcrualOrderResult{
			Order: &procOrder,
			Err:   nil,
		}
		service.processOrder(ctx, *o)

		storageOrder, err := service.storages.Order().Get(ctx, o.Number)
		require.NoError(t, err)
		assert.Equal(t, procOrder.Number, storageOrder.Number)
		assert.Equal(t, procOrder.Status, storageOrder.Status)
		assert.True(t, procOrder.Accrual.RoundBank(4).Equal(storageOrder.Accrual.RoundBank(4)))

		u, err := service.storages.User().Get(ctx, login)
		require.NoError(t, err)
		assert.True(t, u.Balance.RoundBank(4).Equal(accrual.RoundBank(4)))
	})

	t.Run("server not respond", func(t *testing.T) {
		orderNumber := generateOrderNumber(t)

		o, err := order.New(orderNumber, login)
		require.NoError(t, err)

		err = service.storages.Order().Create(ctx, *o)
		require.NoError(t, err)

		ch <- processor.AcrualOrderResult{
			Order: o,
			Err:   processor.ErrServerNotRespond,
		}
		service.processOrder(ctx, *o)

		storageOrder, err := service.storages.Order().Get(ctx, o.Number)
		require.NoError(t, err)
		assert.Equal(t, o.Number, storageOrder.Number)
		assert.Equal(t, o.Status, storageOrder.Status)
		assert.True(t, o.Accrual.RoundBank(4).Equal(storageOrder.Accrual.RoundBank(4)))
	})

	t.Run("user not exists in database", func(t *testing.T) {
		login2 := user.Login(faker.Username())

		orderNumber := generateOrderNumber(t)

		o, err := order.New(orderNumber, login2)
		require.NoError(t, err)

		err = service.storages.Order().Create(ctx, *o)
		require.NoError(t, err)

		procOrder := *o
		procOrder.Status = order.StatusProcessed
		procOrder.Accrual = decimal.NewFromFloat(120)
		ch <- processor.AcrualOrderResult{
			Order: &procOrder,
			Err:   nil,
		}
		service.processOrder(ctx, *o)

		storageOrder, err := service.storages.Order().Get(ctx, o.Number)
		require.NoError(t, err)
		assert.Equal(t, o.Number, storageOrder.Number)
		assert.Equal(t, o.Status, storageOrder.Status)
	})

	t.Run("order status is invalid", func(t *testing.T) {
		orderNumber := generateOrderNumber(t)

		o, err := order.New(orderNumber, login)
		require.NoError(t, err)

		err = service.storages.Order().Create(ctx, *o)
		require.NoError(t, err)

		procOrder := *o
		procOrder.Status = order.StatusInvalid
		ch <- processor.AcrualOrderResult{
			Order: &procOrder,
			Err:   nil,
		}
		service.processOrder(ctx, *o)

		storageOrder, err := service.storages.Order().Get(ctx, o.Number)
		require.NoError(t, err)
		assert.Equal(t, o.Number, storageOrder.Number)
		assert.Equal(t, procOrder.Status, storageOrder.Status)
	})

	t.Run("order status is processing", func(t *testing.T) {
		orderNumber := generateOrderNumber(t)

		o, err := order.New(orderNumber, login)
		require.NoError(t, err)

		err = service.storages.Order().Create(ctx, *o)
		require.NoError(t, err)

		procOrder := *o
		procOrder.Status = order.StatusProcessing
		ch <- processor.AcrualOrderResult{
			Order: &procOrder,
			Err:   nil,
		}
		service.processOrder(ctx, *o)

		storageOrder, err := service.storages.Order().Get(ctx, o.Number)
		require.NoError(t, err)
		assert.Equal(t, o.Number, storageOrder.Number)
		assert.Equal(t, procOrder.Status, storageOrder.Status)
	})
}
