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
