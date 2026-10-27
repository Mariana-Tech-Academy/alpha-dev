package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"alpha-dev/auth"
	"alpha-dev/service"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
)

type MenuHandler struct {
	MenuService service.MenuService
}

func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: menuService}
}

// Keep the existing menu creation feature.
func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Read one menu item from the request.
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req service.CreateMenuItemRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid menu item JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid menu item JSON")
		return
	}

	// Create the item and handle any errors.
	item, err := h.MenuService.CreateMenuItem(r.Context(), claims.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNameRequired),
			errors.Is(err, service.ErrInvalidPrice),
			errors.Is(err, service.ErrRestaurantIDRequired),
			errors.Is(err, service.ErrCategoryIDRequired):
			utils.Error(w, http.StatusBadRequest, err.Error())

		case errors.Is(err, service.ErrRestaurantNotFound),
			errors.Is(err, service.ErrCategoryNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())

		case errors.Is(err, service.ErrNotYourRestaurant),
			errors.Is(err, service.ErrRestaurantClosed):
			utils.Error(w, http.StatusForbidden, err.Error())

		default:
			log.Printf("create menu item: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to create menu item")
		}
		return
	}

	utils.Success(w, http.StatusCreated, "menu item created", item)
}

// Handle a customer's request to view a menu.
func (h *MenuHandler) GetRestaurantMenu(w http.ResponseWriter, r *http.Request) {
	// Read and check the restaurant ID.
	restaurantID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, strconv.IntSize)
	if err != nil || restaurantID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	// Use zero for all categories, or read the selected category.
	var categoryID uint
	values, supplied := r.URL.Query()["category_id"]
	if supplied {
		if len(values) != 1 {
			utils.Error(w, http.StatusBadRequest, "provide one category ID")
			return
		}

		id, err := strconv.ParseUint(values[0], 10, strconv.IntSize)
		if err != nil || id == 0 {
			utils.Error(w, http.StatusBadRequest, "invalid category ID")
			return
		}
		categoryID = uint(id)
	}

	// Ask the service for the menu.
	menu, err := h.MenuService.GetRestaurantMenu(
		r.Context(), uint(restaurantID), categoryID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantIDRequired):
			utils.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrMenuUnavailable):
			utils.Error(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("get restaurant menu: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to load menu")
		}
		return
	}

	// Return the menu to the customer.
	utils.Success(w, http.StatusOK, "restaurant menu", menu)
}
