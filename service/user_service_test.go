package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"alpha-dev/models"
	"alpha-dev/utils"

	"gorm.io/gorm"
)

func validRequest() RegisterRequest {
	return RegisterRequest{
		FirstName:       "Ada",
		LastName:        "Lovelace",
		Email:           "ada@example.com",
		Phone:           "+44 7700 900123",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
}

func TestValidateRegistration(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*RegisterRequest)
		wantField string
	}{
		{"valid", func(*RegisterRequest) {}, ""},
		{"missing first name", func(r *RegisterRequest) { r.FirstName = "" }, "first_name"},
		{"long last name", func(r *RegisterRequest) { r.LastName = strings.Repeat("a", 101) }, "last_name"},
		{"missing phone", func(r *RegisterRequest) { r.Phone = "" }, "phone"},
		{"phone with letters", func(r *RegisterRequest) { r.Phone = "hello" }, "phone"},
		{"phone too short", func(r *RegisterRequest) { r.Phone = "123" }, "phone"},
		{"invalid email", func(r *RegisterRequest) { r.Email = "not-an-email" }, "email"},
		{"email with display name", func(r *RegisterRequest) { r.Email = "Ada <ada@example.com>" }, "email"},
		{"short password", func(r *RegisterRequest) {
			r.Password, r.ConfirmPassword = "short", "short"
		}, "password"},
		{"password over 72 bytes", func(r *RegisterRequest) {
			r.Password = strings.Repeat("a", 73)
			r.ConfirmPassword = r.Password
		}, "password"},
		{"passwords do not match", func(r *RegisterRequest) { r.ConfirmPassword = "different123" }, "confirm_password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validRequest()
			tt.modify(&req)
			fields := validateRegistration(req)

			if tt.wantField == "" {
				if len(fields) != 0 {
					t.Fatalf("expected no errors, got %v", fields)
				}
				return
			}
			if _, ok := fields[tt.wantField]; !ok {
				t.Fatalf("expected error on %q, got %v", tt.wantField, fields)
			}
		})
	}
}

type fakeUserRepo struct {
	emailExists bool
	createErr   error
	created     *models.User
	found       models.User
	findErr     error
}

func (f *fakeUserRepo) FindByID(context.Context, uint) (models.User, error) {
	return f.found, f.findErr
}

func (f *fakeUserRepo) EmailExists(context.Context, string) (bool, error) {
	return f.emailExists, nil
}

func (f *fakeUserRepo) CreateUser(_ context.Context, user *models.User) error {
	f.created = user
	return f.createErr
}

func TestRegister(t *testing.T) {
	t.Run("creates active customer with hashed password", func(t *testing.T) {
		repo := &fakeUserRepo{}
		req := validRequest()
		req.Email = "  ADA@Example.com "

		user, err := NewUserService(repo, nil).Register(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Email != "ada@example.com" {
			t.Errorf("email not normalised: %q", user.Email)
		}
		if user.Role != models.RoleCustomer || user.Status != models.UserStatusActive {
			t.Errorf("unexpected role/status: %s/%s", user.Role, user.Status)
		}
		if !utils.CheckPassword(user.Password, req.Password) {
			t.Error("password was not hashed correctly")
		}
	})

	t.Run("existing email", func(t *testing.T) {
		repo := &fakeUserRepo{emailExists: true}
		_, err := NewUserService(repo, nil).Register(context.Background(), validRequest())
		if !errors.Is(err, ErrEmailTaken) {
			t.Fatalf("expected ErrEmailTaken, got %v", err)
		}
	})

	t.Run("duplicate key on insert", func(t *testing.T) {
		repo := &fakeUserRepo{createErr: gorm.ErrDuplicatedKey}
		_, err := NewUserService(repo, nil).Register(context.Background(), validRequest())
		if !errors.Is(err, ErrEmailTaken) {
			t.Fatalf("expected ErrEmailTaken, got %v", err)
		}
	})

	t.Run("validation error on empty email", func(t *testing.T) {
		req := validRequest()
		req.Email = ""
		_, err := NewUserService(&fakeUserRepo{}, nil).Register(context.Background(), req)
		var validationError *RegistrationValidationError
		if !errors.As(err, &validationError) {
			t.Fatalf("expected validation error, got %v", err)
		}
	})
}

func TestGetActiveUser(t *testing.T) {
	tests := []struct {
		name    string
		repo    *fakeUserRepo
		wantErr error
	}{
		{"active user", &fakeUserRepo{found: models.User{Status: models.UserStatusActive}}, nil},
		{"suspended user", &fakeUserRepo{found: models.User{Status: models.UserStatusSuspended}}, ErrInactiveUser},
		{"missing user", &fakeUserRepo{findErr: gorm.ErrRecordNotFound}, ErrUserNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUserService(tt.repo, nil).GetActiveUser(context.Background(), 1)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
