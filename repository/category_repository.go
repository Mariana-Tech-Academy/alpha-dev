package repository

import (
	"alpha-dev/models"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(cat *models.Category) error
	FindAll() ([]models.Category, error)
	FindByID(id uint) (*models.Category, error)
}

type PostgresCategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &PostgresCategoryRepo{db: db}
}

//Create a new category in the database

func (r *PostgresCategoryRepo) Create(cat *models.Category) error {
	return r.db.Create(cat).Error
}

// Get all categories with thier menu items with it.

func (r *PostgresCategoryRepo) FindAll() ([]models.Category, error) {
	var cats []models.Category
	err := r.db.Preload("MenuItems").Find(&cats).Error

	return cats, err
}

// Get a category by ID wirh its Menu items loaded.
func (r *PostgresCategoryRepo) FindByID(id uint) (*models.Category, error) {
	var cat models.Category
	err := r.db.Preload("MenuItems").Find(&cat, id).Error

	if err != nil {
		return nil, err
	}
	return &cat, nil

}
