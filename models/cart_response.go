package models

type CartItemResponse struct {
	MenuItemID   uint    `json:"menu_item_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`        // Current live price of the menu item
	LockedPrice  float64 `json:"locked_price"` // Price when added to cart
	Quantity     int     `json:"quantity"`
	LineTotal    float64 `json:"line_total"`
	IsAvailable  bool    `json:"is_available"`
	PriceChanged bool    `json:"price_changed"`
	Notice       string  `json:"notice,omitempty"`
}

type CartViewResponse struct {
	CartID       uint               `json:"cart_id"`
	RestaurantID *uint              `json:"restaurant_id,omitempty"`
	Items        []CartItemResponse `json:"items"`
	Subtotal     float64            `json:"subtotal"`
}