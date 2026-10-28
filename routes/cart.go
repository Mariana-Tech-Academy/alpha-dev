package routes

import (
	"alpha-dev/auth"
	"alpha-dev/handlers"
	"alpha-dev/middleware"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type menuFinderAdapter struct {
	db *gorm.DB
}

func (m *menuFinderAdapter) FindByID(ctx context.Context, id uint) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := m.db.WithContext(ctx).Preload("Restaurant").First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func CartRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	cartRepo := repository.NewCartRepository(db)
	menuRepo := &menuFinderAdapter{db: db}

	cartService := service.NewCartService(cartRepo, menuRepo)
	cartHandler := handlers.NewCartHandler(cartService)

	customer := protected(router, tokens)
	customer.Use(middleware.RequireRole(models.RoleCustomer))

	// ACTION: CART-02 - Only AddItem exists in handler right now
	customer.HandleFunc(
		"/cart/items",
		cartHandler.AddItem,
	).Methods(http.MethodPost)
}
