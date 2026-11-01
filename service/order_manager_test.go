package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type managerOrderRepo struct {
	restaurantID  uint
	findErr       error
	orders        []models.Order
	listErr       error
	sawRestaurant uint
	sawStatus     models.OrderStatus
	listed        bool
}

func (f *managerOrderRepo) PlaceOrder(context.Context, uint, *models.Order) (*models.Order, error) {
	return nil, errors.New("not used")
}

func (f *managerOrderRepo) FindCustomerOrders(context.Context, uint, int, int) ([]models.Order, int64, error) {
	return nil, 0, errors.New("not used")
}

func (f *managerOrderRepo) FindByID(context.Context, uint) (models.Order, error) {
	return models.Order{}, errors.New("not used")
}

func (f *managerOrderRepo) UpdateStatus(context.Context, uint, models.OrderStatus, models.OrderStatus) (bool, error) {
	return false, errors.New("not used")
}

func (f *managerOrderRepo) GetRestaurantIDByManager(context.Context, uint) (uint, error) {
	if f.findErr != nil {
		return 0, f.findErr
	}
	return f.restaurantID, nil
}

func (f *managerOrderRepo) GetOrdersByRestaurant(_ context.Context, restaurantID uint, status models.OrderStatus) ([]models.Order, error) {
	f.listed = true
	f.sawRestaurant = restaurantID
	f.sawStatus = status
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.orders, nil
}

func TestGetManagerOrders(t *testing.T) {
	created := time.Date(2026, 11, 1, 4, 0, 0, 0, time.UTC)
	order := models.Order{
		OrderNumber:          "ORD-20261101-004",
		Status:               models.OrderStatusPreparing,
		Subtotal:             1700,
		DeliveryFee:          0,
		Total:                1700,
		DeliveryName:         "Ada Lovelace",
		DeliveryPhone:        "08012345678",
		DeliveryAddress:      "12 Broad Street",
		DeliveryCity:         "Lagos",
		DeliveryState:        "Lagos",
		DeliveryInstructions: "Ring the bell",
		Items: []models.OrderItem{{
			MenuItemID: 9,
			Name:       "Chicken Burger",
			UnitPrice:  850,
			Quantity:   2,
			LineTotal:  1700,
		}},
	}
	order.ID = 4
	order.CreatedAt = created

	repo := &managerOrderRepo{restaurantID: 3, orders: []models.Order{order}}
	got, err := NewOrderService(repo).GetManagerOrders(context.Background(), 8, " preparing ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.sawRestaurant != 3 || repo.sawStatus != models.OrderStatusPreparing {
		t.Fatalf("restaurant %d status %s", repo.sawRestaurant, repo.sawStatus)
	}
	if len(got) != 1 {
		t.Fatalf("orders = %+v", got)
	}
	row := got[0]
	if row.ID != 4 || row.OrderNumber != "ORD-20261101-004" || row.Total != 1700 || row.DeliveryFee != 0 || row.DeliveryCity != "Lagos" || row.DeliveryState != "Lagos" || row.DeliveryAddress != "12 Broad Street" || !row.CreatedAt.Equal(created) {
		t.Fatalf("order = %+v", row)
	}
	if len(row.Items) != 1 || row.Items[0].UnitPrice != 850 || row.Items[0].LineTotal != 1700 || row.Items[0].Name != "Chicken Burger" {
		t.Fatalf("items = %+v", row.Items)
	}
}

func TestGetManagerOrdersEdges(t *testing.T) {
	t.Run("no orders", func(t *testing.T) {
		got, err := NewOrderService(&managerOrderRepo{restaurantID: 3}).GetManagerOrders(context.Background(), 8, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Fatalf("orders = %#v", got)
		}
	})

	t.Run("no restaurant", func(t *testing.T) {
		repo := &managerOrderRepo{findErr: gorm.ErrRecordNotFound}
		_, err := NewOrderService(repo).GetManagerOrders(context.Background(), 8, "")
		if !errors.Is(err, ErrNoRestaurantManaged) {
			t.Fatalf("expected no restaurant, got %v", err)
		}
		if repo.listed {
			t.Fatal("orders were loaded without a restaurant")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		repo := &managerOrderRepo{restaurantID: 3}
		_, err := NewOrderService(repo).GetManagerOrders(context.Background(), 8, "teleported")
		if !errors.Is(err, ErrInvalidOrderStatus) {
			t.Fatalf("expected invalid status, got %v", err)
		}
		if repo.listed {
			t.Fatal("invalid status reached the repository")
		}
	})

	t.Run("missing manager", func(t *testing.T) {
		repo := &managerOrderRepo{restaurantID: 3}
		_, err := NewOrderService(repo).GetManagerOrders(context.Background(), 0, "")
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected unauthenticated, got %v", err)
		}
		if repo.listed {
			t.Fatal("missing manager reached the repository")
		}
	})
}
