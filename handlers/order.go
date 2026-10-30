package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"alpha-dev/auth"
	"alpha-dev/service"
	"alpha-dev/utils"
)

type OrderHandler struct {
	OrderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: orderService}
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	var req service.PlaceOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid order JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid order JSON")
		return
	}

	placed, err := h.OrderService.PlaceOrder(r.Context(), claims.UserID, req)
	if err != nil {
		var validationErr *service.OrderValidationError
		switch {
		case errors.As(err, &validationErr):
			utils.ValidationError(w, "check the delivery details", validationErr.Fields)
		case errors.Is(err, service.ErrInvalidUser):
			utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		case errors.Is(err, service.ErrEmptyCart):
			utils.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrCheckoutRestaurantInactive),
			errors.Is(err, service.ErrCheckoutItemUnavailable):
			utils.Error(w, http.StatusConflict, err.Error())
		default:
			log.Printf("place order: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to place order")
		}
		return
	}

	utils.Success(w, http.StatusCreated, "order placed", placed)
}

func (h *OrderHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			page = parsed
		}
	}
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	history, err := h.OrderService.GetOrderHistory(r.Context(), claims.UserID, page, limit)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUser) {
			utils.Error(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		log.Printf("order history: %v", err)
		utils.Error(w, http.StatusInternalServerError, "unable to load order history")
		return
	}

	utils.Success(w, http.StatusOK, "orders retrieved", history)
}
