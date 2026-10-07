package main

import (
	"log"
	"os"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if os.Getenv("PGPASSWORD") == "" {
		log.Fatal("Set PGPASSWORD before running this command")
	}

	dsn := "host=localhost port=5432 user=postgres dbname=foodhub sslmode=disable connect_timeout=5"

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
