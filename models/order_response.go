package models

import "time"

// Manager order money fields are minor units, so 1250 means 12.50.
type ManagerOrderItemResponse struct {
	MenuItemID uint   `json:"menu_item_id"`
	Name       string `json:"name"`
	UnitPrice  int64  `json:"unit_price"`
	Quantity   int    `json:"quantity"`
	LineTotal  int64  `json:"line_total"`
}

type ManagerOrderResponse struct {
	ID                   uint                       `json:"id"`
	OrderNumber          string                     `json:"order_number"`
	Status               OrderStatus                `json:"status"`
	Subtotal             int64                      `json:"subtotal"`
	DeliveryFee          int64                      `json:"delivery_fee"`
	Total                int64                      `json:"total"`
	DeliveryName         string                     `json:"delivery_name"`
	DeliveryPhone        string                     `json:"delivery_phone"`
	DeliveryAddress      string                     `json:"delivery_address"`
	DeliveryCity         string                     `json:"delivery_city"`
	DeliveryState        string                     `json:"delivery_state"`
	DeliveryInstructions string                     `json:"delivery_instructions,omitempty"`
	Items                []ManagerOrderItemResponse `json:"items"`
	CreatedAt            time.Time                  `json:"created_at"`
}
