package models

import "gorm.io/gorm"

type MenuItem struct {
	Record
	RestaurantID uint       `gorm:"not null;index" json:"restaurant_id"`
	Restaurant   Restaurant `gorm:"foreignKey:RestaurantID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	CategoryID   uint       `gorm:"not null;index" json:"category_id"`
	Category     Category   `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"category,omitempty"`
	Name         string     `gorm:"size:150;not null;index" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	PriceMinor   Money      `gorm:"type:bigint;not null;check:chk_menu_price,price_minor >= 0" json:"price_minor"`
	ImageURL     string     `gorm:"type:text" json:"image_url"`
	IsAvailable  bool       `gorm:"not null;index" json:"is_available"`
	// Soft deletion preserves historical order references.
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
