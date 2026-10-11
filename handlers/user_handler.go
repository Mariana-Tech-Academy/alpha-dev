package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Mariana-Tech-Academy/alpha-dev/service"
)

type ChangePassword struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{userService: service}
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	//Check if the user is logged in
	
	//Get User ID from auth middleware context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false, 
			"message": "Unauthorized",
		})
		return
	}
	userID := userIDVal.(uint)

	//Collect and validate the request payload
	var req ChangePassword
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false, 
			"message": "Invalid request payload",
		})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false, 
			"message": "Passwords do not match",
		})
		return
	}

	//Call the service layer to change the password
	err := h.userService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "incorrect current password" {
			statusCode = http.StatusUnauthorized
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false, 
			"message": err.Error(),
		})
		return
	}

	//Response for successful password change
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}