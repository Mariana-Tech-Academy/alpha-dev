package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"alpha-dev/models"

	"gorm.io/gorm"
)

// Record what the service asks the repository to retrieve.
type orderHistoryRepoStub struct {
	orders     []models.Order
	total      int64
	err        error
	customerID uint
	limit      int
	offset     int
	called     bool
}

func (f *orderHistoryRepoStub) PlaceOrder(
	context.Context,
	uint,
	*models.Order,
) (*models.Order, error) {
	return nil, errors.New("checkout is not used in these tests")
}

func (f *orderHistoryRepoStub) FindCustomerOrders(
	_ context.Context,
	customerID uint,
	limit int,
	offset int,
) ([]models.Order, int64, error) {
	f.called = true
	f.customerID = customerID
	f.limit = limit
	f.offset = offset

	return f.orders, f.total, f.err
}

func TestOrderHistoryReturnsRequestedPage(t *testing.T) {
	date := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)

	repo := &orderHistoryRepoStub{
		total: 3,
		orders: []models.Order{
			{
				Model:        gorm.Model{ID: 7, CreatedAt: date},
				OrderNumber:  "ORD-TEST-7",
				CustomerID:   4,
				RestaurantID: 2,
				Restaurant:   models.Restaurant{Name: "Test Restaurant"},
				Total:        1500,
				Status:       models.OrderStatusDelivered,
			},
		},
	}

	history, err := NewOrderService(repo).GetOrderHistory(
		context.Background(), 4, 2, 2,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The repository must receive the logged-in customer's ID.
	if repo.customerID != 4 || repo.limit != 2 || repo.offset != 2 {
		t.Fatalf("wrong customer or page: %+v", repo)
	}

	if history.Page != 2 || history.Limit != 2 || history.Total != 3 {
		t.Fatalf("wrong pagination: %+v", history)
	}

	if len(history.Orders) != 1 {
		t.Fatalf("expected one order, got %d", len(history.Orders))
	}

	item := history.Orders[0]
	if item.ID != 7 ||
		item.OrderNumber != "ORD-TEST-7" ||
		item.RestaurantID != 2 ||
		item.RestaurantName != "Test Restaurant" ||
		item.Total != 1500 ||
		item.Status != models.OrderStatusDelivered ||
		!item.CreatedAt.Equal(date) {
		t.Fatalf("wrong history details: %+v", item)
	}
}

func TestOrderHistoryRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name       string
		customerID uint
		page       int
		limit      int
		want       error
	}{
		{"missing customer", 0, 1, 10, ErrOrderHistoryUnauthenticated},
		{"zero page", 4, 0, 10, ErrOrderHistoryPagination},
		{"negative page", 4, -1, 10, ErrOrderHistoryPagination},
		{"zero limit", 4, 1, 0, ErrOrderHistoryPagination},
		{"large limit", 4, 1, 101, ErrOrderHistoryPagination},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &orderHistoryRepoStub{}

			_, err := NewOrderService(repo).GetOrderHistory(
				context.Background(),
				test.customerID,
				test.page,
				test.limit,
			)

			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
			if repo.called {
				t.Fatal("invalid request reached the repository")
			}
		})
	}
}

func TestOrderHistoryReturnsEmptyList(t *testing.T) {
	history, err := NewOrderService(&orderHistoryRepoStub{}).
		GetOrderHistory(context.Background(), 4, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if history.Orders == nil || len(history.Orders) != 0 {
		t.Fatal("expected an empty list")
	}
}

func TestOrderHistoryReturnsDatabaseError(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	repo := &orderHistoryRepoStub{err: databaseErr}

	_, err := NewOrderService(repo).GetOrderHistory(
		context.Background(), 4, 1, 10,
	)
	if !errors.Is(err, databaseErr) {
		t.Fatalf("expected database error, got %v", err)
	}
}
