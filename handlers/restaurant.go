package handlers

import (
	"alpha-dev/utils"
	"errors"
	"log"

	"encoding/json"
	"net/http"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
)

type RestaurantHandler struct {
	RestaurantService service.RestaurantService
}

func NewRestaurantHandler(s service.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{RestaurantService: s}
}

type RegisterRestaurantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	OpeningTime string `json:"opening_time"`
	ClosingTime string `json:"closing_time"`
}

func (h *RestaurantHandler) RegisterRestaurant(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != models.RoleManager &&
		claims.Role != models.RoleAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req RegisterRestaurantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	restaurant := &models.Restaurant{
		ManagerID:   claims.UserID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Phone:       req.Phone,
		Email:       req.Email,
		OpeningTime: req.OpeningTime,
		ClosingTime: req.ClosingTime,
	}

	if err := h.RestaurantService.RegisterRestaurant(
		r.Context(), restaurant,
	); err != nil {
		var validationErr *service.RestaurantValidationError

		if errors.As(err, &validationErr) {
			utils.ValidationError(
				w,
				"Please correct the restaurant details",
				validationErr.Fields,
			)
			return
		}

		log.Printf("restaurant registration failed: %v", err)

		utils.Error(
			w,
			http.StatusInternalServerError,
			"Unable to register restaurant",
		)
		return
	}

	utils.Success(
		w,
		http.StatusCreated,
		"Restaurant registered successfully",
		restaurant,
	)
}
