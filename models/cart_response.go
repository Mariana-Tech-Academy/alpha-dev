package models

// Prices are stored in minor units: 1250 means 12.50.
type CartItemResponse struct {
	ID           uint   `json:"id"`
	MenuItemID   uint   `json:"menu_item_id"`
	Name         string `json:"name"`
	Price        int64  `json:"price"`
	LockedPrice  int64  `json:"locked_price"`
	Quantity     int    `json:"quantity"`
	LineTotal    int64  `json:"line_total"`
	IsAvailable  bool   `json:"is_available"`
	PriceChanged bool   `json:"price_changed"`
	Notice       string `json:"notice,omitempty"`
}

type CartViewResponse struct {
	CartID       uint               `json:"cart_id"`
	RestaurantID *uint              `json:"restaurant_id,omitempty"`
	Items        []CartItemResponse `json:"items"`
	Subtotal     int64              `json:"subtotal"`
}