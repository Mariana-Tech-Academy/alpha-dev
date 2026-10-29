package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

type fakeCategoryRepo struct {
	categories []models.Category
	items      []models.CategoryMenuItem
	nextID     uint
}

func (f *fakeCategoryRepo) Create(_ context.Context, cat *models.Category) error {
	for _, existing := range f.categories {
		if existing.Name == cat.Name {
			return gorm.ErrDuplicatedKey
		}
	}
	f.nextID++
	cat.ID = f.nextID
	f.categories = append(f.categories, *cat)
	return nil
}

func (f *fakeCategoryRepo) FindAll(_ context.Context, page, limit int) ([]models.Category, int64, error) {
	total := int64(len(f.categories))
	start := min((page-1)*limit, len(f.categories))
	end := min(start+limit, len(f.categories))
	return f.categories[start:end], total, nil
}

func (f *fakeCategoryRepo) FindByID(_ context.Context, id uint) (models.Category, error) {
	for _, category := range f.categories {
		if category.ID == id {
			return category, nil
		}
	}
	return models.Category{}, gorm.ErrRecordNotFound
}

func (f *fakeCategoryRepo) FindPublicMenuItems(context.Context, uint) ([]models.CategoryMenuItem, error) {
	if f.items == nil {
		return []models.CategoryMenuItem{}, nil
	}
	return f.items, nil
}

var _ repository.CategoryRepository = (*fakeCategoryRepo)(nil)

func TestCreateCategory(t *testing.T) {
	ctx := context.Background()
	repo := &fakeCategoryRepo{}
	svc := NewCategoryService(repo)

	created, err := svc.CreateCategory(ctx, "  Pizza  ")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if created.Name != "Pizza" || created.ID == 0 {
		t.Fatalf("stored %#v", created)
	}

	_, err = svc.CreateCategory(ctx, "   ")
	var validationErr *CategoryValidationError
	if !errors.As(err, &validationErr) || validationErr.Fields["name"] != "name is required" {
		t.Fatalf("empty name: %v", err)
	}

	_, err = svc.CreateCategory(ctx, "Pizza")
	if !errors.Is(err, ErrCategoryExists) {
		t.Fatalf("duplicate name: %v", err)
	}
}

func TestGetCategories(t *testing.T) {
	repo := &fakeCategoryRepo{categories: []models.Category{
		{Name: "Pizza"},
		{Name: "Drinks"},
	}}
	repo.categories[0].ID = 1
	repo.categories[1].ID = 2
	svc := NewCategoryService(repo)

	result, err := svc.GetCategories(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	categories := result.Items.([]models.CategorySummary)
	if len(categories) != 2 || categories[0].Name != "Pizza" || categories[1].Name != "Drinks" {
		t.Fatalf("categories = %#v", categories)
	}
	if result.Page != 1 || result.Limit != models.DefaultPageSize || result.Total != 2 {
		t.Fatalf("page = %d limit = %d total = %d", result.Page, result.Limit, result.Total)
	}

	result, err = svc.GetCategories(context.Background(), 2, 1)
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	categories = result.Items.([]models.CategorySummary)
	if len(categories) != 1 || categories[0].Name != "Drinks" || result.Page != 2 || result.Total != 2 {
		t.Fatalf("second page = %#v", result)
	}
}

func TestGetCategoryByID(t *testing.T) {
	repo := &fakeCategoryRepo{
		categories: []models.Category{{Name: "Pizza"}},
		items: []models.CategoryMenuItem{{
			ID: 9, Name: "Margherita", Price: 1250, RestaurantID: 3, RestaurantName: "Luigi",
		}},
	}
	repo.categories[0].ID = 4
	svc := NewCategoryService(repo)

	category, err := svc.GetCategoryByID(context.Background(), 4)
	if err != nil {
		t.Fatalf("get category: %v", err)
	}
	if category.Name != "Pizza" || len(category.MenuItems) != 1 || category.MenuItems[0].Name != "Margherita" {
		t.Fatalf("category = %#v", category)
	}

	_, err = svc.GetCategoryByID(context.Background(), 99)
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("missing category: %v", err)
	}
}
