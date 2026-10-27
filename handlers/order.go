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

// PlaceOrder handles the customer's checkout request.
func (h *OrderHandler) PlaceOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Step 1: Identify the authenticated customer.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Step 2: Read the customer's delivery information.
	order := &models.Order{}

	if err := json.NewDecoder(r.Body).Decode(order); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid order information")
		return
	}

	// Validate required delivery information.
	if order.DeliveryName == "" ||
		order.DeliveryPhone == "" ||
		order.DeliveryAddress == "" ||
		order.DeliveryCity == "" ||
		order.DeliveryState == "" {

		utils.Error(w, http.StatusBadRequest, "delivery information is required")
		return
	}

	// Payment is simulated and assumed successful for the MVP.
	// No real payment provider is used.

	// Step 3: Send the order to the service.
	placedOrder, err := h.OrderService.PlaceOrder(
		r.Context(),
		claims.UserID,
		order,
	)

	// Step 4: Handle checkout errors.
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

	// Step 5: Return the successful order.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(placedOrder)

}
