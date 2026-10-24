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

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

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

// UpdateMenuItemAvailability allows a restaurant manager
// to mark a menu item as available or unavailable.
func (h *MenuHandler) UpdateMenuItemAvailability(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Get the logged-in user's details from the JWT.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Get the menu item ID from the URL.
	vars := mux.Vars(r)
	itemID, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil || itemID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	// Define the JSON request body.
	var req struct {
		Available *bool `json:"available"`
	}

	// Limit the request body size.
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	// Decode the JSON request.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	// Make sure no additional JSON is supplied.
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	// Ensure the available field was provided.
	if req.Available == nil {
		utils.Error(w, http.StatusBadRequest, "available field is required")
		return
	}

	// Call the service to update availability.
	err = h.MenuService.UpdateMenuItemAvailability(
		r.Context(),
		claims.UserID,
		uint(itemID),
		*req.Available,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrMenuItemNotFound),
			errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())

		case errors.Is(err, service.ErrNotYourRestaurant):
			utils.Error(w, http.StatusForbidden, err.Error())

		default:
			log.Printf("update menu availability: %v", err)
			utils.Error(w, http.StatusInternalServerError, "failed to update availability")
		}
		return
	}

	// Return a success response.
	w.WriteHeader(http.StatusNoContent)
}
