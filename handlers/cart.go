package handlers

import (
	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
	"alpha-dev/utils"
	"encoding/json"
	"errors"
	"net/http"
)

type CartHandler struct {
	CartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{CartService: cartService}
}

type addItemRequest struct {
	MenuItemID uint `json:"menu_item_id"`
	Quantity   int  `json:"quantity"`
}

// ACTION: CART-02 - one restaurant per cart, validate add item
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if claims.Role != models.RoleCustomer {
		utils.Error(w, http.StatusForbidden, "you do not have permission to perform this action")
		return
	}

	var req addItemRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.MenuItemID == 0 {
		utils.ValidationError(w, "check your input", map[string]string{
			"menu_item_id": "menu item id is required",
		})
		return
	}
	if req.Quantity < 1 {
		utils.ValidationError(w, "check your input", map[string]string{
			"quantity": "quantity must be at least 1",
		})
		return
	}

	item, err := h.CartService.AddItem(r.Context(), claims.UserID, req.MenuItemID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCartInvalidUser):
			utils.Error(w, http.StatusUnauthorized, "unauthorized")
		case errors.Is(err, service.ErrCartInvalidMenuItem),
			errors.Is(err, service.ErrCartInvalidQuantity),
			errors.Is(err, service.ErrCartQuantityLimit):
			utils.ValidationError(w, "check your input", map[string]string{
				"quantity": err.Error(),
			})
		case errors.Is(err, service.ErrCartDifferentRestaurant):
			utils.ValidationError(w, "check your input", map[string]string{
				"restaurant_id": err.Error(),
			})
		case errors.Is(err, service.ErrCartMenuItemNotFound):
			utils.Error(w, http.StatusNotFound, "menu item not found")
		case errors.Is(err, service.ErrCartItemUnavailable):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.Success(w, http.StatusOK, "item added to cart", item)
}
