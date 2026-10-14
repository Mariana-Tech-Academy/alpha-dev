package models

import "gorm.io/gorm"

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


type ProfileUpdate struct {

	FirstName string ` json:"first_name"`
	LastName  string ` json:"last_name"`
	Email     string ` json:"email"`
	Phone string  ` json:"phone_number"`
}

func (u *User) UpdateProfile (db *gorm.DB, input ProfileUpdate) error {
	u.FirstName = input.FirstName
	u.LastName = input.LastName
	u.Phone = input.Phone

	if input.Email != "" && input.Email != u.Email {
		u.Email = input.Email
	}

	return db.Save (u).Error

}
