package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Karzoug/loyalty_program/internal/delivery/rest/helper"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/service"
	"github.com/Karzoug/loyalty_program/pkg/e"
	"github.com/go-chi/jwtauth"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	tokenLifetime = 24 * time.Hour
)

var (
	ErrEmptyLogin    = errors.New("login must be non empty")
	ErrEmptyPassword = errors.New("password must be non empty")
)

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (a authRequest) validate() error {
	if a.Login == "" {
		return ErrEmptyLogin
	}
	if a.Password == "" {
		return ErrEmptyPassword
	}
	return nil
}

func (s *server) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), handlerTimeout)
	defer cancel()

	var (
		authReq authRequest
		hErr    *helper.HandlerError
	)
	err := helper.DecodeJSON(r, &authReq)
	if err != nil {
		if errors.As(err, &hErr) {
			helper.WriteJSONError(w, hErr.Message, hErr.Code, s.logger)
		} else {
			s.logger.Error("Register user handler: decode request from JSON error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	if err := authReq.validate(); err != nil {
		helper.WriteJSONError(w, err.Error(), http.StatusBadRequest, s.logger)
		return
	}

	u, err := s.service.RegisterUser(ctx, user.Login(authReq.Login), authReq.Password)
	if err != nil {
		switch err {
		case service.ErrLoginAlreadyExists:
			helper.WriteJSONError(w, err.Error(), http.StatusConflict, s.logger)
		case service.ErrInvalidPasswordFormat, service.ErrInvalidLoginFormat:
			helper.WriteJSONError(w, err.Error(), http.StatusBadRequest, s.logger)
		default:
			s.logger.Error("Register user handler: user register service error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	if err := writeAuthToken(w, u.Login, s.cfg.SecretKey()); err != nil {
		s.logger.Error("Register user handler: write token to response error", zap.Error(err))
		helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		return
	}
}

func (s *server) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), handlerTimeout)
	defer cancel()

	var (
		authReq authRequest
		hErr    *helper.HandlerError
	)
	err := helper.DecodeJSON(r, &authReq)
	if err != nil {
		if errors.As(err, &hErr) {
			helper.WriteJSONError(w, hErr.Message, hErr.Code, s.logger)
		} else {
			s.logger.Error("Login user handler: decode auth request from JSON error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	if err := authReq.validate(); err != nil {
		helper.WriteJSONError(w, err.Error(), http.StatusBadRequest, s.logger)
		return
	}

	u, err := s.service.LoginUser(ctx, user.Login(authReq.Login), authReq.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidAuthData:
			helper.WriteJSONError(w, err.Error(), http.StatusUnauthorized, s.logger)
		default:
			s.logger.Error("Login user handler: user login service error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	if err := writeAuthToken(w, u.Login, s.cfg.SecretKey()); err != nil {
		s.logger.Error("Login user handler: write token to response error", zap.Error(err))
		helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		return
	}
}

func writeAuthToken(w http.ResponseWriter, login user.Login, secretKey string) error {
	tokenAuth := jwtauth.New("HS256", []byte(secretKey), nil)

	claims := make(map[string]interface{})
	claims["sub"] = string(login)
	jwtauth.SetExpiryIn(claims, tokenLifetime)

	_, tokenString, err := tokenAuth.Encode(claims)
	if err != nil {
		return e.Wrap("encode JWT", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Authorization", "BEARER "+tokenString)
	w.WriteHeader(http.StatusOK)
	return nil
}

type balanceResponse struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn int             `json:"withdrawn"`
}

func (s *server) getUserBalanceHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), handlerTimeout)
	defer cancel()

	var hErr *helper.HandlerError
	login, err := helper.GetLoginFromJWTInContext(ctx, s.logger)
	if err != nil {
		if errors.As(err, &hErr) {
			helper.WriteJSONError(w, hErr.Message, hErr.Code, s.logger)
		} else {
			s.logger.Error("Create order handler: get login from context error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	balance, err := s.service.GetUserBalance(ctx, *login)
	if err != nil {
		switch err {
		case service.ErrInvalidAuthData:
			helper.WriteJSONError(w, err.Error(), http.StatusUnauthorized, s.logger)
		default:
			s.logger.Error("Get user balance handler: user balance service error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	count, err := s.service.CountUserWithdrawals(ctx, *login)
	if err != nil {
		s.logger.Error("Get user balance handler: withdrawals count service error", zap.Error(err))
		helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		return
	}

	balanceResp := balanceResponse{
		Balance:   *balance,
		Withdrawn: count,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(balanceResp); err != nil {
		s.logger.Error("Get user balance handler: encode json response error", zap.Error(err))
		helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		return
	}
}
