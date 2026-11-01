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

func (f *statusOrderRepo) GetRestaurantIDByManager(context.Context, uint) (uint, error) {
	return 0, errors.New("not used")
}

func (f *statusOrderRepo) GetOrdersByRestaurant(context.Context, uint, models.OrderStatus) ([]models.Order, error) {
	return nil, errors.New("not used")
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

func TestCompleteNativeOrderFlowAndCancellationPolicy(t *testing.T) {
	flow := []models.OrderStatus{models.OrderStatusPending, models.OrderStatusConfirmed, models.OrderStatusPreparing, models.OrderStatusReadyForDelivery, models.OrderStatusOutForDelivery, models.OrderStatusDelivered}
	repo := &statusOrderRepo{order: pendingOrder()}
	svc := NewOrderService(repo)
	for _, next := range flow[1:] {
		result, err := svc.UpdateOrderStatus(context.Background(), 4, models.RoleManager, 9, next)
		if err != nil || result.Status != next {
			t.Fatalf("step %s: %+v %v", next, result, err)
		}
	}
	for _, from := range append(flow, models.OrderStatusCancelled) {
		for _, next := range append(flow, models.OrderStatusCancelled) {
			repo.order.Status = from
			allowed, ok := from.Next()
			_, err := svc.UpdateOrderStatus(context.Background(), 4, models.RoleManager, 9, next)
			if (err == nil) != (ok && next == allowed) {
				t.Fatalf("transition %s -> %s: %v", from, next, err)
			}
			if err != nil && repo.order.Status != from {
				t.Fatal("rejected transition changed order")
			}
		}
	}
}

func TestDemoLifecyclePersistsFourStages(t *testing.T) {
	repo := &statusOrderRepo{order: pendingOrder()}
	svc := NewOrderService(repo)
	for _, step := range []struct{ from, to models.OrderStatus }{
		{models.OrderStatusPending, models.OrderStatusPreparing},
		{models.OrderStatusPreparing, models.OrderStatusOutForDelivery},
		{models.OrderStatusOutForDelivery, models.OrderStatusDelivered},
	} {
		got, err := svc.AdvanceDemoOrder(context.Background(), 4, models.RoleManager, 9, step.from)
		if err != nil || got.Status != step.to || repo.order.Status != step.to || repo.sawCurrent != step.from {
			t.Fatalf("step %+v got %+v err %v", step, got, err)
		}
		stored, err := repo.FindByID(context.Background(), 9)
		if err != nil || stored.Status != step.to {
			t.Fatal("status not persisted")
		}
	}
}
func TestDemoLifecycleRejectsUnauthorisedAndStaleRequests(t *testing.T) {
	for _, tc := range []struct {
		name              string
		user              uint
		role              models.Role
		expected, current models.OrderStatus
		stale             bool
		want              error
	}{
		{name: "customer", user: 4, role: models.RoleCustomer, expected: models.OrderStatusPending, current: models.OrderStatusPending, want: ErrOrderForbidden},
		{name: "other manager", user: 8, role: models.RoleManager, expected: models.OrderStatusPending, current: models.OrderStatusPending, want: ErrOrderForbidden},
		{name: "unauthenticated", role: models.RoleAdmin, expected: models.OrderStatusPending, current: models.OrderStatusPending, want: ErrInvalidUser},
		{name: "stale expected status", user: 4, role: models.RoleManager, expected: models.OrderStatusPreparing, current: models.OrderStatusPending, want: ErrOrderChanged},
		{name: "concurrent update", user: 4, role: models.RoleManager, expected: models.OrderStatusPending, current: models.OrderStatusPending, stale: true, want: ErrOrderChanged},
		{name: "invalid status", user: 4, role: models.RoleManager, expected: "complete", current: models.OrderStatusPending, want: ErrInvalidOrderStatus},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &statusOrderRepo{order: pendingOrder(), stale: tc.stale}
			repo.order.Status = tc.current
			_, err := NewOrderService(repo).AdvanceDemoOrder(context.Background(), tc.user, tc.role, 9, tc.expected)
			if !errors.Is(err, tc.want) || repo.order.Status != tc.current {
				t.Fatalf("error %v state %s", err, repo.order.Status)
			}
		})
	}
	for _, terminal := range []models.OrderStatus{models.OrderStatusDelivered, models.OrderStatusCancelled} {
		repo := &statusOrderRepo{order: pendingOrder()}
		repo.order.Status = terminal
		_, err := NewOrderService(repo).AdvanceDemoOrder(context.Background(), 99, models.RoleAdmin, 9, terminal)
		var transition *OrderTransitionError
		if !errors.As(err, &transition) || repo.order.Status != terminal {
			t.Fatal("terminal state changed")
		}
	}
}
