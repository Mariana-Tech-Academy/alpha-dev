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
