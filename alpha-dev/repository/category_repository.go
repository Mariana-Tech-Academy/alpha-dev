package repository

import "gorm.io/gorm"

type CategoryRepository interface {
}

type PostgresCategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &PostgresCategoryRepo{db: db}
}
