package repository

import "gorm.io/gorm"

type AdminRepository interface {
}

type PostgresAdminRepo struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &PostgresAdminRepo{db: db}
}
