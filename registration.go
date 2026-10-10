package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	FirstName       string `json:"first_name"`       // First name.
	LastName        string `json:"last_name"`        // Last name.
	Email           string `json:"email"`            // Email address.
	PhoneNumber     string `json:"phone_number"`     // Phone number.
	Password        string `json:"password"`         // Password to hash.
	ConfirmPassword string `json:"confirm_password"` // Must match.
}

// Flow: receive details → validate → check email → hash → save → respond.

func validateRegistration(req RegisterRequest) map[string]string {
	fields := make(map[string]string) // Collect all field errors.

	if strings.TrimSpace(req.FirstName) == "" {
		fields["first_name"] = "First name is required"
	} else if len([]rune(req.FirstName)) > 100 {
		fields["first_name"] = "First name must not exceed 100 characters"
	}

	if strings.TrimSpace(req.LastName) == "" {
		fields["last_name"] = "Last name is required"
	} else if len([]rune(req.LastName)) > 100 {
		fields["last_name"] = "Last name must not exceed 100 characters"
	}

	if strings.TrimSpace(req.PhoneNumber) == "" {
		fields["phone_number"] = "Phone number is required"
	} else if len([]rune(req.PhoneNumber)) > 30 {
		fields["phone_number"] = "Phone number must not exceed 30 characters"
	}

	if req.Email == "" {
		fields["email"] = "Email is required"
	} else {
		address, err := mail.ParseAddress(req.Email)
		if err != nil || address.Address != req.Email || len(req.Email) > 254 {
			fields["email"] = "Enter a valid email address"
		}
	}

	// Provisional minimum; confirm the password rule with the team.
	if req.Password == "" {
		fields["password"] = "Password is required"
	} else if len([]rune(req.Password)) < 8 {
		fields["password"] = "Password must contain at least 8 characters"
	} else if len(req.Password) > 72 {
		fields["password"] = "Password must not exceed 72 bytes"
	}

	if req.ConfirmPassword == "" {
		fields["confirm_password"] = "Confirm password is required"
	} else if req.Password != req.ConfirmPassword {
		fields["confirm_password"] = "Passwords do not match"
	}

	return fields
}

func registrationJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) // Send the response as JSON.
}

func registerHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			registrationJSON(w, http.StatusMethodNotAllowed, map[string]any{
				"success": false,
				"message": "Use POST to register",
			})
			return
		}

		var req RegisterRequest
		r.Body = http.MaxBytesReader(w, r.Body, 16*1024) // Limit request size.
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			registrationJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"message": "Invalid registration JSON",
			})
			return
		}

		req.FirstName = strings.TrimSpace(req.FirstName)
		req.LastName = strings.TrimSpace(req.LastName)
		req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))

		if fields := validateRegistration(req); len(fields) > 0 {
			registrationJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"success": false,
				"message": "Check the registration fields",
				"errors":  fields,
			})
			return
		}

		database := db.WithContext(r.Context())
		var count int64

		if err := database.Model(&models.User{}).
			Where("LOWER(email) = ?", req.Email).Count(&count).Error; err != nil {
			registrationJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"message": "Unable to check email",
			})
			return
		}

		if count > 0 {
			registrationJSON(w, http.StatusConflict, map[string]any{
				"success": false,
				"message": "Email is already taken",
			})
			return
		}

		hash, err := bcrypt.GenerateFromPassword(
			[]byte(req.Password), bcrypt.DefaultCost,
		)
		if err != nil {
			registrationJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"message": "Unable to process password",
			})
			return
		}

		user := models.User{
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Email:        req.Email,
			PhoneNumber:  req.PhoneNumber,
			PasswordHash: string(hash), // Save only the hash.
			Role:         models.RoleCustomer,
			Status:       models.UserActive,
		}

		if err := database.Create(&user).Error; err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) && postgresError.Code == "23505" {
				// Also catch duplicates detected during saving.
				registrationJSON(w, http.StatusConflict, map[string]any{
					"success": false,
					"message": "Email is already taken",
				})
				return
			}

			registrationJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"message": "Unable to create account",
			})
			return
		}

		registrationJSON(w, http.StatusCreated, map[string]any{
			"success": true,
			"message": "Account created successfully",
			"data": map[string]any{ // Return only these safe fields.
				"first_name":   user.FirstName,
				"last_name":    user.LastName,
				"email":        user.Email,
				"phone_number": user.PhoneNumber,
				"role":         user.Role,
				"status":       user.Status,
			},
		})
	}
}
