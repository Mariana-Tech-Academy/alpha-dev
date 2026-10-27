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

func (h *MenuHandler) UpdateMenuItemAvailability(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	itemID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || itemID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	var req struct {
		Available *bool `json:"available"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	if req.Available == nil {
		utils.Error(w, http.StatusBadRequest, "available field is required")
		return
	}

	item, err := h.MenuService.UpdateMenuItemAvailability(r.Context(), claims.UserID, uint(itemID), *req.Available)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMenuItemNotFound),
			errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotYourRestaurant),
			errors.Is(err, service.ErrRestaurantClosed):
			utils.Error(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("update menu availability: %v", err)
			utils.Error(w, http.StatusInternalServerError, "failed to update availability")
		}
		return
	}

	utils.Success(w, http.StatusOK, "menu item availability updated", item)
}

func (h *MenuHandler) GetMenuItemDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	item, err := h.MenuService.GetMenuItemDetails(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrMenuItemNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("get menu item %d: %v", id, err)
		utils.Error(w, http.StatusInternalServerError, "unable to fetch menu item")
		return
	}

	utils.Success(w, http.StatusOK, "menu item retrieved", item)
}

func (h *MenuHandler) SearchMenuItems(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	search := service.MenuSearch{Query: params.Get("q")}

	var err error
	if search.CategoryID, err = parseUintParam(params.Get("category_id")); err != nil {
		utils.Error(w, http.StatusBadRequest, "category_id must be a positive number")
		return
	}
	if search.RestaurantID, err = parseUintParam(params.Get("restaurant_id")); err != nil {
		utils.Error(w, http.StatusBadRequest, "restaurant_id must be a positive number")
		return
	}
	if search.MinPrice, err = parseInt64Param(params.Get("min_price")); err != nil {
		utils.Error(w, http.StatusBadRequest, "min_price must be a number")
		return
	}
	if search.MaxPrice, err = parseInt64Param(params.Get("max_price")); err != nil {
		utils.Error(w, http.StatusBadRequest, "max_price must be a number")
		return
	}
	if raw := params.Get("available"); raw != "" {
		available, err := strconv.ParseBool(raw)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "available must be true or false")
			return
		}
		search.Available = &available
	}
	if search.Page, err = parseQueryInt(params.Get("page")); err != nil {
		utils.Error(w, http.StatusBadRequest, "page must be a number")
		return
	}
	if search.Limit, err = parseQueryInt(params.Get("limit")); err != nil {
		utils.Error(w, http.StatusBadRequest, "limit must be a number")
		return
	}

	items, err := h.MenuService.SearchMenuItems(r.Context(), search)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCategoryID),
			errors.Is(err, service.ErrInvalidRestaurantID),
			errors.Is(err, service.ErrInvalidPriceBound),
			errors.Is(err, service.ErrInvalidPriceRange):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("search menu: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to search menu items")
		}
		return
	}

	utils.Success(w, http.StatusOK, "menu items retrieved", items)
}

func parseUintParam(value string) (*uint, error) {
	if value == "" {
		return nil, nil
	}
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil || n == 0 {
		return nil, errors.New("must be a positive number")
	}
	id := uint(n)
	return &id, nil
}

func parseInt64Param(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func parseQueryInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}
