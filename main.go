package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL") // Read the database connection.
	if dsn == "" {
		log.Fatal("Set DATABASE_URL before starting the API")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Unable to connect to the database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Unable to access the database connection")
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Database is not responding")
	}

	router := http.NewServeMux()
	router.HandleFunc("/api/auth/register", registerHandler(db))

	server := &http.Server{
		Addr:              "127.0.0.1:8080", // Run locally for testing.
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Registration API running at http://127.0.0.1:8080")
	log.Fatal(server.ListenAndServe())
}
