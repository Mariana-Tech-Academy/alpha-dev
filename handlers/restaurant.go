package handlers

import (
	"alpha-dev/service"
	"alpha-dev/utils"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type RestaurantHandler struct {
    restaurantService service.RestaurantService
}

func NewRestaurantHandler(service service.RestaurantService) *RestaurantHandler {
    return &RestaurantHandler{restaurantService: service}
}

// GetRestaurantDetails handles the GET /restaurants/{id} endpoint to retrieve restaurant details.
func (h *RestaurantHandler) GetRestaurantDetails(w http.ResponseWriter, r *http.Request) {
    // Extract the restaurant ID from the URL path
    vars := mux.Vars(r)
    idStr := vars["id"]

    // Convert the ID from string to uint
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        utils.Error(w, http.StatusBadRequest, "Invalid restaurant ID format")
        return
    }
    // Call the service to get restaurant details
    restaurant, err := h.restaurantService.GetRestaurantDetails(r.Context(), uint(id))
    if err != nil {
        utils.Error(w, http.StatusNotFound, err.Error())
        return
    }
    // Return the restaurant details in the response
    utils.Success(w, http.StatusOK, "Restaurant retrieved successfully", restaurant)
}

func (h *RestaurantHandler) SearchRestaurants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract query parameter from URL: e.g. /api/restaurants/search?q=London or ?q=Pizza
	req := r.URL.Query().Get("q")

	restaurants, err := h.restaurantService.SearchRestaurants(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch search results"})
		return
	}

	// Always returns 200 OK with [] if empty, or populated list if matches exist
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(restaurants)
}