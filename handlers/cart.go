package handlers

import "alpha-dev/service"

type CartHandler struct {
	CartService service.CartService
}

func NewCartHandler(service service.CartService) *CartHandler {
	return &CartHandler{CartService: service}

}
