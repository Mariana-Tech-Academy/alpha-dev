package service

import (
	"context"
	"net/mail"
	"strings"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/repository"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type RegistrationValidationError struct {
	Fields map[string]string
}

func (err *RegistrationValidationError) Error() string {
	return "Check the registration fields"
}

type UserService interface {
	Register(ctx context.Context, req RegisterRequest) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	tokens   *auth.JWTManager
}

func NewUserService(
	repo repository.UserRepository,
	tokens *auth.JWTManager,
) UserService {
	return &userService{userRepo: repo, tokens: tokens}
}

func validateRegistration(req RegisterRequest) map[string]string {
	fields := make(map[string]string)

	if req.FirstName == "" {
		fields["first_name"] = "First name is required"
	} else if len([]rune(req.FirstName)) > 100 {
		fields["first_name"] = "First name must not exceed 100 characters"
	}

	if req.LastName == "" {
		fields["last_name"] = "Last name is required"
	} else if len([]rune(req.LastName)) > 100 {
		fields["last_name"] = "Last name must not exceed 100 characters"
	}

	if req.PhoneNumber == "" {
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

	// Keep our existing minimum until the team confirms the password rule.
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

func (s *userService) Register(
	ctx context.Context,
	req RegisterRequest,
) (*models.User, error) {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if fields := validateRegistration(req); len(fields) > 0 {
		return nil, &RegistrationValidationError{Fields: fields}
	}

	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, repository.ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.PhoneNumber,
		Password:  string(hash), // Store the hash in the professor's Password field.
		Role:      models.RoleCustomer,
		Status:    models.UserStatusActive,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
