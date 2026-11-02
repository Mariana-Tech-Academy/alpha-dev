package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"alpha-dev/auth"
	"alpha-dev/handlers"
	"alpha-dev/middleware"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// Isolated state store: exercise the real JWT middleware, handler and service
// without connecting to PostgreSQL or changing any existing order.
type demoFlowRepo struct {
	repository.OrderRepository
	order  models.Order
	writes int
}

func (r *demoFlowRepo) FindByID(_ context.Context, id uint) (models.Order, error) {
	if id != r.order.ID {
		return models.Order{}, gorm.ErrRecordNotFound
	}
	return r.order, nil
}
func (r *demoFlowRepo) UpdateStatus(_ context.Context, id uint, current, next models.OrderStatus) (bool, error) {
	if id != r.order.ID || current != r.order.Status {
		return false, nil
	}
	r.order.Status = next
	r.order.UpdatedAt = time.Now()
	r.writes++
	return true, nil
}
func (r *demoFlowRepo) FindCustomerOrders(_ context.Context, user uint, limit, offset int) ([]models.Order, int64, error) {
	if user != r.order.CustomerID {
		return []models.Order{}, 0, nil
	}
	return []models.Order{r.order}, 1, nil
}
func TestDemoHTTPFlowPersistsAndRefreshesOwnHistory(t *testing.T) {
	for key, value := range map[string]string{"APP_ENV": "development", "FOODHUB_DEMO_STATUS": "1", "DB_HOST": "localhost", "DB_NAME": "foodhub_new_structure", "GO_ENV": "", "NODE_ENV": "", "ENVIRONMENT": ""} {
		t.Setenv(key, value)
	}
	repo := &demoFlowRepo{order: models.Order{
		Model: gorm.Model{ID: 9}, CustomerID: 7, RestaurantID: 3, Restaurant: models.Restaurant{ManagerID: 4},
		OrderNumber: "ORD-SNAPSHOT", Status: models.OrderStatusPending, Subtotal: 2500, DeliveryFee: 100, Total: 2600,
		DeliveryName: "Saved Customer", DeliveryAddress: "Saved Street", DeliveryPhone: "07000000000",
		Items: []models.OrderItem{{Name: "Saved Burger", Quantity: 2, UnitPrice: 1250, LineTotal: 2500}},
	}}
	original := repo.order
	manager := auth.NewJWTManager("test-only-flow-secret", time.Hour, demoRevocations{})
	handler := handlers.NewOrderHandler(service.NewOrderService(repo))
	router := mux.NewRouter()
	registerDemoOrderRoute(router, manager, handler)
	customer := protected(router, manager)
	customer.Use(middleware.RequireRole(models.RoleCustomer))
	customer.HandleFunc("/orders", handler.GetOrderHistory).Methods(http.MethodGet)
	call := func(method, path, body string, user uint, role models.Role) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, "http://localhost:8080"+path, strings.NewReader(body))
		req.RemoteAddr = "127.0.0.1:1234"
		if user != 0 {
			token, err := manager.Generate(user, role)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	for _, tc := range []struct {
		user uint
		role models.Role
		body string
		want int
	}{
		{0, models.RoleManager, `{"expected_status":"pending"}`, 401},
		{7, models.RoleCustomer, `{"expected_status":"pending"}`, 403},
		{8, models.RoleManager, `{"expected_status":"pending"}`, 403},
		{4, models.RoleManager, `{"expected_status":"preparing"}`, 409},
		{4, models.RoleManager, `{"expected_status":"complete"}`, 422},
	} {
		w := call(http.MethodPost, "/demo/orders/9/advance", tc.body, tc.user, tc.role)
		if w.Code != tc.want || repo.writes != 0 {
			t.Fatalf("denied operation: status %d writes %d body %s", w.Code, repo.writes, w.Body)
		}
	}
	for i, step := range []struct{ from, to models.OrderStatus }{
		{models.OrderStatusPending, models.OrderStatusPreparing},
		{models.OrderStatusPreparing, models.OrderStatusOutForDelivery},
		{models.OrderStatusOutForDelivery, models.OrderStatusDelivered},
	} {
		// Managers and admins both use their own staff JWT, never the customer JWT.
		staff, role := uint(4), models.RoleManager
		if i == 1 {
			staff, role = 99, models.RoleAdmin
		}
		w := call(http.MethodPost, "/demo/orders/9/advance", `{"expected_status":"`+string(step.from)+`"}`, staff, role)
		if w.Code != 200 || repo.order.Status != step.to {
			t.Fatalf("advance %s: %d %s", step.to, w.Code, w.Body)
		}
		w = call(http.MethodGet, "/orders", "", 7, models.RoleCustomer)
		var response struct {
			Data service.OrderHistory `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || len(response.Data.Orders) != 1 || response.Data.Orders[0].Status != step.to {
			t.Fatalf("history did not reload stored stage: %s", w.Body)
		}
		item := response.Data.Orders[0]
		if item.Items[0].Name != "Saved Burger" || item.Items[0].UnitPrice != 1250 || item.DeliveryAddress != "Saved Street" || item.Total != 2600 {
			t.Fatal("purchased details changed")
		}
	}
	w := call(http.MethodGet, "/orders", "", 8, models.RoleCustomer)
	var other struct {
		Data service.OrderHistory `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &other); err != nil {
		t.Fatal(err)
	}
	if len(other.Data.Orders) != 0 {
		t.Fatal("customer could see another customer's history")
	}
	w = call(http.MethodPost, "/demo/orders/9/advance", `{"expected_status":"delivered"}`, 4, models.RoleManager)
	if w.Code != 409 || repo.writes != 3 {
		t.Fatal("completed order advanced again")
	}
	repo.order.Status = original.Status
	repo.order.UpdatedAt = original.UpdatedAt
	if !reflect.DeepEqual(repo.order, original) {
		t.Fatal("status operation modified order snapshots or other fields")
	}
}
