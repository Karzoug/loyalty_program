package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Karzoug/loyalty_program/internal/delivery/rest/helper"
	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/service"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func (s *server) createOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), handlerTimeout)
	defer cancel()

	var hErr *helper.HandlerError
	login, err := helper.GetLoginFromJWTInContext(ctx, s.logger)
	if err != nil {
		if errors.As(err, &hErr) {
			http.Error(w, hErr.Message, hErr.Code)
		} else {
			s.logger.Error("Create order handler: get login from context error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Create order handler: read request body error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(data) == 0 {
		http.Error(w, "order number empty", http.StatusBadRequest)
		return
	}
	number, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		http.Error(w, service.ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity)
		return
	}
	orderNumber := order.Number(number)

	_, isExistedOrder, err := s.service.CreateOrder(ctx, *login, orderNumber)
	if err != nil {
		switch err {
		case service.ErrInvalidOrderNumber:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case service.ErrAnotherUserOrderNumber:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			s.logger.Error("Create order handler: create order service error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if isExistedOrder {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type orderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (s *server) listUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), handlerTimeout)
	defer cancel()

	var hErr *helper.HandlerError
	login, err := helper.GetLoginFromJWTInContext(ctx, s.logger)
	if err != nil {
		if errors.As(err, &hErr) {
			helper.WriteJSONError(w, hErr.Message, hErr.Code, s.logger)
		} else {
			s.logger.Error("List user orders handler: get login from context error", zap.Error(err))
			helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		}
		return
	}

	orders, err := s.service.ListUserOrders(ctx, *login)
	if err != nil {
		s.logger.Error("List user orders handler: list user orders service error", zap.Error(err))
		helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ordersResp := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		ordersResp = append(ordersResp, orderResponse{
			Number:     strconv.FormatInt(int64(o.Number), 10),
			Status:     o.Status.String(),
			Accrual:    o.Accrual.InexactFloat64(),
			UploadedAt: o.UploadedAt,
		})
	}
	sort.Slice(ordersResp, func(i, j int) bool {
		return ordersResp[i].UploadedAt.Before(ordersResp[j].UploadedAt)
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ordersResp); err != nil {
		s.logger.Error("List user orders handler: encode json response error", zap.Error(err))
		helper.WriteJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, s.logger)
		return
	}
}
