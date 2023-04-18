package helper

import (
	"context"
	"errors"
	"net/http"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/service"
	"github.com/go-chi/jwtauth"
	"go.uber.org/zap"
)

var ErrInvalidLoginType = errors.New("invalid login type: not string")

func GetLoginFromJWTInContext(ctx context.Context, logger *zap.Logger) (*user.Login, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	sub, exists := claims["sub"]
	if !exists {
		return nil, &HandlerError{
			Message: service.ErrInvalidAuthData.Error(),
			Code:    http.StatusUnauthorized,
		}
	}
	loginString, ok := sub.(string)
	if !ok {
		return nil, ErrInvalidLoginType
	}
	useLogin := user.Login(loginString)

	return &useLogin, nil
}
