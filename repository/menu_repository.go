package repository

import "gorm.io/gorm"

type MenuRepository interface {
}

type PostgresMenuRepo struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &PostgresMenuRepo{db: db}
}
