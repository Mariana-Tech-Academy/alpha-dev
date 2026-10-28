package service

import (
	"context"
	"errors"
	"strings"

	"alpha-dev/models"

	"gorm.io/gorm"
)

// UpdateMenuItemRequest is the body for PUT /menu/{id}.
// Price is an integer in minor units, so 1250 means 12.50.
type UpdateMenuItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	ImageURL    string `json:"image_url"`
	CategoryID  uint   `json:"category_id"`
}

func (s *menuService) findOwnedMenuItem(ctx context.Context, userID, id uint) (models.MenuItem, error) {
	item, err := s.menuRepo.FindMenuItemByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.MenuItem{}, ErrMenuItemNotFound
	}
	if err != nil {
		return models.MenuItem{}, err
	}

	restaurant, err := s.menuRepo.FindRestaurantByID(ctx, item.RestaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.MenuItem{}, ErrMenuItemNotFound
	}
	if err != nil {
		return models.MenuItem{}, err
	}
	if restaurant.ManagerID != userID {
		return models.MenuItem{}, ErrNotYourRestaurant
	}
	if restaurant.Status != models.RestaurantStatusActive && restaurant.Status != models.RestaurantStatusInactive {
		return models.MenuItem{}, ErrRestaurantClosed
	}
	return item, nil
}

func (s *menuService) UpdateMenuItem(ctx context.Context, userID uint, id uint, req UpdateMenuItemRequest) (*models.MenuItem, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.ImageURL = strings.TrimSpace(req.ImageURL)
	if req.Name == "" {
		return nil, ErrNameRequired
	}
	if req.Price <= 0 {
		return nil, ErrInvalidPrice
	}
	if req.CategoryID == 0 {
		return nil, ErrCategoryIDRequired
	}

	item, err := s.findOwnedMenuItem(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if _, err := s.menuRepo.FindCategoryByID(ctx, req.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	fields := map[string]any{
		"name":        req.Name,
		"description": req.Description,
		"price":       req.Price,
		"image_url":   req.ImageURL,
		"category_id": req.CategoryID,
	}
	if err := s.menuRepo.UpdateMenuItem(ctx, item.ID, fields); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuItemNotFound
		}
		return nil, err
	}

	item.Name = req.Name
	item.Description = req.Description
	item.Price = req.Price
	item.ImageURL = req.ImageURL
	item.CategoryID = req.CategoryID
	item.Restaurant = models.Restaurant{}
	item.Category = models.Category{}
	return &item, nil
}

func (s *menuService) DeleteMenuItem(ctx context.Context, userID uint, id uint) error {
	item, err := s.findOwnedMenuItem(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := s.menuRepo.DeleteMenuItem(ctx, item.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuItemNotFound
		}
		return err
	}
	return nil
}
