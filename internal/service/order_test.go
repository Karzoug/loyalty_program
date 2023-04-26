package service

import (
	"context"
	"testing"
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/pioz/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	service := newMockServiceWithEmptyProcessor(ctx, t)

	// first user
	login := user.Login(faker.Username())
	password := faker.StringWithSize(15)
	_, err := service.RegisterUser(ctx, login, password)
	require.NoError(t, err)

	// second user
	login2 := user.Login(faker.Username())
	password2 := faker.StringWithSize(15)
	_, err = service.RegisterUser(ctx, login2, password2)
	require.NoError(t, err)

	orderNumber := generateOrderNumber(t)

	t.Run("positive", func(t *testing.T) {
		o, err := order.New(orderNumber, login)
		require.NoError(t, err)

		o, exists, err := service.CreateOrder(ctx, login, orderNumber)
		require.NoError(t, err)
		assert.Equal(t, false, exists)
		assert.Equal(t, orderNumber, o.Number)
		assert.Equal(t, order.StatusNew, o.Status)

		so, err := service.storages.Order().Get(ctx, orderNumber)
		require.NoError(t, err)
		assert.Equal(t, o.Number, so.Number)
		assert.True(t, o.Accrual.RoundBank(4).Equal(so.Accrual.RoundBank(4)))
	})

	t.Run("negative: invalid order number", func(t *testing.T) {
		invalidOrderNumber := generateInvalidOrderNumber(t)
		_, _, err = service.CreateOrder(ctx, login, invalidOrderNumber)
		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})

	t.Run("negative: order exists, the same user", func(t *testing.T) {
		_, exists, err := service.CreateOrder(ctx, login, orderNumber)
		assert.NoError(t, err)
		assert.Equal(t, true, exists)
	})

	t.Run("negative: order exists, another user", func(t *testing.T) {
		_, exists, err := service.CreateOrder(ctx, login2, orderNumber)
		assert.ErrorIs(t, err, ErrAnotherUserOrderNumber)
		assert.Equal(t, true, exists)
	})
}

func TestService_ListUserOrders(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	service := newMockServiceWithEmptyProcessor(ctx, t)

	// first user
	login := user.Login(faker.Username())
	password := faker.StringWithSize(15)
	_, err := service.RegisterUser(ctx, login, password)
	require.NoError(t, err)

	// second user
	login2 := user.Login(faker.Username())
	password2 := faker.StringWithSize(15)
	_, err = service.RegisterUser(ctx, login2, password2)
	require.NoError(t, err)

	// third user
	login3 := user.Login(faker.Username())
	password3 := faker.StringWithSize(15)
	_, err = service.RegisterUser(ctx, login3, password3)
	require.NoError(t, err)

	// first user: first order
	orderNumber := generateOrderNumber(t)
	_, err = order.New(orderNumber, login)
	require.NoError(t, err)
	_, _, err = service.CreateOrder(ctx, login, orderNumber)
	require.NoError(t, err)

	// first user: second order
	orderNumber = generateOrderNumber(t)
	_, err = order.New(orderNumber, login)
	require.NoError(t, err)
	_, _, err = service.CreateOrder(ctx, login, orderNumber)
	require.NoError(t, err)

	// second user: first order
	orderNumber = generateOrderNumber(t)
	_, err = order.New(orderNumber, login2)
	require.NoError(t, err)
	_, _, err = service.CreateOrder(ctx, login2, orderNumber)
	require.NoError(t, err)

	orders, err := service.ListUserOrders(ctx, login)
	assert.Equal(t, 2, len(orders))

	orders, err = service.ListUserOrders(ctx, login2)
	require.Equal(t, 1, len(orders))
	assert.Equal(t, orderNumber, orders[0].Number)

	orders, err = service.ListUserOrders(ctx, login3)
	require.Equal(t, 0, len(orders))
}
