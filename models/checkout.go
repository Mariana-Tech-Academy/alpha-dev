package models

// CheckoutItem is one line shown before the customer places the order.
// Money fields are minor units, so 1250 means 12.50.
type CheckoutItem struct {
	MenuItemID uint   `json:"menu_item_id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
	LineTotal  int64  `json:"line_total"`
}

type CheckoutRestaurant struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// CheckoutSummary is the review the customer sees before placing the order.
type CheckoutSummary struct {
	Restaurant  CheckoutRestaurant `json:"restaurant"`
	Items       []CheckoutItem     `json:"items"`
	Subtotal    int64              `json:"subtotal"`
	DeliveryFee int64              `json:"delivery_fee"`
	Total       int64              `json:"total"`
}
