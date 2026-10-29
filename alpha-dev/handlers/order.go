
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type OrderHandler struct {
	OrderService service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: service}
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Read the order ID from the URL.
	idString := mux.Vars(r)["id"]
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil || id == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Retrieve the authenticated user's claims.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Retrieve the order and check that it belongs to this customer.
	order, err := h.OrderService.GetOrderForCustomer(
		r.Context(),
		uint(id),
		claims.UserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderForbidden):
			utils.Error(w, http.StatusForbidden, "you cannot access this order")
		case errors.Is(err, gorm.ErrRecordNotFound):
			utils.Error(w, http.StatusNotFound, "order not found")
		default:
			utils.Error(w, http.StatusInternalServerError, "failed to fetch order")
		}
		return
	}

	// Return only the information needed by the customer.
	response := struct {
		ID                   uint              `json:"id"`
		OrderNumber          string            `json:"order_number"`
		Items                []models.OrderItem `json:"items"`
		Subtotal             int64             `json:"subtotal"`
		DeliveryFee          int64             `json:"delivery_fee"`
		Total                int64             `json:"total"`
		DeliveryName         string            `json:"delivery_name"`
		DeliveryPhone        string            `json:"delivery_phone"`
		DeliveryAddress      string            `json:"delivery_address"`
		DeliveryCity         string            `json:"delivery_city"`
		DeliveryState        string            `json:"delivery_state"`
		DeliveryInstructions string            `json:"delivery_instructions"`
		Status               models.OrderStatus `json:"status"`
	}{
		ID:                   order.ID,
		OrderNumber:          order.OrderNumber,
		Items:                order.Items,
		Subtotal:             order.Subtotal,
		DeliveryFee:          order.DeliveryFee,
		Total:                order.Total,
		DeliveryName:         order.DeliveryName,
		DeliveryPhone:        order.DeliveryPhone,
		DeliveryAddress:      order.DeliveryAddress,
		DeliveryCity:         order.DeliveryCity,
		DeliveryState:        order.DeliveryState,
		DeliveryInstructions: order.DeliveryInstructions,
		Status:               order.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
