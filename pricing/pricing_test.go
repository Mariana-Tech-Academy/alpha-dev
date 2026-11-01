package pricing

import "testing"

func TestNewOrderTotals(t *testing.T) {
	for _, tc := range []struct {
		name       string
		subtotal   int64
		items      bool
		fee, total int64
	}{
		{"empty", 0, false, 0, 0}, {"empty ignores stale subtotal", 2500, false, 0, 0},
		{"one line", 1250, true, 299, 1549}, {"multiple lines and quantities", 1250*2 + 450*3, true, 299, 4149},
		{"free food still has delivery", 0, true, 299, 299},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Calculate(tc.subtotal, tc.items)
			if got.DeliveryFee != tc.fee || got.Total != tc.total {
				t.Fatalf("totals %+v", got)
			}
		})
	}
}
