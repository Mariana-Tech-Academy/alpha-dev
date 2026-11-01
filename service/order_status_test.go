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

type statusOrderRepo struct {
	order      models.Order
	findErr    error
	updateErr  error
	stale      bool
	sawCurrent models.OrderStatus
	sawNext    models.OrderStatus
}

func (f *statusOrderRepo) PlaceOrder(context.Context, uint, *models.Order) (*models.Order, error) {
	return nil, errors.New("not used")
}

func (f *statusOrderRepo) FindCustomerOrders(context.Context, uint, int, int) ([]models.Order, int64, error) {
	return nil, 0, errors.New("not used")
}

func (f *statusOrderRepo) FindByID(context.Context, uint) (models.Order, error) {
	if f.findErr != nil {
		return models.Order{}, f.findErr
	}
	return f.order, nil
}

func (f *statusOrderRepo) UpdateStatus(_ context.Context, id uint, current, next models.OrderStatus) (bool, error) {
	f.sawCurrent = current
	f.sawNext = next
	if f.updateErr != nil {
		return false, f.updateErr
	}
	if f.stale || f.order.ID != id || f.order.Status != current {
		return false, nil
	}
	f.order.Status = next
	f.order.UpdatedAt = time.Date(2026, 11, 1, 3, 0, 0, 0, time.UTC)
	return true, nil
}

var _ repository.OrderRepository = (*statusOrderRepo)(nil)

func pendingOrder() models.Order {
	restaurant := models.Restaurant{ManagerID: 4}
	restaurant.ID = 3
	order := models.Order{
		OrderNumber:  "ORD-20261101-001",
		RestaurantID: 3,
		Restaurant:   restaurant,
		Status:       models.OrderStatusPending,
	}
	order.ID = 9
	return order
}

func TestUpdateOrderStatus(t *testing.T) {
	t.Run("manager moves the order one step", func(t *testing.T) {
		repo := &statusOrderRepo{order: pendingOrder()}
		got, err := NewOrderService(repo).UpdateOrderStatus(context.Background(), 4, models.RoleManager, 9, models.OrderStatusConfirmed)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 9 || got.OrderNumber != "ORD-20261101-001" || got.Status != models.OrderStatusConfirmed || got.UpdatedAt.IsZero() {
			t.Fatalf("order = %+v", got)
		}
		if repo.sawCurrent != models.OrderStatusPending || repo.sawNext != models.OrderStatusConfirmed {
			t.Fatalf("update current %s next %s", repo.sawCurrent, repo.sawNext)
		}
	})

	t.Run("admin can update another restaurant", func(t *testing.T) {
		repo := &statusOrderRepo{order: pendingOrder()}
		got, err := NewOrderService(repo).UpdateOrderStatus(context.Background(), 99, models.RoleAdmin, 9, models.OrderStatusConfirmed)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Status != models.OrderStatusConfirmed {
			t.Fatalf("status = %s", got.Status)
		}
	})

	t.Run("another manager is forbidden", func(t *testing.T) {
		repo := &statusOrderRepo{order: pendingOrder()}
		_, err := NewOrderService(repo).UpdateOrderStatus(context.Background(), 8, models.RoleManager, 9, models.OrderStatusConfirmed)
		if !errors.Is(err, ErrOrderForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
		if repo.order.Status != models.OrderStatusPending {
			t.Fatal("forbidden request changed the status")
		}
	})

	t.Run("customer role is forbidden", func(t *testing.T) {
		_, err := NewOrderService(&statusOrderRepo{order: pendingOrder()}).UpdateOrderStatus(context.Background(), 4, models.RoleCustomer, 9, models.OrderStatusConfirmed)
		if !errors.Is(err, ErrOrderForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("missing user", func(t *testing.T) {
		_, err := NewOrderService(&statusOrderRepo{order: pendingOrder()}).UpdateOrderStatus(context.Background(), 0, models.RoleManager, 9, models.OrderStatusConfirmed)
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected unauthenticated, got %v", err)
		}
	})

	t.Run("unknown status", func(t *testing.T) {
		_, err := NewOrderService(&statusOrderRepo{order: pendingOrder()}).UpdateOrderStatus(context.Background(), 4, models.RoleManager, 9, "teleported")
		if !errors.Is(err, ErrInvalidOrderStatus) {
			t.Fatalf("expected invalid status, got %v", err)
		}
	})

	t.Run("skipped status", func(t *testing.T) {
		repo := &statusOrderRepo{order: pendingOrder()}
		err := mustTransition(NewOrderService(repo).UpdateOrderStatus(context.Background(), 4, models.RoleManager, 9, models.OrderStatusPreparing))
		if err == nil || err.From != models.OrderStatusPending || err.Allowed != models.OrderStatusConfirmed {
			t.Fatalf("transition = %+v", err)
		}
		if repo.order.Status != models.OrderStatusPending {
			t.Fatal("skipped status was saved")
		}
	})

	t.Run("delivered is final", func(t *testing.T) {
		order := pendingOrder()
		order.Status = models.OrderStatusDelivered
		err := mustTransition(NewOrderService(&statusOrderRepo{order: order}).UpdateOrderStatus(context.Background(), 4, models.RoleManager, 9, models.OrderStatusCancelled))
		if err == nil || err.Allowed != "" || err.From != models.OrderStatusDelivered {
			t.Fatalf("transition = %+v", err)
		}
	})

	t.Run("missing order", func(t *testing.T) {
		_, err := NewOrderService(&statusOrderRepo{findErr: gorm.ErrRecordNotFound}).UpdateOrderStatus(context.Background(), 4, models.RoleManager, 9, models.OrderStatusConfirmed)
		if !errors.Is(err, ErrOrderNotFound) {
			t.Fatalf("expected not found, got %v", err)
		}
	})

	t.Run("status changed before the update", func(t *testing.T) {
		repo := &statusOrderRepo{order: pendingOrder(), stale: true}
		_, err := NewOrderService(repo).UpdateOrderStatus(context.Background(), 4, models.RoleManager, 9, models.OrderStatusConfirmed)
		if !errors.Is(err, ErrOrderChanged) {
			t.Fatalf("expected conflict, got %v", err)
		}
		if repo.order.Status != models.OrderStatusPending {
			t.Fatal("stale update changed the status")
		}
	})
}

func mustTransition(order UpdatedOrder, err error) *OrderTransitionError {
	if order.ID != 0 {
		return nil
	}
	var transition *OrderTransitionError
	if !errors.As(err, &transition) {
		return nil
	}
	return transition
}
