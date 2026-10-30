package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

type orderHistoryRepoStub struct {
	orders     []models.Order
	total      int64
	err        error
	customerID uint
	limit      int
	offset     int
	called     bool
}

func (f *orderHistoryRepoStub) PlaceOrder(context.Context, uint, *models.Order) (*models.Order, error) {
	return nil, errors.New("checkout is not used in these tests")
}

func (f *orderHistoryRepoStub) FindCustomerOrders(_ context.Context, customerID uint, limit, offset int) ([]models.Order, int64, error) {
	f.called = true
	f.customerID = customerID
	f.limit = limit
	f.offset = offset
	return f.orders, f.total, f.err
}

var _ repository.OrderRepository = (*orderHistoryRepoStub)(nil)

func TestOrderHistoryReturnsRequestedPage(t *testing.T) {
	date := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	repo := &orderHistoryRepoStub{
		total: 3,
		orders: []models.Order{{
			Model:        gorm.Model{ID: 7, CreatedAt: date},
			OrderNumber:  "ORD-TEST-7",
			CustomerID:   4,
			RestaurantID: 2,
			Restaurant:   models.Restaurant{Name: "Test Restaurant"},
			Total:        1500,
			Status:       models.OrderStatusDelivered,
		}},
	}

	history, err := NewOrderService(repo).GetOrderHistory(context.Background(), 4, 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.customerID != 4 || repo.limit != 2 || repo.offset != 2 {
		t.Fatalf("wrong customer or page: customer %d limit %d offset %d", repo.customerID, repo.limit, repo.offset)
	}
	if history.Page != 2 || history.Limit != 2 || history.Total != 3 {
		t.Fatalf("wrong pagination: %+v", history)
	}
	if len(history.Orders) != 1 {
		t.Fatalf("expected one order, got %d", len(history.Orders))
	}
	item := history.Orders[0]
	if item.ID != 7 || item.OrderNumber != "ORD-TEST-7" || item.RestaurantID != 2 || item.RestaurantName != "Test Restaurant" || item.Total != 1500 || item.Status != models.OrderStatusDelivered || !item.CreatedAt.Equal(date) {
		t.Fatalf("wrong history details: %+v", item)
	}
}

func TestOrderHistoryClampsPageSize(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		wantPage   int
		wantLimit  int
		wantOffset int
	}{
		{"zero page", 0, 10, 1, 10, 0},
		{"negative page", -1, 10, 1, 10, 0},
		{"zero limit", 1, 0, 1, 10, 0},
		{"large limit", 1, 101, 1, 20, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &orderHistoryRepoStub{}
			history, err := NewOrderService(repo).GetOrderHistory(context.Background(), 4, test.page, test.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if history.Page != test.wantPage || history.Limit != test.wantLimit {
				t.Fatalf("page %d limit %d", history.Page, history.Limit)
			}
			if repo.offset != test.wantOffset || repo.limit != test.wantLimit {
				t.Fatalf("repository offset %d limit %d", repo.offset, repo.limit)
			}
		})
	}
}

func TestOrderHistoryRejectsMissingCustomer(t *testing.T) {
	repo := &orderHistoryRepoStub{}
	_, err := NewOrderService(repo).GetOrderHistory(context.Background(), 0, 1, 10)
	if !errors.Is(err, ErrInvalidUser) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if repo.called {
		t.Fatal("missing customer reached the repository")
	}
}

func TestOrderHistoryReturnsEmptyList(t *testing.T) {
	history, err := NewOrderService(&orderHistoryRepoStub{}).GetOrderHistory(context.Background(), 4, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if history.Orders == nil || len(history.Orders) != 0 {
		t.Fatal("expected an empty list")
	}
}

func TestOrderHistoryReturnsDatabaseError(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	_, err := NewOrderService(&orderHistoryRepoStub{err: databaseErr}).GetOrderHistory(context.Background(), 4, 1, 10)
	if !errors.Is(err, databaseErr) {
		t.Fatalf("expected database error, got %v", err)
	}
}
