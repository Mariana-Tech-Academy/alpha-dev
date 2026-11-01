package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
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

type UpdateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status"`
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	var req UpdateOrderStatusRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid status JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid status JSON")
		return
	}

	order, err := h.OrderService.UpdateOrderStatus(r.Context(), claims.UserID, claims.Role, uint(id), req.Status)
	if err != nil {
		var transitionErr *service.OrderTransitionError
		switch {
		case errors.As(err, &transitionErr):
			utils.Error(w, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrInvalidOrderStatus):
			utils.ValidationError(w, "check the status field", map[string]string{
				"status": err.Error(),
			})
		case errors.Is(err, service.ErrInvalidUser):
			utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		case errors.Is(err, service.ErrOrderNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrOrderForbidden):
			utils.Error(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrOrderChanged):
			utils.Error(w, http.StatusConflict, err.Error())
		default:
			log.Printf("update order %d status: %v", id, err)
			utils.Error(w, http.StatusInternalServerError, "unable to update order status")
		}
		return
	}

	utils.Success(w, http.StatusOK, "order status updated", order)
}
