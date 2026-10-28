package service

import (
	"context"
	"errors"
	"strings"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

const maxCategoryNameLength = 100

var ErrCategoryExists = errors.New("category already exists")

type CategoryValidationError struct {
	Fields map[string]string
}

func (err *CategoryValidationError) Error() string {
	return "invalid category"
}

type CategoryService interface {
	CreateCategory(ctx context.Context, name string) (models.CategorySummary, error)
	GetCategories(ctx context.Context) ([]models.CategorySummary, error)
	GetCategoryByID(ctx context.Context, id uint) (models.CategoryDetails, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, name string) (models.CategorySummary, error) {
	name = strings.TrimSpace(name)
	fields := map[string]string{}
	if name == "" {
		fields["name"] = "name is required"
	} else if len([]rune(name)) > maxCategoryNameLength {
		fields["name"] = "name must not exceed 100 characters"
	}
	if len(fields) > 0 {
		return models.CategorySummary{}, &CategoryValidationError{Fields: fields}
	}

	cat := &models.Category{Name: name}
	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return models.CategorySummary{}, ErrCategoryExists
		}
		return models.CategorySummary{}, err
	}
	return models.CategorySummary{ID: cat.ID, Name: cat.Name}, nil
}

func (s *categoryService) GetCategories(ctx context.Context) ([]models.CategorySummary, error) {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	summaries := make([]models.CategorySummary, 0, len(categories))
	for _, category := range categories {
		summaries = append(summaries, models.CategorySummary{ID: category.ID, Name: category.Name})
	}
	return summaries, nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id uint) (models.CategoryDetails, error) {
	if id == 0 {
		return models.CategoryDetails{}, ErrCategoryNotFound
	}
	category, err := s.categoryRepo.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.CategoryDetails{}, ErrCategoryNotFound
	}
	if err != nil {
		return models.CategoryDetails{}, err
	}
	items, err := s.categoryRepo.FindPublicMenuItems(ctx, category.ID)
	if err != nil {
		return models.CategoryDetails{}, err
	}
	if items == nil {
		items = []models.CategoryMenuItem{}
	}
	return models.CategoryDetails{
		ID:        category.ID,
		Name:      category.Name,
		MenuItems: items,
	}, nil
}
