package models

// ManagerOrderItemResponse is the DTO for items within a manager's order view, with prices in decimal pounds.
type ManagerOrderItemResponse struct {
	MenuItemID uint    `json:"menu_item_id"`
	Name       string  `json:"name"`
	UnitPrice  float64 `json:"unit_price"`
	Quantity   int     `json:"quantity"`
	LineTotal  float64 `json:"line_total"`
}

// ManagerOrderResponse is the DTO returned to restaurant managers for order management.
type ManagerOrderResponse struct {
	ID                   uint                       `json:"id"`
	OrderNumber          string                     `json:"order_number"`
	Status               OrderStatus                `json:"status"`
	Subtotal             float64                    `json:"subtotal"`
	DeliveryFee          float64                    `json:"delivery_fee"`
	Total                float64                    `json:"total"`
	DeliveryName         string                     `json:"delivery_name"`
	DeliveryPhone        string                     `json:"delivery_phone"`
	DeliveryAddress      string                     `json:"delivery_address"`
	DeliveryInstructions string                     `json:"delivery_instructions"`
	Items                []ManagerOrderItemResponse `json:"items"`
	CreatedAt            string                     `json:"created_at"`
}
