package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"alpha-dev/auth"
	"alpha-dev/repository"
	"alpha-dev/service"
	"alpha-dev/utils"
)

type CartHandler struct {
	CartService service.CartService
}

func NewCartHandler(s service.CartService) *CartHandler {
	return &CartHandler{
		CartService: s,
	}
}

// AddCartItemRequest represents the JSON sent by the customer.
type AddCartItemRequest struct {
	MenuItemID uint `json:"menu_item_id"`
	Quantity   int  `json:"quantity"`
}

// AddItem handles POST /cart/items.
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {

	// STEP 1: Identify the authenticated customer.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// STEP 2: Limit the size of the request body.
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	// STEP 3: Decode the customer's JSON request.
	var req AddCartItemRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid cart JSON")
		return
	}

	// Reject additional JSON objects or trailing content.
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid cart JSON")
		return
	}

	// STEP 4: Validate the menu item ID.
	if req.MenuItemID == 0 {
		utils.ValidationError(
			w,
			"check the cart details",
			map[string]string{
				"menu_item_id": "menu item ID must be greater than 0",
			},
		)
		return
	}

	// STEP 5: Validate the quantity.
	if req.Quantity < 1 {
		utils.ValidationError(
			w,
			"check the cart details",
			map[string]string{
				"quantity": "quantity must be at least 1",
			},
		)
		return
	}

	// STEP 6: Call the cart service.
	item, err := h.CartService.AddItem(
		r.Context(),
		claims.UserID,
		req.MenuItemID,
		req.Quantity,
	)

	// STEP 7: Handle errors.
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUser):
			utils.Error(
				w,
				http.StatusUnauthorized,
				"unauthenticated",
			)

		case errors.Is(err, service.ErrInvalidQuantity):
			utils.ValidationError(
				w,
				"check the cart details",
				map[string]string{
					"quantity": err.Error(),
				},
			)

		case errors.Is(err, service.ErrInvalidMenuItem):
			utils.ValidationError(
				w,
				"check the cart details",
				map[string]string{
					"menu_item_id": err.Error(),
				},
			)

		case errors.Is(err, service.ErrCartMenuItemNotFound):
			utils.Error(
				w,
				http.StatusNotFound,
				"menu item not found",
			)

		case errors.Is(err, service.ErrItemUnavailable):
			utils.Error(
				w,
				http.StatusBadRequest,
				"menu item is unavailable",
			)

		case errors.Is(err, repository.ErrDifferentRestaurant):
			utils.Error(
				w,
				http.StatusConflict,
				"cart contains items from another restaurant",
			)

		default:
			log.Printf("add cart item failed: %v", err)

			utils.Error(
				w,
				http.StatusInternalServerError,
				"unable to add item to cart",
			)
		}
		return
	}

	// STEP 8: Return the updated cart item.
	utils.Success(
		w,
		http.StatusOK,
		"item added to cart successfully",
		map[string]any{
			"id":           item.ID,
			"cart_id":      item.CartID,
			"menu_item_id": item.MenuItemID,
			"quantity":     item.Quantity,
			"price":        item.MenuItem.Price,
		},
	)
}
