package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
	"alpha-dev/utils"
)

type unusedOrderRepo struct{ t *testing.T }

func (u unusedOrderRepo) FindByID(context.Context, uint) (models.Order, error) {
	return models.Order{}, nil
}

func (u unusedOrderRepo) UpdateStatus(context.Context, uint, models.OrderStatus, models.OrderStatus) (bool, error) {
	return false, nil
}

func (u unusedOrderRepo) FindCustomerOrders(context.Context, uint, int, int) ([]models.Order, int64, error) {
	return nil, 0, nil
}

func (u unusedOrderRepo) GetRestaurantIDByManager(context.Context, uint) (uint, error) {
	return 0, nil
}

func (u unusedOrderRepo) GetOrdersByRestaurant(context.Context, uint, models.OrderStatus) ([]models.Order, error) {
	return nil, nil
}

func (u unusedOrderRepo) PlaceOrder(context.Context, uint, *models.Order) (*models.Order, error) {
	u.t.Fatal("repository called with invalid delivery details")
	return nil, nil
}

func TestPlaceOrderInvalidDeliveryReturns422(t *testing.T) {
	handler := NewOrderHandler(service.NewOrderService(unusedOrderRepo{t}))

	body := `{"delivery_name":"  ","delivery_phone":"abc"}`
	r := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(body))
	r = r.WithContext(auth.WithClaims(r.Context(), &auth.Claims{UserID: 7, Role: models.RoleCustomer}))
	w := httptest.NewRecorder()

	handler.PlaceOrder(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	var resp utils.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"delivery_name", "delivery_phone", "delivery_address", "delivery_city", "delivery_state"} {
		if resp.Errors[key] == "" {
			t.Errorf("expected error for %q, got %v", key, resp.Errors)
		}
	}
}

type historyHandlerRepo struct {
	unusedOrderRepo
	customer uint
}

func (r *historyHandlerRepo) FindCustomerOrders(_ context.Context, customer uint, limit, offset int) ([]models.Order, int64, error) {
	r.customer = customer
	return []models.Order{{CustomerID: customer, DeliveryAddress: "Saved address", Subtotal: 1250, Total: 1250, Items: []models.OrderItem{{Name: "Saved dish", Quantity: 1, UnitPrice: 1250, LineTotal: 1250}}}}, 1, nil
}
func TestHistoryDetailsUseAuthenticatedCustomer(t *testing.T) {
	repo := &historyHandlerRepo{unusedOrderRepo: unusedOrderRepo{t}}
	handler := NewOrderHandler(service.NewOrderService(repo))
	req := httptest.NewRequest(http.MethodGet, "/api/orders?customer_id=999", nil)
	req = req.WithContext(auth.WithClaims(req.Context(), &auth.Claims{UserID: 7, Role: models.RoleCustomer}))
	w := httptest.NewRecorder()
	handler.GetOrderHistory(w, req)
	if w.Code != http.StatusOK || repo.customer != 7 {
		t.Fatalf("customer %d status %d", repo.customer, w.Code)
	}
	var response struct {
		Data service.OrderHistory `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Data.Orders) != 1 || response.Data.Orders[0].DeliveryAddress != "Saved address" || response.Data.Orders[0].Items[0].Name != "Saved dish" {
		t.Fatalf("missing snapshots: %s", w.Body)
	}
	repo.customer = 0
	w = httptest.NewRecorder()
	handler.GetOrderHistory(w, httptest.NewRequest(http.MethodGet, "/api/orders", nil))
	if w.Code != http.StatusUnauthorized || repo.customer != 0 {
		t.Fatal("unauthenticated history access")
	}
}

func TestPlaceOrderRejectsClientPricing(t *testing.T) {
	handler := NewOrderHandler(service.NewOrderService(unusedOrderRepo{t}))
	for _, field := range []string{"subtotal", "delivery_fee", "total"} {
		t.Run(field, func(t *testing.T) {
			body := `{"delivery_name":"Demo Customer","delivery_phone":"07000000000","delivery_address":"12 Demo Street","delivery_city":"London","delivery_state":"London","` + field + `":1}`
			r := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(body))
			r = r.WithContext(auth.WithClaims(r.Context(), &auth.Claims{UserID: 7, Role: models.RoleCustomer}))
			w := httptest.NewRecorder()
			handler.PlaceOrder(w, r)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("client pricing accepted: %d", w.Code)
			}
		})
	}
}
