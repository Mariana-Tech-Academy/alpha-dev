package routes

import (
	"alpha-dev/auth"
	"alpha-dev/handlers"
	"alpha-dev/middleware"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func CategoryRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(repo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// public routes: GET /categories

	// admin routes: POST /categories, PUT /categories/{id}, DELETE /categories/{id}
	admin := protected(router, tokens)
	admin.Use(middleware.RequireRole(models.RoleAdmin))

	_, _ = categoryHandler, admin
}
