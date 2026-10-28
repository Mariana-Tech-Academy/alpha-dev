// EPIC 6 - CHK-03: CART ITEM MODEL - FILE 5

// PURPOSE:
// Defines the structure of shopping cart items stored in PostgreSQL.

// IMPLEMENTATION:
// 1. Uses gorm.Model for standard database fields.
// 2. Links each cart item to a shopping cart using CartID.
// 3. Links each cart item to a menu item using MenuItemID.
// 4. Stores the menu item details through the MenuItem relationship.
// 5. Ensures that the quantity is greater than zero.
// 6. Uses a partial unique index to prevent duplicate active
//    cart/menu item combinations.

// HOW IT WORKS:
// Each cart item belongs to a shopping cart and references a menu item.
// During checkout, the repository retrieves these items to calculate
// the order total and create the order items.
// After successful checkout, the cart items are soft-deleted.

// I updated the cart item model to support checkout.
// It links shopping carts to menu items and includes validation
// to prevent invalid quantities and duplicate active cart items.
// The model also supports soft deletion when checkout succeeds.

package models

import "gorm.io/gorm"

type CartItem struct {
	gorm.Model
	CartID     uint     `gorm:"not null;uniqueIndex:idx_cart_menu_item,where:deleted_at IS NULL" json:"cart_id"`
	MenuItemID uint     `gorm:"not null;uniqueIndex:idx_cart_menu_item,where:deleted_at IS NULL" json:"menu_item_id"`
	MenuItem   MenuItem `json:"menu_item"`
	Quantity   int      `gorm:"not null;check:quantity > 0" json:"quantity"`
	Price      int64    `gorm:"not null;default:0" json:"price"`
}
