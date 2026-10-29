package routes

import (
	"net/http"

	"alpha-dev/auth"
	"alpha-dev/middleware"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func Router(db *gorm.DB, tokens *auth.JWTManager) *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		utils.Error(w, http.StatusNotFound, "route not found")
	})
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		utils.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		utils.Success(w, http.StatusOK, "ok", nil)
	}).Methods(http.MethodGet)

	UserRoutes(api, db, tokens)
	RestaurantRoutes(api, db, tokens)
	MenuRoutes(api, db, tokens)
	CategoryRoutes(api, db, tokens)
	CartRoutes(api, db, tokens)
	CheckoutRoutes(api, db, tokens)   
	OrderRoutes(api, db, tokens)
	AdminRoutes(api, db, tokens)

	return r
}

// protected returns a subrouter whose routes require a valid JWT.
func protected(router *mux.Router, tokens *auth.JWTManager) *mux.Router {
	sub := router.NewRoute().Subrouter()
	sub.Use(middleware.Authenticate(tokens))
	return sub
}
