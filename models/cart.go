package models

import "gorm.io/gorm"

type Cart struct {
	gorm.Model
	UserID       uint        `gorm:"uniqueIndex;not null" json:"user_id"`
	RestaurantID *uint       `gorm:"index" json:"restaurant_id"`
	Restaurant   *Restaurant `json:"restaurant,omitempty"`
	Items        []CartItem  `json:"items"`
}
