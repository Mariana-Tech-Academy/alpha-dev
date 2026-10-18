package handlers

import (

 "alpha-dev/service"
 "encoding/json"
 "errors"
 "net/http"

 "alpha-dev/auth"
 "alpha-dev/utils"

)

type MenuHandler struct {
	MenuService service.MenuService
}
//the Authenticate middleware checks the login token and stores the user’s details on the request.
//r.Context() is the request’s context, which holds those details.
//auth.ClaimsFromContext takes the details out. You get claims.UserID (who they are) and claims.Role.
//If something is wrong, it replies 401 and return stops the function.

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
		// Who is logged in?
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Read the JSON the client sent

	var req service.CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid menu item JSON")
		return
	}


	//// Ask the service to do the work

	item, err := h.MenuService.CreateMenuItem(r.Context(), claims.UserID, req)
	if errors.Is(err, service.ErrInvalidInput) {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrNotYourRestaurant) {
		utils.Error(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "unable to create menu item")
		return
	}

	utils.Success(w, http.StatusCreated, "menu item created", item)
}

func NewMenuHandler(service service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: service}

}
