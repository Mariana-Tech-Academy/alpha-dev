// routes/router.go — connects the components

package routes

import (
	"net/http"

	"github.com/Mariana-Tech-Academy/alpha-dev/handlers"
	"github.com/Mariana-Tech-Academy/alpha-dev/middleware"
	"github.com/Mariana-Tech-Academy/alpha-dev/repository"
	"github.com/Mariana-Tech-Academy/alpha-dev/service"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, secret []byte) *http.ServeMux {
	mux := http.NewServeMux()

	userRepository := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepository)
	authMiddleware := middleware.NewAuthMiddleware(authService, secret)
	authHandler := handlers.NewAuthHandler()

	RegisterAuthRoutes(mux, authHandler, authMiddleware)

	return mux
}
