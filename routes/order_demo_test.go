package routes

import (
	"alpha-dev/auth"
	"alpha-dev/handlers"
	"alpha-dev/models"
	"alpha-dev/service"
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type demoRevocations struct{ revoked bool }

func (s demoRevocations) IsRevoked(context.Context, string) (bool, error) { return s.revoked, nil }
func (s demoRevocations) Revoke(context.Context, string, time.Time) error { return nil }

type demoService struct {
	service.OrderService
	calls    int
	user     uint
	expected models.OrderStatus
}

func (s *demoService) AdvanceDemoOrder(_ context.Context, user uint, role models.Role, id uint, expected models.OrderStatus) (service.UpdatedOrder, error) {
	s.calls++
	s.user = user
	s.expected = expected
	return service.UpdatedOrder{ID: id, Status: models.OrderStatusPreparing}, nil
}
func TestDemoRouteAuthenticationAndValidation(t *testing.T) {
	for _, tc := range []struct {
		name                              string
		role                              models.Role
		token, disabled, revoked, expired bool
		body                              string
		want                              int
	}{
		{name: "no credentials", want: 401},
		{name: "customer", token: true, role: models.RoleCustomer, want: 403},
		{name: "manager", token: true, role: models.RoleManager, want: 200},
		{name: "admin", token: true, role: models.RoleAdmin, want: 200},
		{name: "expired token", token: true, expired: true, role: models.RoleAdmin, want: 401},
		{name: "revoked token", token: true, revoked: true, role: models.RoleAdmin, want: 401},
		{name: "disabled even for admin", token: true, disabled: true, role: models.RoleAdmin, want: 404},
		{name: "cannot choose target", token: true, role: models.RoleAdmin, body: `{"status":"delivered"}`, want: 400},
		{name: "unknown field", token: true, role: models.RoleAdmin, body: `{"expected_status":"pending","user_id":99}`, want: 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range map[string]string{"APP_ENV": "development", "FOODHUB_DEMO_STATUS": "1", "DB_HOST": "localhost", "DB_NAME": "foodhub_new_structure", "GO_ENV": "", "NODE_ENV": "", "ENVIRONMENT": ""} {
				t.Setenv(k, v)
			}
			if tc.disabled {
				t.Setenv("FOODHUB_DEMO_STATUS", "")
			}
			expiry := time.Hour
			if tc.expired {
				expiry = -time.Hour
			}
			tokens := auth.NewJWTManager("test-only-signing-secret", expiry, demoRevocations{tc.revoked})
			svc := &demoService{}
			router := mux.NewRouter()
			registerDemoOrderRoute(router, tokens, handlers.NewOrderHandler(svc))
			body := tc.body
			if body == "" {
				body = `{"expected_status":"pending"}`
			}
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/demo/orders/9/advance", strings.NewReader(body))
			req.RemoteAddr = "127.0.0.1:1234"
			if tc.token {
				token, err := tokens.Generate(4, tc.role)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("status %d body %s", w.Code, w.Body)
			}
			if tc.want == 200 {
				if svc.calls != 1 || svc.user != 4 || svc.expected != models.OrderStatusPending {
					t.Fatal("incorrect authorized operation")
				}
			} else if svc.calls != 0 {
				t.Fatal("rejected request reached status service")
			}
		})
	}
}
