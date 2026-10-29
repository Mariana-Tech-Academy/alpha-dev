package handlers

//handlers/checkout.go is the code that runs when someone calls GET /api/checkout/summary
//It receives the HTTP request, calls the service, and sends back the HTTP response. It does no calculations and no database work.

import (
	"errors"
	"log"
	"net/http"

	"alpha-dev/auth"
	"alpha-dev/service"
	"alpha-dev/utils"
)

//The handler needs the service to do its work, so it stores one. NewCheckoutHandler creates a handler and gives it that service
type CheckoutHandler struct {
	CheckoutService service.CheckoutService
}

func NewCheckoutHandler(s service.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{CheckoutService: s}
}
// Summary handles GET /checkout/summary (CHK-01).
func (h *CheckoutHandler) Summary(w http.ResponseWriter, r *http.Request) {
	//The auth middleware has already read the customer’s login token and stored the user’s details on the request. This line reads them back out. If they aren’t there, it responds with 401 Unauthorized and stops.
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
			utils.Error(w, http.StatusBadRequest, "your cart is empty, add items before checking out")
		case errors.Is(err, service.ErrItemUnavailable):
			utils.Error(w, http.StatusConflict, "the restaurant in your cart is unavailable")
		default:
			log.Printf("checkout summary failed: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to load checkout summary")
		}
		return
	}

	utils.Success(w, http.StatusOK, "checkout summary retrieved successfully", summary)
}