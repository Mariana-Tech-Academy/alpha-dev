package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/utils"

	"gorm.io/gorm"
)

var (
	ErrEmailTaken         = errors.New("email is already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrInactiveUser       = errors.New("account is not active")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrIncorrectPassword  = errors.New("incorrect current password")
	ErrPasswordMismatch   = errors.New("passwords do not match")
	ErrWeakPassword       = errors.New("password must contain at least 8 characters")
	ErrPasswordTooLong    = errors.New("password must not exceed 72 bytes")
	ErrSamePassword       = errors.New("new password must be different from the current password")
	ErrPasswordRequired   = errors.New("password is required")
	ErrConfirmRequired    = errors.New("confirm password is required")
)

type RegisterRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	Token string             `json:"token"`
	User  models.UserProfile `json:"user"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type RegistrationValidationError struct {
	Fields map[string]string
}

func (err *RegistrationValidationError) Error() string {
	return "invalid registration fields"
}

type UserService interface {
	Register(ctx context.Context, req RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResult, error)
	GetActiveUser(ctx context.Context, id uint) (models.User, error)
	Logout(ctx context.Context, claims *auth.Claims) error
	ChangePassword(ctx context.Context, claims *auth.Claims, req ChangePasswordRequest) error
	UpdateUserProfile(ctx context.Context, userID uint, input models.ProfileUpdate) error
}

type userService struct {
	userRepo repository.UserRepository
	tokens   *auth.JWTManager
}

func NewUserService(repo repository.UserRepository, tokens *auth.JWTManager) UserService {
	return &userService{userRepo: repo, tokens: tokens}
}

func validPhone(phone string) bool {
	digits := 0
	for _, c := range phone {
		switch {
		case c >= '0' && c <= '9':
			digits++
		case c == ' ' || c == '+' || c == '-' || c == '(' || c == ')':
		default:
			return false
		}
	}
	return digits >= 7
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

	if req.Phone == "" {
		fields["phone"] = "Phone number is required"
	} else if len([]rune(req.Phone)) > 30 {
		fields["phone"] = "Phone number must not exceed 30 characters"
	} else if !validPhone(req.Phone) {
		fields["phone"] = "Enter a valid phone number"
	}

	if req.Email == "" {
		fields["email"] = "Email is required"
	} else {
		address, err := mail.ParseAddress(req.Email)
		if err != nil || address.Address != req.Email || len(req.Email) > 254 {
			fields["email"] = "Enter a valid email address"
		}
	}

	// The 8-character minimum is provisional until the team confirms the password rule.
	// bcrypt only uses the first 72 bytes of a password.
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

func validateNewPassword(password, confirm string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	if len([]rune(password)) < 8 {
		return ErrWeakPassword
	}
	if len(password) > 72 {
		return ErrPasswordTooLong
	}
	if confirm == "" {
		return ErrConfirmRequired
	}
	if password != confirm {
		return ErrPasswordMismatch
	}
	return nil
}

func (s *userService) Register(ctx context.Context, req RegisterRequest) (*models.User, error) {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if fields := validateRegistration(req); len(fields) > 0 {
		return nil, &RegistrationValidationError{Fields: fields}
	}

	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hash,
		Role:      models.RoleCustomer,
		Status:    models.UserStatusActive,
	}

	// The unique index still catches a concurrent registration that passed the check above.
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if !utils.CheckPassword(user.Password, req.Password) {
		return nil, ErrInvalidCredentials
	}
	if user.Status != models.UserStatusActive {
		return nil, ErrInactiveUser
	}

	token, err := s.tokens.Generate(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: token, User: user.Profile()}, nil
}

func (s *userService) GetActiveUser(ctx context.Context, id uint) (models.User, error) {
	if id == 0 {
		return models.User{}, ErrUserNotFound
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if user.Status != models.UserStatusActive {
		return models.User{}, ErrInactiveUser
	}

	return user, nil
}

func (s *userService) Logout(ctx context.Context, claims *auth.Claims) error {
	return s.tokens.Revoke(ctx, claims)
}

func (s *userService) ChangePassword(ctx context.Context, claims *auth.Claims, req ChangePasswordRequest) error {
	if claims == nil || claims.UserID == 0 {
		return ErrUserNotFound
	}

	if err := validateNewPassword(req.NewPassword, req.ConfirmPassword); err != nil {
		return err
	}
	if req.NewPassword == req.CurrentPassword {
		return ErrSamePassword
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	if !utils.CheckPassword(user.Password, req.CurrentPassword) {
		return ErrIncorrectPassword
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, hash); err != nil {
		return err
	}

	// End this session so the client must log in again with the new password.
	return s.tokens.Revoke(ctx, claims)
}



func (s *userService) UpdateUserProfile(ctx context.Context, userID uint, input models.ProfileUpdate) error {
    user, err := s.userRepo.FindByID(ctx, userID) 
    if err != nil {
        return err
    }

    if err := user.UpdateProfile(input); err != nil {
        return err
    }
    return s.userRepo.Update(ctx, &user)
}

