package models

// DeliverySnapshot is embedded into orders, independent of a customer's profile.
type DeliverySnapshot struct {
	RecipientName string `gorm:"size:200;not null" json:"recipient_name"`
	PhoneNumber   string `gorm:"size:30;not null" json:"phone_number"`
	Address       string `gorm:"type:text;not null" json:"address"`
	City          string `gorm:"size:100;not null" json:"city"`
	State         string `gorm:"size:100;not null" json:"state"`
	Instructions  string `gorm:"type:text" json:"instructions"`
}

type Order struct {
	Record
	OrderNumber      string           `gorm:"size:40;not null;uniqueIndex" json:"order_number"`
	CustomerID       uint             `gorm:"not null;index" json:"customer_id"`
	Customer         User             `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	RestaurantID     uint             `gorm:"not null;index" json:"restaurant_id"`
	Restaurant       Restaurant       `gorm:"foreignKey:RestaurantID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	SubtotalMinor    Money            `gorm:"type:bigint;not null;check:chk_order_subtotal,subtotal_minor >= 0" json:"subtotal_minor"`
	DeliveryFeeMinor Money            `gorm:"type:bigint;not null;check:chk_order_delivery_fee,delivery_fee_minor >= 0" json:"delivery_fee_minor"`
	TotalMinor       Money            `gorm:"type:bigint;not null;check:chk_order_total,total_minor = subtotal_minor + delivery_fee_minor" json:"total_minor"`
	Currency         string           `gorm:"size:3;not null;default:USD" json:"currency"`
	Delivery         DeliverySnapshot `gorm:"embedded;embeddedPrefix:delivery_" json:"delivery"`
	Status           OrderStatus      `gorm:"size:30;not null;default:pending;index;check:chk_order_status,status IN ('pending','confirmed','preparing','ready_for_delivery','out_for_delivery','delivered','cancelled')" json:"status"`
	Items            []OrderItem      `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	Record
	OrderID        uint     `gorm:"not null;index" json:"order_id"`
	Order          Order    `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	MenuItemID     uint     `gorm:"not null;index" json:"menu_item_id"`
	MenuItem       MenuItem `gorm:"foreignKey:MenuItemID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	ItemName       string   `gorm:"size:150;not null" json:"item_name"`
	UnitPriceMinor Money    `gorm:"type:bigint;not null;check:chk_order_item_price,unit_price_minor >= 0" json:"unit_price_minor"`
	Quantity       int      `gorm:"not null;check:chk_order_item_quantity,quantity > 0" json:"quantity"`
}
