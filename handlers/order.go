package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"
	"alpha-dev/utils"
)

type OrderHandler struct {
	OrderService service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: service}
}

// ORD-01: returns readable order_number ORD-YYYYMMDD-NNN
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// get user from JWT
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// read delivery info
	order := &models.Order{}
	if err := json.NewDecoder(r.Body).Decode(order); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid order information")
		return
	}

	// validate delivery fields
	if order.DeliveryName == "" || order.DeliveryPhone == "" || order.DeliveryAddress == "" || order.DeliveryCity == "" || order.DeliveryState == "" {
		utils.Error(w, http.StatusBadRequest, "delivery information is required")
		return
	}

	// place order - ORD-01 number generated in repo
	placedOrder, err := h.OrderService.PlaceOrder(r.Context(), claims.UserID, order)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmptyCart):
			utils.Error(w, http.StatusBadRequest, "your cart is empty")
		case errors.Is(err, repository.ErrInactiveRestaurant):
			utils.Error(w, http.StatusBadRequest, "restaurant is not active")
		case errors.Is(err, repository.ErrUnavailableItem):
			utils.Error(w, http.StatusBadRequest, "one or more menu items are unavailable")
		default:
			utils.Error(w, http.StatusInternalServerError, "failed to place order")
		}
		return
	}

	// return 201 with order_number
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(placedOrder)
}
