package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name      string     `gorm:"uniqueIndex;not null" json:"name"`
	MenuItems []MenuItem `json:"menu_items,omitempty"`
}

// CategorySummary is a category name returned by the public list.
type CategorySummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// CategoryMenuItem is a public menu row shown on a category page.
type CategoryMenuItem struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Price          int64  `json:"price"`
	ImageURL       string `json:"image_url,omitempty"`
	RestaurantID   uint   `json:"restaurant_id"`
	RestaurantName string `json:"restaurant_name"`
}

// CategoryDetails is one category and the menu items customers can order.
type CategoryDetails struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	MenuItems []CategoryMenuItem `json:"menu_items"`
}
