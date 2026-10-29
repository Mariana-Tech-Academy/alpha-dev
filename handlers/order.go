package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"
	"alpha-dev/utils"
)

type OrderHandler struct {
	OrderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: orderService}
}

// Read the customer's identity and requested page.
func (h *OrderHandler) GetOrderHistory(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Use page 1 and 10 orders unless the customer selects otherwise.
	page, err := orderHistoryParam(r, "page", 1)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid page")
		return
	}

	limit, err := orderHistoryParam(r, "limit", 10)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid limit")
		return
	}

	// Ask the service for this customer's history.
	history, err := h.OrderService.GetOrderHistory(
		r.Context(), claims.UserID, page, limit,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderHistoryUnauthenticated):
			utils.Error(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, service.ErrOrderHistoryPagination):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("get order history: %v", err)
			utils.Error(
				w,
				http.StatusInternalServerError,
				"unable to load order history",
			)
		}
		return
	}

	utils.Success(w, http.StatusOK, "order history", history)
}

// Accept one positive number for each pagination parameter.
func orderHistoryParam(
	r *http.Request,
	name string,
	fallback int,
) (int, error) {
	values, supplied := r.URL.Query()[name]
	if !supplied {
		return fallback, nil
	}
	if len(values) != 1 {
		return 0, errors.New("provide one value")
	}

	value, err := strconv.Atoi(values[0])
	if err != nil || value < 1 {
		return 0, errors.New("provide a positive number")
	}

	return value, nil
}

// Keep the existing checkout request handling.
func (h *OrderHandler) PlaceOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Read and check the delivery information.
	order := &models.Order{}

	if err := json.NewDecoder(r.Body).Decode(order); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid order information")
		return
	}

	if order.DeliveryName == "" ||
		order.DeliveryPhone == "" ||
		order.DeliveryAddress == "" ||
		order.DeliveryCity == "" ||
		order.DeliveryState == "" {
		utils.Error(w, http.StatusBadRequest, "delivery information is required")
		return
	}

	// Payment remains simulated by the existing checkout feature.
	placedOrder, err := h.OrderService.PlaceOrder(
		r.Context(), claims.UserID, order,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmptyCart):
			utils.Error(w, http.StatusBadRequest, "your cart is empty")
		case errors.Is(err, repository.ErrInactiveRestaurant):
			utils.Error(w, http.StatusBadRequest, "restaurant is not active")
		case errors.Is(err, repository.ErrUnavailableItem):
			utils.Error(
				w,
				http.StatusBadRequest,
				"one or more menu items are unavailable",
			)
		default:
			utils.Error(w, http.StatusInternalServerError, "failed to place order")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(placedOrder)
}
