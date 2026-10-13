package models

import "gorm.io/gorm"

type CartItem struct {
	gorm.Model
	CartID     uint     `gorm:"not null;uniqueIndex:idx_cart_menu_item" json:"cart_id"`
	MenuItemID uint     `gorm:"not null;uniqueIndex:idx_cart_menu_item" json:"menu_item_id"`
	MenuItem   MenuItem `json:"menu_item"`
	Quantity   int      `gorm:"not null;check:quantity > 0" json:"quantity"`
}
