package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name      string     `gorm:"uniqueIndex;not null" json:"name"`
	MenuItems []MenuItem `json:"menu_items,omitempty"`
}
