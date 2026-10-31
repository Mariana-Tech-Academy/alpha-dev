package models

import "gorm.io/gorm"

type CartItem struct {
	gorm.Model  // // adds ID, CreatedAt, UpdatedAt, DeletedAt
	CartID     uint     `gorm:"not null;uniqueIndex:idx_cart_menu_item" json:"cart_id"`
	MenuItemID uint     `gorm:"not null;uniqueIndex:idx_cart_menu_item" json:"menu_item_id"`
	MenuItem   MenuItem `json:"menu_item"`  // the full menu item, with all its own fields
	Quantity   int      `gorm:"not null;check:quantity > 0" json:"quantity"`
	Price      int64    `gorm:"not null;default:0" json:"price"`
}
