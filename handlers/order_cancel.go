package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"alpha-dev/auth"
	"alpha-dev/service"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
)

// CancelOrder handles POST /orders/{id}/cancel.
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
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

	cancelled, err := h.OrderService.CancelOrder(r.Context(), claims.UserID, claims.Role, uint(id))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUser):
			utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		case errors.Is(err, service.ErrOrderNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotYourOrder):
			utils.Error(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrOrderAlreadyCancelled),
			errors.Is(err, service.ErrOrderNotCancellable):
			utils.Error(w, http.StatusConflict, err.Error())
		default:
			log.Printf("cancel order %d: %v", id, err)
			utils.Error(w, http.StatusInternalServerError, "unable to cancel order")
		}
		return
	}

	utils.Success(w, http.StatusOK, "order cancelled", cancelled)
}
