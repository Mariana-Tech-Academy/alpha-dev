package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"alpha-dev/models"
)

func TestCancelOrderAndSalesTotal(t *testing.T) {
	db := setupOrderTestDB(t)

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to start test transaction: %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	db = tx
	ctx := context.Background()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	customer := models.User{FirstName: "Cancel", LastName: "Customer", Email: "cancel.customer." + suffix + "@test.local", Phone: "07123456789", Password: "x", Role: models.RoleCustomer}
	manager := models.User{FirstName: "Cancel", LastName: "Manager", Email: "cancel.manager." + suffix + "@test.local", Phone: "07987654321", Password: "x", Role: models.RoleManager}
	for _, u := range []*models.User{&customer, &manager} {
		if err := db.Create(u).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	restaurant := models.Restaurant{ManagerID: manager.ID, Name: "Cancel Test " + suffix, Address: "1 Test Street", Status: models.RestaurantStatusActive}
	if err := db.Create(&restaurant).Error; err != nil {
		t.Fatalf("create restaurant: %v", err)
	}

	newOrder := func(n int, status models.OrderStatus, total int64) *models.Order {
		order := &models.Order{
			OrderNumber: fmt.Sprintf("ORD-TEST-%s-%d", suffix, n), CustomerID: customer.ID, RestaurantID: restaurant.ID,
			Status: status, Subtotal: total, Total: total,
			DeliveryName: "A", DeliveryPhone: "1", DeliveryAddress: "B", DeliveryCity: "C", DeliveryState: "D",
		}
		if err := db.Create(order).Error; err != nil {
			t.Fatalf("create order: %v", err)
		}
		return order
	}
	pending := newOrder(1, models.OrderStatusPending, 1000)
	preparing := newOrder(2, models.OrderStatusPreparing, 2000)
	newOrder(3, models.OrderStatusDelivered, 4000)

	repo := NewOrderRepository(db)

	salesTotal := func() int64 {
		var total int64
		if err := db.Model(&models.Order{}).Scopes(CountedAsSales).
			Where("restaurant_id = ?", restaurant.ID).
			Select("COALESCE(SUM(total), 0)").Scan(&total).Error; err != nil {
			t.Fatalf("sales total: %v", err)
		}
		return total
	}
	if got := salesTotal(); got != 7000 {
		t.Fatalf("sales before cancel = %d, want 7000", got)
	}

	found, err := repo.FindOrderByID(ctx, pending.ID)
	if err != nil || found.Restaurant.ManagerID != manager.ID {
		t.Fatalf("FindOrderByID = %+v, %v; want restaurant manager %d", found, err, manager.ID)
	}

	if err := repo.CancelOrder(ctx, pending.ID); err != nil {
		t.Fatalf("cancel pending order: %v", err)
	}
	reloaded, _ := repo.FindOrderByID(ctx, pending.ID)
	if reloaded.Status != models.OrderStatusCancelled {
		t.Fatalf("status = %s, want cancelled", reloaded.Status)
	}
	if got := salesTotal(); got != 6000 {
		t.Fatalf("sales after cancel = %d, want 6000 (cancelled order excluded)", got)
	}

	if err := repo.CancelOrder(ctx, pending.ID); !errors.Is(err, ErrOrderNotCancellable) {
		t.Fatalf("second cancel err = %v, want ErrOrderNotCancellable", err)
	}
	if err := repo.CancelOrder(ctx, preparing.ID); !errors.Is(err, ErrOrderNotCancellable) {
		t.Fatalf("cancel preparing err = %v, want ErrOrderNotCancellable", err)
	}
	stillPreparing, _ := repo.FindOrderByID(ctx, preparing.ID)
	if stillPreparing.Status != models.OrderStatusPreparing {
		t.Fatalf("preparing order changed to %s", stillPreparing.Status)
	}
}
