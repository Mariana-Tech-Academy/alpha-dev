package models

// CheckoutItem is one line of the order review shown before payment.
// All money values are in minor units (kobo), like the rest of the API.


//models/checkout.go defines the shape of the JSON the endpoint sends back.
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

// CheckoutSummary is what the customer reviews before placing the order.
type CheckoutSummary struct {
	Restaurant  CheckoutRestaurant `json:"restaurant"`
	Items       []CheckoutItem     `json:"items"`
	Subtotal    int64              `json:"subtotal"`
	DeliveryFee int64              `json:"delivery_fee"`
	Total       int64              `json:"total"`
}
 
//why not just return existing cart model?

/*
1. Cart stores no subtotals, no delivery fee or total, those numbers are calculated not stored in a database
2. The cart stores Price, but the summary talks about unit_price, which matches your OrderItem model.
4.
*/