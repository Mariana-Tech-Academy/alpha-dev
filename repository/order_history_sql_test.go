package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
	"testing"
)

func TestCustomerHistorySQLScopesSnapshotsWithoutRestaurantVisibilityFilter(t *testing.T) {
	c := &cartSQL{t: t, steps: []cartSQLStep{
		{contains: `count(*) FROM "orders"`, columns: []string{"count"}, rows: [][]driver.Value{{int64(1)}}},
		{contains: `FROM "orders"`, columns: []string{"id", "customer_id", "restaurant_id", "subtotal", "delivery_fee", "total", "delivery_address"}, rows: [][]driver.Value{{int64(22), int64(7), int64(1), int64(2500), int64(0), int64(2500), "Saved street"}}},
		{contains: `FROM "order_items"`, columns: []string{"id", "order_id", "name", "quantity", "unit_price", "line_total"}, rows: [][]driver.Value{{int64(23), int64(22), "Snapshot burger", int64(2), int64(1250), int64(2500)}}},
		{contains: `FROM "restaurants"`, columns: []string{"id", "name", "image_url"}, rows: [][]driver.Value{{int64(1), "FoodHub Test Restaurant", ""}}},
	}}
	conn := sql.OpenDB(c)
	defer conn.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	orders, total, err := NewOrderRepository(db).FindCustomerOrders(context.Background(), 7, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.steps) != 0 || total != 1 || len(orders) != 1 || len(orders[0].Items) != 1 {
		t.Fatalf("missing snapshots: %+v %v", orders, err)
	}
	if orders[0].Items[0].Name != "Snapshot burger" || orders[0].DeliveryAddress != "Saved street" || orders[0].Restaurant.Name != "FoodHub Test Restaurant" {
		t.Fatal("wrong saved data")
	}
	for i := 0; i < 2; i++ {
		if !strings.Contains(c.statements[i], "customer_id = $1") || c.arguments[i][0].Value != int64(7) {
			t.Fatalf("history query lacks ownership scope: %s", c.statements[i])
		}
	}
	for _, query := range c.statements {
		if strings.Contains(query, "image_url =") || strings.Contains(query, "status =") || strings.Contains(query, `FROM "menu_items"`) {
			t.Fatalf("history relies on current public/menu data: %s", query)
		}
	}
	if c.arguments[2][0].Value != int64(22) {
		t.Fatal("items are not scoped to the selected order")
	}
}
