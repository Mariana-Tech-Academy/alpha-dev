package models

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleManager  Role = "manager"
	RoleAdmin    Role = "admin"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	gorm.Model
	FirstName string     `gorm:"not null" json:"first_name"`
	LastName  string     `gorm:"not null" json:"last_name"`
	Email     string     `gorm:"uniqueIndex;not null" json:"email"`
	Phone     string     `gorm:"not null" json:"phone"`
	Password  string     `gorm:"not null" json:"-"`
	Role      Role       `gorm:"type:varchar(20);not null;default:customer;index" json:"role"`
	Status    UserStatus `gorm:"type:varchar(20);not null;default:active;index" json:"status"`

	Cart   *Cart   `json:"cart,omitempty"`
	Orders []Order `gorm:"foreignKey:CustomerID" json:"orders,omitempty"`
}

// UserProfile is the public view of a user returned by the API.
type UserProfile struct {
	ID        uint       `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Role      Role       `json:"role"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

func (u *User) Profile() UserProfile {
	return UserProfile{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}
