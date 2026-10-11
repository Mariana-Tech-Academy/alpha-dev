package models

import ("gorm.io/gorm")

type User struct {
	Record
	FirstName          string       `gorm:"size:100;not null" json:"first_name"`
	LastName           string       `gorm:"size:100;not null" json:"last_name"`
	Email              string       `gorm:"size:254;not null;uniqueIndex" json:"email"`
	PhoneNumber        string       `gorm:"size:30;not null" json:"phone_number"`
	PasswordHash       string       `gorm:"size:255;not null" json:"-"`
	Role               UserRole     `gorm:"size:20;not null;default:customer;index;check:chk_user_role,role IN ('customer','manager','admin')" json:"role"`
	Status             UserStatus   `gorm:"size:20;not null;default:active;index;check:chk_user_status,status IN ('active','suspended')" json:"status"`
	Cart               *Cart        `gorm:"foreignKey:UserID" json:"-"`
	Orders             []Order      `gorm:"foreignKey:CustomerID" json:"-"`
	ManagedRestaurants []Restaurant `gorm:"foreignKey:ManagerID" json:"-"`
}

type ProfileUpdate struct {

	FirstName string ` json:"first_name"`
	LastName  string ` json:"last_name"`
	Email     string ` json:"email"`
	PhoneNumber string  ` json:"phone_number"`
}

func (u *User) UpdateProfile (db *gorm.DB, input ProfileUpdate) error {
	u.FirstName = input.FirstName
	u.LastName = input.LastName
	u.PhoneNumber = input.PhoneNumber

	if input.Email != "" && input.Email != u.Email {
		u.Email = input.Email
	}

	return db.Save (u).Error

}