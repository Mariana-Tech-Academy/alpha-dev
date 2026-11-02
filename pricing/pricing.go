// Package pricing is the backend policy for new-order pricing, in integer pence.
package pricing

const deliveryFeePence int64 = 299

type Totals struct {
	Subtotal    int64
	DeliveryFee int64
	Total       int64
}

// Calculate charges once per nonempty basket/order, even when its food is free.
// Historical orders must use their saved totals rather than this current policy.
func Calculate(subtotal int64, hasItems bool) Totals {
	if !hasItems {
		return Totals{}
	}
	return Totals{Subtotal: subtotal, DeliveryFee: deliveryFeePence, Total: subtotal + deliveryFeePence}
}
