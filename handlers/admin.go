package handlers

import (
	"alpha-dev/auth"
	"alpha-dev/service"
	"fmt"
	"net/http"
)

type AdminHandler struct {
	AdminService service.AdminService
}

func NewAdminHandler(service service.AdminService) *AdminHandler {
	return &AdminHandler{AdminService: service}

}

func (h *AdminHandler) TestAdmin(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		http.Error(w, "UnathoriZed", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "Welcome Admin! Your User ID is %d and your role is %s", claims.UserID, claims.Role)
}
