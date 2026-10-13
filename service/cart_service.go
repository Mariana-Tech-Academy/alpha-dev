package service

import "alpha-dev/repository"

type CartService interface {
}

type cartService struct {
	cartRepo repository.CartRepository
}

func NewCartService(repo repository.CartRepository) CartService {
	return &cartService{cartRepo: repo}
}
