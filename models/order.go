package models

import "gorm.io/gorm"

type OrderStatus string

const (
	OrderStatusPending          OrderStatus = "pending"
	OrderStatusConfirmed        OrderStatus = "confirmed"
	OrderStatusPreparing        OrderStatus = "preparing"
	OrderStatusReadyForDelivery OrderStatus = "ready_for_delivery"
	OrderStatusOutForDelivery   OrderStatus = "out_for_delivery"
	OrderStatusDelivered        OrderStatus = "delivered"
	OrderStatusCancelled        OrderStatus = "cancelled"
)

// Delivery fields are a snapshot taken at checkout; they must not reference the user's profile.
type Order struct {
	gorm.Model
	OrderNumber  string      `gorm:"uniqueIndex;not null" json:"order_number"` // ORD-YYYYMMDD-NNN
	CustomerID   uint        `gorm:"not null;index" json:"customer_id"`
	Customer     User        `json:"customer,omitempty"`
	RestaurantID uint        `gorm:"not null;index" json:"restaurant_id"`
	Restaurant   Restaurant  `json:"restaurant,omitempty"`
	Status       OrderStatus `gorm:"type:varchar(30);not null;default:pending;index" json:"status"`

	Subtotal    int64 `gorm:"not null" json:"subtotal"`
	DeliveryFee int64 `gorm:"not null" json:"delivery_fee"`
	Total       int64 `gorm:"not null" json:"total"`

	DeliveryName         string `gorm:"not null" json:"delivery_name"`
	DeliveryPhone        string `gorm:"not null" json:"delivery_phone"`
	DeliveryAddress      string `gorm:"not null" json:"delivery_address"`
	DeliveryCity         string `gorm:"not null" json:"delivery_city"`
	DeliveryState        string `gorm:"not null" json:"delivery_state"`
	DeliveryInstructions string `json:"delivery_instructions"`

	Items []OrderItem `json:"items"`
}
