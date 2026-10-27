package routes

import (
	"alpha-dev/auth"
	"alpha-dev/handlers"
	"alpha-dev/middleware"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func CategoryRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(repo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// public routes: GET /categories
	router.HandleFunc("/categories", categoryHandler.GetCategories).Methods(http.MethodGet)
	router.HandleFunc("/categories/{id}", categoryHandler.GetCategory).Methods(http.MethodGet)

	manager := protected(router, tokens)
	manager.Use(middleware.RequireRole(models.RoleManager))

	manager.HandleFunc("/categories", categoryHandler.CreateCategory).Methods(http.MethodPost)
}
