package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Mariana-Tech-Academy/alpha-dev/handlers"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB failed: %v", err)
	}
	handlers.DB = db
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", handlers.Login)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
