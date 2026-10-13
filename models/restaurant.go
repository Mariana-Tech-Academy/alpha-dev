package models

import "gorm.io/gorm"

type RestaurantStatus string

const (
	RestaurantStatusPending   RestaurantStatus = "pending_approval"
	RestaurantStatusActive    RestaurantStatus = "active"
	RestaurantStatusInactive  RestaurantStatus = "inactive"
	RestaurantStatusSuspended RestaurantStatus = "suspended"
)

type Restaurant struct {
	gorm.Model
	ManagerID   uint             `gorm:"not null;index" json:"manager_id"`
	Manager     User             `json:"manager,omitempty"`
	Name        string           `gorm:"not null;index" json:"name"`
	Description string           `json:"description"`
	Address     string           `gorm:"not null" json:"address"`
	City        string           `gorm:"index" json:"city"`
	State       string           `gorm:"index" json:"state"`
	Phone       string           `json:"phone"`
	Email       string           `json:"email"`
	ImageURL    string           `json:"image_url"`
	OpeningTime string           `gorm:"type:varchar(5)" json:"opening_time"` // "HH:MM"
	ClosingTime string           `gorm:"type:varchar(5)" json:"closing_time"` // "HH:MM"
	Status      RestaurantStatus `gorm:"type:varchar(20);not null;default:pending_approval;index" json:"status"`

	MenuItems []MenuItem `json:"menu_items,omitempty"`
	Orders    []Order    `json:"orders,omitempty"`
}
