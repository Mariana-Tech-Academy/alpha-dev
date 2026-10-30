package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
)

type fakeCancelService struct {
	service.OrderService
	err error
}

func (f fakeCancelService) CancelOrder(_ context.Context, _ uint, _ models.Role, orderID uint) (service.CancelledOrder, error) {
	if f.err != nil {
		return service.CancelledOrder{}, f.err
	}
	return service.CancelledOrder{ID: orderID, OrderNumber: "ORD-20261030-001", Status: models.OrderStatusCancelled}, nil
}

func TestCancelOrderStatusCodes(t *testing.T) {
	cases := []struct {
		name     string
		id       string
		err      error
		wantCode int
		wantMsg  string
	}{
		{"cancelled", "21", nil, http.StatusOK, "order cancelled"},
		{"bad id", "abc", nil, http.StatusBadRequest, "invalid order ID"},
		{"zero id", "0", nil, http.StatusBadRequest, "invalid order ID"},
		{"not found", "21", service.ErrOrderNotFound, http.StatusNotFound, "order not found"},
		{"not your order", "21", service.ErrNotYourOrder, http.StatusForbidden, "you cannot cancel this order"},
		{"past the cut-off", "21", service.ErrOrderNotCancellable, http.StatusConflict, "order can no longer be cancelled"},
		{"already cancelled", "21", service.ErrOrderAlreadyCancelled, http.StatusConflict, "order is already cancelled"},
		{"database error", "21", errors.New("db down"), http.StatusInternalServerError, "unable to cancel order"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewOrderHandler(fakeCancelService{err: tc.err})
			r := httptest.NewRequest(http.MethodPost, "/api/orders/"+tc.id+"/cancel", nil)
			r = r.WithContext(auth.WithClaims(r.Context(), &auth.Claims{UserID: 4, Role: models.RoleCustomer}))
			r = mux.SetURLVars(r, map[string]string{"id": tc.id})
			w := httptest.NewRecorder()

			handler.CancelOrder(w, r)

			if w.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d, body %s", w.Code, tc.wantCode, w.Body)
			}
			var resp utils.Response
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
			if resp.Message != tc.wantMsg {
				t.Fatalf("message = %q, want %q", resp.Message, tc.wantMsg)
			}
		})
	}
}
