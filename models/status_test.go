package models

import "testing"

func TestOrderStatusTransitions(t *testing.T) {
	statuses := []OrderStatus{
		OrderPending, OrderConfirmed, OrderPreparing, OrderReadyForDelivery,
		OrderOutForDelivery, OrderDelivered, OrderCancelled, "unknown",
	}
	allowed := map[[2]OrderStatus]bool{
		{OrderPending, OrderConfirmed}:               true,
		{OrderConfirmed, OrderPreparing}:             true,
		{OrderPreparing, OrderReadyForDelivery}:      true,
		{OrderReadyForDelivery, OrderOutForDelivery}: true,
		{OrderOutForDelivery, OrderDelivered}:        true,
		{OrderPending, OrderCancelled}:               true,
		{OrderConfirmed, OrderCancelled}:             true,
	}
	for _, from := range statuses {
		for _, to := range statuses {
			if got, want := from.CanTransitionTo(to), allowed[[2]OrderStatus{from, to}]; got != want {
				t.Errorf("%s -> %s: got %v, want %v", from, to, got, want)
			}
		}
	}
}
