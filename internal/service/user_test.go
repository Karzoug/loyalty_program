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

func TestService_RegisterUser(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	service := newMockServiceWithEmptyProcessor(ctx, t)

	login := user.Login(faker.Username())
	password := faker.StringWithSize(15)

	t.Run("positive", func(t *testing.T) {
		u, err := service.RegisterUser(ctx, login, password)
		require.NoError(t, err)
		assert.Equal(t, login, u.Login)
		assert.True(t, decimal.Zero.Equal(u.Balance))
		assert.True(t, u.VerifyPassword(password))

		stUser, err := service.storages.User().Get(ctx, login)
		require.NoError(t, err)
		assert.Equal(t, login, stUser.Login)
		assert.True(t, decimal.Zero.Equal(stUser.Balance))
		assert.True(t, stUser.VerifyPassword(password))
	})

	t.Run("negative: user already exists", func(t *testing.T) {
		password2 := faker.StringWithSize(15)
		_, err := service.RegisterUser(ctx, login, password2)
		assert.ErrorIs(t, err, ErrLoginAlreadyExists)

		stUser, err := service.storages.User().Get(ctx, login)
		require.NoError(t, err)
		assert.False(t, stUser.VerifyPassword(password2))
		assert.True(t, stUser.VerifyPassword(password))
	})

	t.Run("negative: empty login/password", func(t *testing.T) {
		_, err := service.RegisterUser(ctx, "", password)
		assert.ErrorIs(t, err, ErrInvalidLoginFormat)

		login2 := user.Login(faker.Username())
		_, err = service.RegisterUser(ctx, login2, "")
		assert.ErrorIs(t, err, ErrInvalidPasswordFormat)
	})

}

func TestService_LoginUser(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	service := newMockServiceWithEmptyProcessor(ctx, t)

	login := user.Login(faker.Username())
	password := faker.StringWithSize(15)

	ru, err := service.RegisterUser(ctx, login, password)
	require.NoError(t, err)
	assert.Equal(t, login, ru.Login)
	assert.True(t, decimal.Zero.Equal(ru.Balance))
	assert.True(t, ru.VerifyPassword(password))

	t.Run("positive", func(t *testing.T) {
		lu, err := service.LoginUser(ctx, login, password)
		require.NoError(t, err)
		assert.Equal(t, login, lu.Login)
		assert.True(t, lu.VerifyPassword(password))
	})
	t.Run("negative: invalid password", func(t *testing.T) {
		password2 := faker.StringWithSize(15)
		_, err = service.LoginUser(ctx, login, password2)
		assert.ErrorIs(t, err, ErrInvalidAuthData)
	})
	t.Run("negative: user not exists", func(t *testing.T) {
		login2 := user.Login(faker.Username())
		_, err = service.LoginUser(ctx, login2, password)
		assert.ErrorIs(t, err, ErrInvalidAuthData)
	})
}

func TestService_GetUserBalance(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	service := newMockServiceWithEmptyProcessor(ctx, t)

	login := user.Login(faker.Username())
	password := faker.StringWithSize(15)

	_, err := service.RegisterUser(ctx, login, password)
	require.NoError(t, err)

	t.Run("positive", func(t *testing.T) {
		balance, err := service.GetUserBalance(ctx, login)
		require.NoError(t, err)
		assert.True(t, decimal.Zero.Equal(*balance))
	})

	t.Run("negative: user not exists", func(t *testing.T) {
		login2 := user.Login(faker.Username())
		_, err = service.GetUserBalance(ctx, login2)
		assert.ErrorIs(t, err, ErrInvalidAuthData)
	})
}
