package main

import (
	"log"
	"os"

	"github.com/Mariana-Tech-Academy/alpha-dev/handlers"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB failed: %v", err)
	}
	handlers.DB = db
	log.Println("Migration ready")
	_ = os.Getenv("PORT")
}
