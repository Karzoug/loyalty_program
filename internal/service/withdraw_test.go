package service

import (
	"context"
	"testing"
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/pioz/faker"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateWithdraw(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	service := newMockServiceWithEmptyProcessor(ctx, t)

	t.Run("positive", func(t *testing.T) {
		login := user.Login(faker.Username())
		password := faker.StringWithSize(15)
		_, err := service.RegisterUser(ctx, login, password)
		require.NoError(t, err)

		delta := decimal.NewFromFloat(750)
		balance, err := service.storages.User().UpdateBalance(ctx, login, delta)
		require.NoError(t, err)
		require.True(t, balance.RoundBank(4).Equal(delta.RoundBank(4)))

		orderNumber := generateOrderNumber(t)

		withdrawSum := decimal.NewFromFloat(150)
		w, err := service.CreateWithdraw(ctx, login, orderNumber, withdrawSum)
		require.NoError(t, err)
		assert.Equal(t, login, w.UserLogin)
		assert.True(t, w.Sum.RoundBank(4).Equal(withdrawSum.RoundBank(4)))

		u, err := service.storages.User().Get(ctx, login)
		require.NoError(t, err)
		assert.Equal(t, login, u.Login)
		assert.True(t, u.Balance.RoundBank(4).Equal(balance.Sub(withdrawSum.RoundBank(4))))
	})

	t.Run("negative: invalid order number", func(t *testing.T) {
		login := user.Login(faker.Username())
		password := faker.StringWithSize(15)
		_, err := service.RegisterUser(ctx, login, password)
		require.NoError(t, err)

		invalidOrderNumber := generateInvalidOrderNumber(t)
		sum := decimal.NewFromFloat(50)
		_, err = service.CreateWithdraw(ctx, login, invalidOrderNumber, sum)
		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})

	t.Run("negative: re-attempt withdraw", func(t *testing.T) {
		login := user.Login(faker.Username())
		password := faker.StringWithSize(15)
		_, err := service.RegisterUser(ctx, login, password)
		require.NoError(t, err)

		delta := decimal.NewFromFloat(750)
		balance, err := service.storages.User().UpdateBalance(ctx, login, delta)
		require.NoError(t, err)
		require.True(t, balance.RoundBank(4).Equal(delta.RoundBank(4)))

		number := generateOrderNumber(t)
		sum := decimal.NewFromFloat(50)
		_, err = service.CreateWithdraw(ctx, login, number, sum)
		require.NoError(t, err)
		_, err = service.CreateWithdraw(ctx, login, number, sum)
		assert.ErrorIs(t, err, ErrReAttemptWithdraw)
	})

	t.Run("negative: insufficient balance", func(t *testing.T) {
		login := user.Login(faker.Username())
		password := faker.StringWithSize(15)
		_, err := service.RegisterUser(ctx, login, password)
		require.NoError(t, err)
		delta := decimal.NewFromFloat(20)
		balance, err := service.storages.User().UpdateBalance(ctx, login, delta)
		require.NoError(t, err)
		require.True(t, balance.RoundBank(4).Equal(delta.RoundBank(4)))

		orderNumber := generateOrderNumber(t)

		withdrawSum := decimal.NewFromFloat(150)
		_, err = service.CreateWithdraw(ctx, login, orderNumber, withdrawSum)
		assert.ErrorIs(t, err, ErrInsufficientBalance)
	})

}

func TestService_SumUserWithdrawals(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	service := newMockServiceWithEmptyProcessor(ctx, t)

	login := user.Login(faker.Username())
	password := faker.StringWithSize(15)
	_, err := service.RegisterUser(ctx, login, password)
	require.NoError(t, err)

	delta := decimal.NewFromFloat(faker.Float64InRange(800, 1000))
	balance, err := service.storages.User().UpdateBalance(ctx, login, delta)
	require.NoError(t, err)
	require.True(t, balance.RoundBank(4).Equal(delta.RoundBank(4)))

	// first withdraw
	orderNumber := generateOrderNumber(t)
	withdrawSum1 := decimal.NewFromFloat(faker.Float64InRange(100, 400))
	_, err = service.CreateWithdraw(ctx, login, orderNumber, withdrawSum1)
	require.NoError(t, err)

	// second withdraw
	orderNumber = generateOrderNumber(t)
	withdrawSum2 := decimal.NewFromFloat(faker.Float64InRange(1, 50))
	_, err = service.CreateWithdraw(ctx, login, orderNumber, withdrawSum2)
	require.NoError(t, err)

	// third withdraw
	orderNumber = generateOrderNumber(t)
	withdrawSum3 := decimal.NewFromFloat(faker.Float64InRange(50, 200))
	_, err = service.CreateWithdraw(ctx, login, orderNumber, withdrawSum3)
	require.NoError(t, err)

	// result sum
	withdrawSum := decimal.Sum(withdrawSum1, withdrawSum2, withdrawSum3)
	sum := decimal.Sum(*balance, withdrawSum.Neg())

	// check user balance
	u, err := service.storages.User().Get(ctx, login)
	require.NoError(t, err)
	assert.Equal(t, login, u.Login)
	assert.True(t, u.Balance.RoundBank(4).Equal(sum.RoundBank(4)))

	// check sum of withdrawals
	ws, err := service.SumUserWithdrawals(ctx, login)
	require.NoError(t, err)
	assert.True(t, ws.RoundBank(4).Equal(withdrawSum.RoundBank(4)))
}
