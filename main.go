package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Mariana-Tech-Academy/alpha-dev/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("Set DATABASE_URL before starting the server")
	}

	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) < 32 {
		log.Fatal("Set JWT_SECRET to a random secret of at least 32 bytes")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Unable to connect to the database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Unable to initialise the database connection")
	}
	defer sqlDB.Close()

	mux := http.NewServeMux()
	auth.RegisterRoutes(mux, db, secret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("FoodHub API listening on port %s", port)
	log.Fatal(server.ListenAndServe())
}
