package service

import (
	"alpha-dev/models"
	"alpha-dev/repository"
	"errors"
	"strings"
)

type CategoryService interface {
	CreateCategory(name string) (*models.Category, error)
	GetCategories() ([]models.Category, error)
	GetCategoryByID(id uint) (*models.Category, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: repo}
}

// Validate name, then save it via the repository
func (s *categoryService) CreateCategory(name string) (*models.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	cat := &models.Category{Name: name}
	err := s.categoryRepo.Create(cat)
	if err != nil {
		return nil, err
	}
	return cat, nil
}

// Get all the categiories from the repository
func (s *categoryService) GetCategories() ([]models.Category, error) {
	return s.categoryRepo.FindAll()

}

// Get one category by ID from the repository
func (s *categoryService) GetCategoryByID(id uint) (*models.Category, error) {
	return s.categoryRepo.FindByID(id)

}
