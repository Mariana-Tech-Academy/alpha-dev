package repository

import "gorm.io/gorm"

type RestaurantRepository interface {
}

type PostgresRestaurantRepo struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &PostgresRestaurantRepo{db: db}
}
