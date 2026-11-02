package repository

import (
	"alpha-dev/models"
	"context"
	"database/sql"
	"database/sql/driver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
	"testing"
)

func TestOrderStatusSQLPersistsConditionalUpdate(t *testing.T) {
	for _, step := range []struct{ from, to models.OrderStatus }{
		{models.OrderStatusPending, models.OrderStatusPreparing},
		{models.OrderStatusPreparing, models.OrderStatusOutForDelivery},
		{models.OrderStatusOutForDelivery, models.OrderStatusDelivered},
	} {
		for _, affected := range []int64{1, 0} {
			t.Run(string(step.from)+"-"+string(step.to)+string(rune('0'+affected)), func(t *testing.T) {
				c := &cartSQL{t: t, steps: []cartSQLStep{{contains: `UPDATE "orders" SET "status"`, affected: affected}}}
				if affected == 1 {
					c.steps = append(c.steps, cartSQLStep{contains: `FROM "orders"`, columns: []string{"id", "status"}, rows: [][]driver.Value{{int64(9), string(step.to)}}})
				}
				conn := sql.OpenDB(c)
				defer conn.Close()
				db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
				if err != nil {
					t.Fatal(err)
				}
				repo := NewOrderRepository(db)
				changed, err := repo.UpdateStatus(context.Background(), 9, step.from, step.to)
				if err != nil || changed != (affected == 1) {
					t.Fatalf("changed %v err %v", changed, err)
				}
				if !strings.Contains(c.statements[0], "id = $3 AND status = $4") {
					t.Fatal("status update lacks compare-and-set scope")
				}
				for _, field := range []string{"delivery_", "subtotal", "total", "order_items", "cart_items"} {
					if strings.Contains(c.statements[0], field) {
						t.Fatal("status update touched unrelated fields")
					}
				}
				args := c.arguments[0]
				if len(args) != 4 || args[0].Value != string(step.to) || args[2].Value != int64(9) || args[3].Value != string(step.from) {
					t.Fatalf("missing conditional update: %v", args)
				}
				if affected == 1 {
					saved, err := repo.FindByID(context.Background(), 9)
					if err != nil || saved.Status != step.to {
						t.Fatal("saved status missing")
					}
				}
				if len(c.steps) != 0 || !c.committed {
					t.Fatal("write did not commit")
				}
			})
		}
	}
}
