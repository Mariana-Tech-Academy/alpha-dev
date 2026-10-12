package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

//Check Method (POST)

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Get JSON from Request body

	var req loginRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Search Database to find user.
	var user models.User

	err = DB.Where("email =?", req.Email).First(&user).Error
	if err != nil {
		http.Error(w, "Invalid email or Password", http.StatusUnauthorized)
		return
	}

	// Check if suspended account
	if user.Status == "suspended" {
		http.Error(w, "Account is suspended", http.StatusForbidden)
		return
	}

	// Compare Passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	// Generate a 24hrs token
	expiry := time.Now().Add(time.Hour * 24)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     expiry.Unix(),
	})

	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	// Return Success

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Login sucessful",
		"data": map[string]any{
			"token":   tokenStr,
			"expires": expiry,
			"user":    user,
		},
	})

}
