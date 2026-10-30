package models

import "testing"

func TestCanBeCancelled(t *testing.T) {
	canCancel := []OrderStatus{OrderStatusPending, OrderStatusConfirmed}
	cannotCancel := []OrderStatus{
		OrderStatusPreparing,
		OrderStatusReadyForDelivery,
		OrderStatusOutForDelivery,
		OrderStatusDelivered,
		OrderStatusCancelled,
	}

	for _, status := range canCancel {
		if !status.CanBeCancelled() {
			t.Errorf("%s should be cancellable", status)
		}
	}
	for _, status := range cannotCancel {
		if status.CanBeCancelled() {
			t.Errorf("%s should not be cancellable", status)
		}
	}
}
