package handlers

import (
	"alpha-dev/service"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
)

//holds service as interface
type OrderHandler struct {
	OrderService service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: service}

}

// DTO (data transfer object) with a struct tag json:"status" mapping the JSON key to the field
type UpdateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status"`
}

// UpdateOrderStatus handles PATCH /orders/{id}/status.
//reads the logged-in user’s identity, which the Authenticate middleware stored from the JWT. We trust the token, never an ID sent in the body.
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	//converts text to an unsigned number

	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	//caps the body at 1 KB so nobody can upload gigabytes to exhaust memory.

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var req UpdateOrderStatusRequest
	//rejects extra keys
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid status JSON")
		return
	}
	//must return io.EOF. If the client sent two JSON objects back to back, we reject it.
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid status JSON")
		return
	}

	//The service call passes claims.UserID and claims.Role from the token and uint(id) from the URL. The uint(id) is a type conversion, because ParseUint returns uint64.

	order, err := h.OrderService.UpdateOrderStatus(r.Context(), claims.UserID, claims.Role, uint(id), req.Status)
	//	//Calls the service. It returns two things: the updated order and an error. If everything worked, err is nil
	if err != nil {
		//asks “is this error (or anything it wraps) of this custom type?”. We need As here, not Is, because we are matching a type, not a fixed value
		var transitionErr *service.OrderTransitionError
		switch {
		case errors.As(err, &transitionErr):
			utils.Error(w, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrInvalidOrderStatus):
			//Sends a 422 response. ValidationError (from your utils/response.go) takes a general message plus a map saying which field was wrong.
			utils.ValidationError(w, "check the status field", map[string]string{
				"status": err.Error(),
			})
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

	utils.Success(w, http.StatusOK, "order status updated", map[string]any{
		"id":           order.ID,
		"order_number": order.OrderNumber,
		"status":       order.Status,
		"updated_at":   order.UpdatedAt,
	})
}
