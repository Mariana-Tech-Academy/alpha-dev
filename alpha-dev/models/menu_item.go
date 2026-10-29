package models

import "gorm.io/gorm"

// Money fields are stored in minor units
type MenuItem struct {
	gorm.Model
	RestaurantID uint       `gorm:"not null;index" json:"restaurant_id"`
	Restaurant   Restaurant `json:"restaurant,omitempty"`
	CategoryID   uint       `gorm:"not null;index" json:"category_id"`
	Category     Category   `json:"category,omitempty"`
	Name         string     `gorm:"not null;index" json:"name"`
	Description  string     `json:"description"`
	Price        int64      `gorm:"not null;check:price >= 0" json:"price"`
	ImageURL     string     `json:"image_url"`
	Available    bool       `gorm:"not null;default:true;index" json:"available"`
}
