package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

const (
	testCustomerID = 4
	testManagerID  = 9
	testOrderID    = 21
)

type cancelOrderRepo struct {
	fakeOrderRepo
	order     *models.Order
	findErr   error
	cancelErr error
	cancelled uint
}

func (f *cancelOrderRepo) FindOrderByID(_ context.Context, id uint) (*models.Order, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	copied := *f.order
	return &copied, nil
}

func (f *cancelOrderRepo) CancelOrder(_ context.Context, id uint) error {
	if f.cancelErr != nil {
		return f.cancelErr
	}
	f.cancelled = id
	return nil
}

func orderIn(status models.OrderStatus) *models.Order {
	order := &models.Order{
		OrderNumber: "ORD-20261030-001",
		CustomerID:  testCustomerID,
		Status:      status,
		Restaurant:  models.Restaurant{ManagerID: testManagerID},
	}
	order.ID = testOrderID
	return order
}

func TestCancelOrder(t *testing.T) {
	ctx := context.Background()

	allowed := []struct {
		name   string
		userID uint
		role   models.Role
		status models.OrderStatus
	}{
		{"customer cancels own pending order", testCustomerID, models.RoleCustomer, models.OrderStatusPending},
		{"customer cancels own confirmed order", testCustomerID, models.RoleCustomer, models.OrderStatusConfirmed},
		{"manager cancels own restaurant's order", testManagerID, models.RoleManager, models.OrderStatusConfirmed},
		{"admin cancels any order", 1, models.RoleAdmin, models.OrderStatusPending},
	}
	for _, tc := range allowed {
		t.Run(tc.name, func(t *testing.T) {
			repo := &cancelOrderRepo{order: orderIn(tc.status)}
			got, err := NewOrderService(repo).CancelOrder(ctx, tc.userID, tc.role, testOrderID)
			if err != nil {
				t.Fatalf("cancel: %v", err)
			}
			if repo.cancelled != testOrderID {
				t.Fatalf("repository cancelled order %d, want %d", repo.cancelled, testOrderID)
			}
			if got.Status != models.OrderStatusCancelled || got.OrderNumber != "ORD-20261030-001" {
				t.Fatalf("got %+v", got)
			}
		})
	}

	rejected := []struct {
		name    string
		userID  uint
		role    models.Role
		status  models.OrderStatus
		wantErr error
	}{
		{"another customer's order", 5, models.RoleCustomer, models.OrderStatusPending, ErrNotYourOrder},
		{"manager of another restaurant", 10, models.RoleManager, models.OrderStatusPending, ErrNotYourOrder},
		{"customer ID used with manager role", testCustomerID, models.RoleManager, models.OrderStatusPending, ErrNotYourOrder},
		{"order is preparing", testCustomerID, models.RoleCustomer, models.OrderStatusPreparing, ErrOrderNotCancellable},
		{"order is out for delivery (admin)", 1, models.RoleAdmin, models.OrderStatusOutForDelivery, ErrOrderNotCancellable},
		{"order is delivered", testCustomerID, models.RoleCustomer, models.OrderStatusDelivered, ErrOrderNotCancellable},
		{"order is already cancelled", testCustomerID, models.RoleCustomer, models.OrderStatusCancelled, ErrOrderAlreadyCancelled},
		{"no user", 0, models.RoleCustomer, models.OrderStatusPending, ErrInvalidUser},
	}
	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			repo := &cancelOrderRepo{order: orderIn(tc.status)}
			_, err := NewOrderService(repo).CancelOrder(ctx, tc.userID, tc.role, testOrderID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if repo.cancelled != 0 {
				t.Fatal("repository cancel should not be called")
			}
		})
	}

	t.Run("order not found", func(t *testing.T) {
		repo := &cancelOrderRepo{findErr: gorm.ErrRecordNotFound}
		_, err := NewOrderService(repo).CancelOrder(ctx, testCustomerID, models.RoleCustomer, testOrderID)
		if !errors.Is(err, ErrOrderNotFound) {
			t.Fatalf("err = %v, want %v", err, ErrOrderNotFound)
		}
	})

	t.Run("status moved past confirmed while cancelling", func(t *testing.T) {
		repo := &cancelOrderRepo{order: orderIn(models.OrderStatusConfirmed), cancelErr: repository.ErrOrderNotCancellable}
		_, err := NewOrderService(repo).CancelOrder(ctx, testCustomerID, models.RoleCustomer, testOrderID)
		if !errors.Is(err, ErrOrderNotCancellable) {
			t.Fatalf("err = %v, want %v", err, ErrOrderNotCancellable)
		}
	})
}
