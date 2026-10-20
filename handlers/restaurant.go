package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
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
	City        string `json:"city"`
	State       string `json:"state"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	OpeningTime string `json:"opening_time"`
	ClosingTime string `json:"closing_time"`
}

func (h *RestaurantHandler) RegisterRestaurant(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req RegisterRestaurantRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant JSON")
		return
	}

	restaurant := &models.Restaurant{
		ManagerID:   claims.UserID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Phone:       req.Phone,
		Email:       req.Email,
		OpeningTime: req.OpeningTime,
		ClosingTime: req.ClosingTime,
	}

	if err := h.RestaurantService.RegisterRestaurant(r.Context(), restaurant); err != nil {
		var validationErr *service.RestaurantValidationError
		switch {
		case errors.As(err, &validationErr):
			utils.ValidationError(w, "check the restaurant details", validationErr.Fields)
		case errors.Is(err, service.ErrMissingManager):
			utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		default:
			log.Printf("restaurant registration failed: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to register restaurant")
		}
		return
	}

	utils.Success(w, http.StatusCreated, "restaurant registered successfully", restaurant.Profile())
}

type RestaurantStatusRequest struct {
	Status models.RestaurantStatus `json:"status"`
}

func (h *RestaurantHandler) SetAvailability(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, strconv.IntSize)
	if err != nil || id == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var req RestaurantStatusRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid status JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid status JSON")
		return
	}

	err = h.RestaurantService.SetAvailability(
		r.Context(),
		claims.UserID,
		claims.Role,
		uint(id),
		req.Status,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrRestaurantForbidden),
			errors.Is(err, service.ErrRestaurantRestricted):
			utils.Error(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrRestaurantInvalidStatus):
			utils.ValidationError(w, "check the status field", map[string]string{
				"status": err.Error(),
			})
		case errors.Is(err, service.ErrRestaurantChanged):
			utils.Error(w, http.StatusConflict, err.Error())
		default:
			log.Printf("change restaurant availability: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to update restaurant")
		}
		return
	}

	utils.Success(w, http.StatusOK, "restaurant availability updated", map[string]any{
		"id":     uint(id),
		"status": req.Status,
	})
}

func (h *RestaurantHandler) UpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil || id == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant id")
		return
	}

	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var input models.UpdateRestaurantInput
	var file *multipart.FileHeader

	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid form data")
			return
		}

		if nameVal := r.FormValue("name"); nameVal != "" {
			input.Name = &nameVal
		}
		if descVal := r.FormValue("description"); descVal != "" {
			input.Description = &descVal
		}
		if addressVal := r.FormValue("address"); addressVal != "" {
			input.Address = &addressVal
		}
		if cityVal := r.FormValue("city"); cityVal != "" {
			input.City = &cityVal
		}
		if stateVal := r.FormValue("state"); stateVal != "" {
			input.State = &stateVal
		}
		if phoneVal := r.FormValue("phone"); phoneVal != "" {
			input.Phone = &phoneVal
		}
		if emailVal := r.FormValue("email"); emailVal != "" {
			input.Email = &emailVal
		}
		if openVal := r.FormValue("opening_time"); openVal != "" {
			input.OpeningTime = &openVal
		}
		if closeVal := r.FormValue("closing_time"); closeVal != "" {
			input.ClosingTime = &closeVal
		}
		if imageURLVal := r.FormValue("image_url"); imageURLVal != "" {
			input.ImageURL = &imageURLVal
		}
		if _, fileHeader, err := r.FormFile("image"); err == nil {
			file = fileHeader
		}
	} else {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
	}

	updated, err := h.RestaurantService.UpdateRestaurant(r.Context(), uint(id), claims.UserID, claims.Role, &input, file)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, "restaurant not found")
		case errors.Is(err, service.ErrRestaurantForbidden):
			utils.Error(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidImage):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("update restaurant %d: %v", id, err)
			utils.Error(w, http.StatusInternalServerError, "failed to update")
		}
		return
	}

	utils.Success(w, http.StatusOK, "restaurant updated", updated.Profile())
}
