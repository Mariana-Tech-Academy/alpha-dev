package handlers

import "alpha-dev/service"

type RestaurantHandler struct {
	RestaurantService service.RestaurantService
}

func NewRestaurantHandler(service service.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{RestaurantService: service}

}
