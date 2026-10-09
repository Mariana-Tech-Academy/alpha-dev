package main

import (
	"log"
	"os"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Set DATABASE_URL before running this command")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := db.Transaction(models.Migrate); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("FoodHub database tables created successfully")
}
