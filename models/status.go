package models

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleManager  UserRole = "manager"
	RoleAdmin    UserRole = "admin"
)

type UserStatus string

const (
	UserActive    UserStatus = "active"
	UserSuspended UserStatus = "suspended"
)

type RestaurantStatus string

const (
	RestaurantPendingApproval RestaurantStatus = "pending_approval"
	RestaurantActive          RestaurantStatus = "active"
	RestaurantSuspended       RestaurantStatus = "suspended"
)

type OrderStatus string

const (
	OrderPending          OrderStatus = "pending"
	OrderConfirmed        OrderStatus = "confirmed"
	OrderPreparing        OrderStatus = "preparing"
	OrderReadyForDelivery OrderStatus = "ready_for_delivery"
	OrderOutForDelivery   OrderStatus = "out_for_delivery"
	OrderDelivered        OrderStatus = "delivered"
	OrderCancelled        OrderStatus = "cancelled"
)

// CanTransitionTo checks the flow only; services must also authorize the actor.
// Proposed MVP policy: cancellation is allowed while pending or confirmed.
func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	if next == OrderCancelled {
		return s == OrderPending || s == OrderConfirmed
	}
	switch s {
	case OrderPending:
		return next == OrderConfirmed
	case OrderConfirmed:
		return next == OrderPreparing
	case OrderPreparing:
		return next == OrderReadyForDelivery
	case OrderReadyForDelivery:
		return next == OrderOutForDelivery
	case OrderOutForDelivery:
		return next == OrderDelivered
	default:
		return false
	}
}
