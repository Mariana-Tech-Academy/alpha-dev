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

// orderFlow maps each status to the only status allowed to follow it.
// Delivered and cancelled have no entry, which makes them final.
var orderFlow = map[OrderStatus]OrderStatus{
	OrderStatusPending:          OrderStatusConfirmed,
	OrderStatusConfirmed:        OrderStatusPreparing,
	OrderStatusPreparing:        OrderStatusReadyForDelivery,
	OrderStatusReadyForDelivery: OrderStatusOutForDelivery,
	OrderStatusOutForDelivery:   OrderStatusDelivered,
}

//func (s OrderStatus) IsValid() is a method on a type. OrderStatus is a custom type built on string (type OrderStatus string), and Go lets you attach behaviour to your own types. You can then write status.IsValid().

//switch s { case A, B, C: return true } lists all seven statuses on one case line. Anything else falls out of the switch to return false, which catches junk like "teleported".
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusPreparing,
		OrderStatusReadyForDelivery, OrderStatusOutForDelivery,
		OrderStatusDelivered, OrderStatusCancelled:
		return true
	}
	return false
}

//named return values
func (s OrderStatus) Next() (next OrderStatus, ok bool) {
	next, ok = orderFlow[s]
	return next, ok
}

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
