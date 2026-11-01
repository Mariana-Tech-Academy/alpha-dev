package repository

// A scripted SQL connection exercises the real GORM repository without a database.
import (
	"alpha-dev/models"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"strings"
	"testing"
	"time"
)

type cartSQLStep struct {
	contains string
	columns  []string
	rows     [][]driver.Value
	affected int64
	err      error
}
type cartSQL struct {
	t                 *testing.T
	steps             []cartSQLStep
	statements        []string
	arguments         [][]driver.NamedValue
	committed, rolled bool
}

func (c *cartSQL) Connect(context.Context) (driver.Conn, error) { return c, nil }
func (c *cartSQL) Driver() driver.Driver                        { return c }
func (c *cartSQL) Open(string) (driver.Conn, error)             { return c, nil }
func (c *cartSQL) Close() error                                 { return nil }
func (c *cartSQL) Prepare(string) (driver.Stmt, error)          { return nil, errors.New("unexpected prepare") }
func (c *cartSQL) Begin() (driver.Tx, error)                    { return c, nil }
func (c *cartSQL) Commit() error                                { c.committed = true; return nil }
func (c *cartSQL) Rollback() error                              { c.rolled = true; return nil }
func (c *cartSQL) step(query string, args []driver.NamedValue) cartSQLStep {
	c.t.Helper()
	c.statements = append(c.statements, query)
	c.arguments = append(c.arguments, args)
	if len(c.steps) == 0 {
		c.t.Errorf("unexpected SQL: %s", query)
		return cartSQLStep{err: errors.New("unexpected SQL")}
	}
	s := c.steps[0]
	c.steps = c.steps[1:]
	if !strings.Contains(query, s.contains) {
		c.t.Errorf("SQL %s lacks %s", query, s.contains)
	}
	if strings.Contains(query, `FROM "carts"`) {
		if !strings.Contains(query, "user_id = $1") || !strings.Contains(query, "FOR UPDATE") || args[0].Value != int64(7) {
			c.t.Errorf("missing customer scope/lock: %s %v", query, args)
		}
	}
	if strings.Contains(query, `UPDATE "cart_items"`) {
		if !strings.Contains(query, "cart_id = $2") || args[1].Value != int64(4) || !strings.Contains(query, `"deleted_at" IS NULL`) {
			c.t.Errorf("unsafe removal: %s %v", query, args)
		}
		if strings.Contains(s.contains, "id = $3") && args[2].Value != int64(11) {
			c.t.Error("wrong line ID")
		}
	}
	// GORM may save loaded cart associations during checkout's cart reset.
	// Its conflict update must not restore a line's soft-delete marker.
	if strings.Contains(query, `INSERT INTO "cart_items"`) {
		parts := strings.Split(query, "DO UPDATE SET")
		if len(parts) != 2 || strings.Contains(parts[1], `"deleted_at"=`) {
			c.t.Errorf("checkout could revive cleared lines: %s", query)
		}
	}
	return s
}
func (c *cartSQL) QueryContext(_ context.Context, q string, a []driver.NamedValue) (driver.Rows, error) {
	s := c.step(q, a)
	return &cartRows{s: s}, s.err
}
func (c *cartSQL) ExecContext(_ context.Context, q string, a []driver.NamedValue) (driver.Result, error) {
	s := c.step(q, a)
	return driver.RowsAffected(s.affected), s.err
}

type cartRows struct {
	s     cartSQLStep
	index int
}

func (r *cartRows) Columns() []string { return r.s.columns }
func (r *cartRows) Close() error      { return nil }
func (r *cartRows) Next(dest []driver.Value) error {
	if r.index == len(r.s.rows) {
		return io.EOF
	}
	copy(dest, r.s.rows[r.index])
	r.index++
	return nil
}
func TestRemoveItemsSQL(t *testing.T) {
	for _, tc := range []struct {
		name        string
		item        *uint
		remaining   int64
		affected    int64
		missingCart bool
		wantError   bool
		fail        bool
	}{
		{name: "remove last line resets restaurant", item: new(uint(11)), affected: 1},
		{name: "remove one keeps restaurant", item: new(uint(11)), remaining: 1, affected: 1},
		{name: "clear resets restaurant", affected: 2},
		{name: "clear empty is safe"},
		{name: "foreign or missing item rolls back", item: new(uint(11)), wantError: true},
		{name: "clear missing cart is safe", missingCart: true},
		{name: "missing cart item returns not found", item: new(uint(11)), missingCart: true, wantError: true},
		{name: "reset failure rolls back removal", affected: 2, fail: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first := cartSQLStep{contains: `FROM "carts"`, columns: []string{"id", "user_id", "restaurant_id"}}
			if !tc.missingCart {
				first.rows = [][]driver.Value{{int64(4), int64(7), int64(3)}}
			}
			c := &cartSQL{t: t, steps: []cartSQLStep{first}}
			if !tc.missingCart {
				scope := "cart_id = $2"
				if tc.item != nil {
					scope += " AND id = $3"
				}
				c.steps = append(c.steps, cartSQLStep{contains: scope, affected: tc.affected})
				if tc.item == nil || tc.affected > 0 {
					c.steps = append(c.steps, cartSQLStep{contains: `count(*) FROM "cart_items"`, columns: []string{"count"}, rows: [][]driver.Value{{tc.remaining}}})
					if tc.remaining == 0 {
						last := cartSQLStep{contains: `UPDATE "carts" SET "restaurant_id"`}
						if tc.fail {
							last.err = errors.New("reset failed")
						}
						c.steps = append(c.steps, last)
					}
				}
			}
			dbConn := sql.OpenDB(c)
			defer dbConn.Close()
			db, err := gorm.Open(postgres.New(postgres.Config{Conn: dbConn}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			err = NewCartRepository(db).RemoveItems(context.Background(), 7, tc.item)
			if (err != nil) != tc.wantError {
				t.Fatalf("error %v", err)
			}
			if len(c.steps) != 0 {
				t.Fatalf("%d unexecuted statements", len(c.steps))
			}
			if tc.wantError {
				if !c.rolled || c.committed {
					t.Fatal("failed operation was not rolled back")
				}
			} else if !c.committed {
				t.Fatal("operation did not commit")
			}
			for _, q := range c.statements {
				if strings.Contains(q, `"orders"`) || strings.Contains(q, `"order_items"`) {
					t.Fatal("removal touched orders")
				}
			}
		})
	}
}

func TestCheckoutSQLClearsBasketAndKeepsPendingOrder(t *testing.T) {
	c := &cartSQL{t: t, steps: []cartSQLStep{
		{contains: `FROM "carts"`, columns: []string{"id", "user_id", "restaurant_id"}, rows: [][]driver.Value{{int64(4), int64(7), int64(3)}}},
		{contains: `FROM "cart_items"`, columns: []string{"id", "cart_id", "menu_item_id", "quantity", "price"}, rows: [][]driver.Value{{int64(11), int64(4), int64(9), int64(2), int64(1250)}, {int64(12), int64(4), int64(10), int64(3), int64(450)}}},
		{contains: `FROM "menu_items"`, columns: []string{"id", "restaurant_id", "name", "available", "price"}, rows: [][]driver.Value{{int64(9), int64(3), "Burger", true, int64(1300)}, {int64(10), int64(3), "Side", true, int64(500)}}},
		{contains: `FROM "restaurants"`, columns: []string{"id", "status"}, rows: [][]driver.Value{{int64(3), "active"}}},
		{contains: `INSERT INTO order_daily_sequences`, columns: []string{"order_date", "last_number"}, rows: [][]driver.Value{{time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), int64(1)}}},
		{contains: `INSERT INTO "orders"`, columns: []string{"id"}, rows: [][]driver.Value{{int64(22)}}},
		{contains: `INSERT INTO "order_items"`, columns: []string{"id"}, rows: [][]driver.Value{{int64(23)}, {int64(24)}}},
		{contains: `UPDATE "cart_items"`, affected: 1},
		{contains: `UPDATE "carts" SET "restaurant_id"`, affected: 1},
		{contains: `INSERT INTO "menu_items"`, columns: []string{"id"}, rows: [][]driver.Value{{int64(9)}, {int64(10)}}},
		{contains: `INSERT INTO "cart_items"`, columns: []string{"id"}, rows: [][]driver.Value{{int64(11)}, {int64(12)}}},
	}}
	conn := sql.OpenDB(c)
	defer conn.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	placed, err := NewOrderRepository(db).PlaceOrder(context.Background(), 7, &models.Order{Subtotal: 1, DeliveryFee: 1, Total: 1})
	if err != nil {
		t.Fatal(err)
	}
	for i, query := range c.statements {
		if strings.HasPrefix(query, `INSERT INTO "orders"`) {
			columns := strings.Split(strings.SplitN(strings.SplitN(query, "(", 2)[1], ")", 2)[0], ",")
			expected := map[string]int64{"subtotal": 3850, "delivery_fee": 299, "total": 4149}
			for j, column := range columns {
				if value, ok := expected[strings.Trim(column, "\" ")]; ok && c.arguments[i][j].Value != value {
					t.Fatal("wrong persisted pricing")
				}
			}
		}
	}
	if !c.committed || c.rolled || len(c.steps) != 0 {
		t.Fatal("checkout did not atomically save and clear")
	}
	if placed.Status != models.OrderStatusPending || placed.ID != 22 || placed.Subtotal != 3850 || placed.DeliveryFee != 299 || placed.Total != 4149 || len(placed.Items) != 2 {
		t.Fatalf("incorrect placed order: %+v", placed)
	}
}
