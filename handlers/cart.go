package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"alpha-dev/auth"
	"alpha-dev/service"
	"alpha-dev/utils"
)

type CartHandler struct {
	CartService service.CartService
}

func NewCartHandler(s service.CartService) *CartHandler {
	return &CartHandler{CartService: s}
}

type AddCartItemRequest struct {
	MenuItemID uint `json:"menu_item_id"`
	Quantity   int  `json:"quantity"`
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req AddCartItemRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid cart JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid cart JSON")
		return
	}

	if req.MenuItemID == 0 {
		utils.ValidationError(w, "check the cart details", map[string]string{
			"menu_item_id": "menu item ID must be greater than 0",
		})
		return
	}
	if req.Quantity < 1 {
		utils.ValidationError(w, "check the cart details", map[string]string{
			"quantity": "quantity must be at least 1",
		})
		return
	}

	item, err := h.CartService.AddItem(r.Context(), claims.UserID, req.MenuItemID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUser):
			utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		case errors.Is(err, service.ErrInvalidQuantity),
			errors.Is(err, service.ErrCartQuantityLimit):
			utils.ValidationError(w, "check the cart details", map[string]string{
				"quantity": err.Error(),
			})
		case errors.Is(err, service.ErrInvalidMenuItem):
			utils.ValidationError(w, "check the cart details", map[string]string{
				"menu_item_id": err.Error(),
			})
		case errors.Is(err, service.ErrCartMenuItemNotFound):
			utils.Error(w, http.StatusNotFound, "menu item not found")
		case errors.Is(err, service.ErrItemUnavailable):
			utils.Error(w, http.StatusBadRequest, "menu item is unavailable")
		case errors.Is(err, service.ErrCartDifferentRestaurant):
			utils.Error(w, http.StatusConflict, err.Error())
		default:
			log.Printf("add cart item failed: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to add item to cart")
		}
		return
	}

	utils.Success(w, http.StatusOK, "item added to cart successfully", map[string]any{
		"id":           item.ID,
		"cart_id":      item.CartID,
		"menu_item_id": item.MenuItemID,
		"quantity":     item.Quantity,
		"price":        item.Price,
	})
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	cartView, err := h.CartService.GetCartView(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUser) {
			utils.Error(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		log.Printf("get cart: %v", err)
		utils.Error(w, http.StatusInternalServerError, "unable to retrieve cart")
		return
	}

	utils.Success(w, http.StatusOK, "cart retrieved successfully", cartView)
}
