package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func registrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() }) // Close after the test.

	if err := models.Migrate(db); err != nil {
		t.Fatal(err)
	}

	return db
}

func validRegistration() RegisterRequest {
	return RegisterRequest{
		FirstName:       "Ahmed",
		LastName:        "Hussain",
		Email:           "ahmed.test@example.com",
		PhoneNumber:     "07700900123",
		Password:        "TestPass123!",
		ConfirmPassword: "TestPass123!",
	}
}

func sendRegistration(t *testing.T, db *gorm.DB, details RegisterRequest) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(details)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost, "/api/auth/register", bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	registerHandler(db).ServeHTTP(response, req)
	return response
}

func TestRegistrationSuccess(t *testing.T) {
	db := registrationTestDB(t)
	details := validRegistration()
	details.FirstName = " Ahmed "
	details.Email = " AHMED.TEST@EXAMPLE.COM "

	response := sendRegistration(t, db, details)
	if response.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %s", response.Code, response.Body.String())
	}

	var user models.User
	if err := db.Where("email = ?", "ahmed.test@example.com").First(&user).Error; err != nil {
		t.Fatal(err)
	}

	if user.FirstName != "Ahmed" || user.LastName != details.LastName ||
		user.PhoneNumber != details.PhoneNumber {
		t.Fatal("Customer details were not saved correctly")
	}

	if user.Role != models.RoleCustomer || user.Status != models.UserActive {
		t.Fatal("Expected an active customer")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(details.Password),
	); err != nil {
		t.Fatal("Saved hash does not match the password")
	}

	var result struct {
		Success bool                       `json:"success"`
		Data    map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Success || len(result.Data) != 6 {
		t.Fatal("Expected success and six safe response fields")
	}

	for _, field := range []string{
		"first_name", "last_name", "email", "phone_number", "role", "status",
	} {
		if _, exists := result.Data[field]; !exists {
			t.Fatalf("Missing response field: %s", field)
		}
	}

	if bytes.Contains(response.Body.Bytes(), []byte(details.Password)) ||
		bytes.Contains(response.Body.Bytes(), []byte(user.PasswordHash)) {
		t.Fatal("Response exposed the password or hash")
	}
}

func TestRegistrationDuplicateEmail(t *testing.T) {
	db := registrationTestDB(t)
	details := validRegistration()

	if response := sendRegistration(t, db, details); response.Code != http.StatusCreated {
		t.Fatalf("First registration failed: %s", response.Body.String())
	}

	details.Email = "AHMED.TEST@EXAMPLE.COM" // Same email, different case.
	response := sendRegistration(t, db, details)
	if response.Code != http.StatusConflict {
		t.Fatalf("Expected 409, got %d", response.Code)
	}

	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Expected one customer, got %d", count)
	}
}

func TestRegistrationInvalidFields(t *testing.T) {
	tests := []struct {
		name   string
		change func(*RegisterRequest)
		field  string
	}{
		{"missing first name", func(r *RegisterRequest) { r.FirstName = "" }, "first_name"},
		{"missing last name", func(r *RegisterRequest) { r.LastName = "" }, "last_name"},
		{"missing email", func(r *RegisterRequest) { r.Email = "" }, "email"},
		{"invalid email", func(r *RegisterRequest) { r.Email = "invalid" }, "email"},
		{"missing phone", func(r *RegisterRequest) { r.PhoneNumber = "" }, "phone_number"},
		{"missing password", func(r *RegisterRequest) { r.Password = "" }, "password"},
		{"short password", func(r *RegisterRequest) {
			r.Password = "123"
			r.ConfirmPassword = "123"
		}, "password"},
		{"missing confirmation", func(r *RegisterRequest) { r.ConfirmPassword = "" }, "confirm_password"},
		{"password mismatch", func(r *RegisterRequest) { r.ConfirmPassword = "Different123!" }, "confirm_password"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := registrationTestDB(t)
			details := validRegistration()
			test.change(&details)

			response := sendRegistration(t, db, details)
			if response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("Expected 422, got %d", response.Code)
			}

			var result struct {
				Errors map[string]string `json:"errors"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Errors[test.field] == "" {
				t.Fatalf("Expected an error for %s", test.field)
			}

			var count int64
			if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != 0 {
				t.Fatal("Invalid registration saved a customer")
			}
		})
	}
}

func TestRegistrationAllFieldsMissing(t *testing.T) {
	response := sendRegistration(t, nil, RegisterRequest{})

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected 422, got %d", response.Code)
	}

	var result struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}

	if len(result.Errors) != 6 {
		t.Fatalf("Expected six field errors, got %d", len(result.Errors))
	}
}

func TestRegistrationInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost, "/api/auth/register", bytes.NewBufferString("{"),
	)
	response := httptest.NewRecorder()
	registerHandler(nil).ServeHTTP(response, req)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400, got %d", response.Code)
	}
}

func TestRegistrationWrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/register", nil)
	response := httptest.NewRecorder()
	registerHandler(nil).ServeHTTP(response, req)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected 405, got %d", response.Code)
	}
	if response.Header().Get("Allow") != "POST" {
		t.Fatal("Expected Allow: POST")
	}
}
