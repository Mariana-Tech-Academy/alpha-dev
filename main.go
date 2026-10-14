package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"alpha-dev/auth"
	"alpha-dev/config"
	"alpha-dev/middleware"
	"alpha-dev/models"
	"alpha-dev/routes"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Server holds our database connection so handlers can access it
type Server struct {
	db *gorm.DB
}

func main() {
	// 1. Load application configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	// 2. Build PostgreSQL connection string from environment variables
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	sslmode := os.Getenv("SSL_MODE")

	connStr := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=%s",
		host, user, password, port, name, sslmode)

	// 3. Open PostgreSQL database connection
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to open db: ", err)
	}

	// 4. Run database migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Restaurant{},
		&models.Category{},
		&models.MenuItem{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
	)
	if err != nil {
		log.Fatal("failed to automigrate: ", err)
	}

	// 5. Seed a test user if the users table is completely empty
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		db.Create(&models.User{
			FirstName: "Fred",
			LastName:  "Flinstone",
			Email:     "Bedrock@telly.com",
			Phone:     "07799 123890",
		})
		log.Println("Seeded initial test user into PostgreSQL.")
	}

	// 6. Initialize JWT and Router
	tokens := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)
	router := routes.Router(db, tokens)

	// 7. Bind the profile update handler to our Server struct instance
	srv := &Server{db: db}
	// Note: If you are using standard library mux, or if you want to add it directly:
	http.HandleFunc("/profile/", srv.handleUpdateProfile)

	// 8. Start the HTTP server
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           middleware.CORS(cfg.CORSAllowedOrigins)(router),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("Server is running on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("error starting server: ", err)
	}
}

// handleUpdateProfile processes PUT/PATCH requests to update user profiles
func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path (e.g., /profile/1 -> 1)
	idStr := strings.TrimPrefix(r.URL.Path, "/profile/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Find the user in PostgreSQL
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Decode incoming JSON update payload
	var input models.ProfileUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Apply updates using your model's logic
	if err := user.UpdateProfile(s.db, input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Profile updated successfully",
		"user":    user,
	})
}
