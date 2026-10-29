package handlers

import (
	"errors"
	"log"
	"net/http"

	"alpha-dev/auth"
	"alpha-dev/service"
	"alpha-dev/utils"
)

type CheckoutHandler struct {
	CheckoutService service.CheckoutService
}

func NewCheckoutHandler(checkoutService service.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{CheckoutService: checkoutService}
}

func (h *CheckoutHandler) Summary(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	summary, err := h.CheckoutService.Summary(r.Context(), claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUser):
			utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		case errors.Is(err, service.ErrEmptyCart):
			utils.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrCheckoutRestaurantInactive),
			errors.Is(err, service.ErrCheckoutItemUnavailable):
			utils.Error(w, http.StatusConflict, err.Error())
		default:
			log.Printf("checkout summary: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to load checkout summary")
		}
		return
	}

	utils.Success(w, http.StatusOK, "checkout summary retrieved successfully", summary)
}
