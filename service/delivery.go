// EPIC 6 - CHK-02: DELIVERY DETAILS
// Trims and validates the delivery details the customer enters at checkout.
// The checkout form pre-fills name and phone from GET /api/auth/me.

package service

import (
	"strconv"
	"strings"

	"alpha-dev/models"
)

// DeliveryValidationError carries one message per invalid delivery field.
type DeliveryValidationError struct {
	Fields map[string]string
}

func (err *DeliveryValidationError) Error() string {
	return "invalid delivery fields"
}

func trimDelivery(order *models.Order) {
	order.DeliveryName = strings.TrimSpace(order.DeliveryName)
	order.DeliveryPhone = strings.TrimSpace(order.DeliveryPhone)
	order.DeliveryAddress = strings.TrimSpace(order.DeliveryAddress)
	order.DeliveryCity = strings.TrimSpace(order.DeliveryCity)
	order.DeliveryState = strings.TrimSpace(order.DeliveryState)
	order.DeliveryInstructions = strings.TrimSpace(order.DeliveryInstructions)
}

func validateDelivery(order *models.Order) map[string]string {
	fields := make(map[string]string)

	requireText(fields, "delivery_name", "Recipient name", order.DeliveryName, 100)
	requireText(fields, "delivery_address", "Address", order.DeliveryAddress, 255)
	requireText(fields, "delivery_city", "City", order.DeliveryCity, 100)
	requireText(fields, "delivery_state", "State", order.DeliveryState, 100)

	if order.DeliveryPhone == "" {
		fields["delivery_phone"] = "Phone number is required"
	} else if len([]rune(order.DeliveryPhone)) > 30 {
		fields["delivery_phone"] = "Phone number must not exceed 30 characters"
	} else if !validPhone(order.DeliveryPhone) {
		fields["delivery_phone"] = "Enter a valid phone number"
	}

	// Instructions are optional.
	if len([]rune(order.DeliveryInstructions)) > 500 {
		fields["delivery_instructions"] = "Delivery instructions must not exceed 500 characters"
	}

	return fields
}

func requireText(fields map[string]string, key, label, value string, max int) {
	if value == "" {
		fields[key] = label + " is required"
	} else if len([]rune(value)) > max {
		fields[key] = label + " must not exceed " + strconv.Itoa(max) + " characters"
	}
}
