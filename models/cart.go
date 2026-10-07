package models

type Cart struct {
	Record
	UserID uint `gorm:"not null;uniqueIndex" json:"user_id"`
	User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	// NULL while empty. Services enforce the one-restaurant rule in a transaction.
	RestaurantID *uint       `gorm:"index" json:"restaurant_id"`
	Restaurant   *Restaurant `gorm:"foreignKey:RestaurantID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Items        []CartItem  `gorm:"foreignKey:CartID" json:"items"`
}

type CartItem struct {
	Record
	CartID     uint     `gorm:"not null;uniqueIndex:idx_cart_menu,priority:1" json:"cart_id"`
	Cart       Cart     `gorm:"foreignKey:CartID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	MenuItemID uint     `gorm:"not null;uniqueIndex:idx_cart_menu,priority:2;index" json:"menu_item_id"`
	MenuItem   MenuItem `gorm:"foreignKey:MenuItemID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"menu_item,omitempty"`
	Quantity   int      `gorm:"not null;check:chk_cart_quantity,quantity > 0" json:"quantity"`
}
