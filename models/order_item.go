package models

import "gorm.io/gorm"

// Name and UnitPrice are snapshots taken at checkout so later menu edits never change past orders.
type OrderItem struct {
	gorm.Model
	OrderID    uint     `gorm:"not null;index" json:"order_id"`
	MenuItemID uint     `gorm:"not null;index" json:"menu_item_id"`
	MenuItem   MenuItem `json:"-"`
	Name       string   `gorm:"not null" json:"name"`
	UnitPrice  int64    `gorm:"not null" json:"unit_price"`
	Quantity   int      `gorm:"not null;check:quantity > 0" json:"quantity"`
	LineTotal  int64    `gorm:"not null" json:"line_total"`
}
