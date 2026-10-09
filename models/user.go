package models

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
