package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"alpha-dev/auth"
	"alpha-dev/config"
	"alpha-dev/middleware"
	"alpha-dev/models"
	"alpha-dev/routes"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	// build connection
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	sslmode := os.Getenv("SSL_MODE")

	connStr := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=%s", host, user, password, port, name, sslmode)

	// open database
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to open db: ", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Restaurant{},
		&models.Category{},
		&models.MenuItem{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.RevokedToken{},
	)
	if err != nil {
		log.Fatal("failed to automigrate: ", err)
	}

	tokens := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry, db)
	router := routes.Router(db, tokens)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           middleware.CORS(cfg.CORSAllowedOrigins)(router),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("server is running on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("error starting server: ", err)
	}
}
