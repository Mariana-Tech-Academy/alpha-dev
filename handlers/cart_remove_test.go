package handlers

import (
	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

type removalService struct {
	service.CartService
	user  uint
	item  *uint
	err   error
	calls int
}

func (s *removalService) RemoveItems(_ context.Context, user uint, item *uint) (*models.CartViewResponse, error) {
	s.calls++
	s.user = user
	s.item = item
	return &models.CartViewResponse{Items: []models.CartItemResponse{}}, s.err
}
func TestRemoveCartHandler(t *testing.T) {
	for _, tc := range []struct {
		name, path    string
		authenticated bool
		err           error
		status        int
	}{
		{"remove", "/cart/items/11", true, nil, 200},
		{"clear", "/cart", true, nil, 200},
		{"missing authentication", "/cart", false, nil, 401},
		{"invalid ID", "/cart/items/no", true, nil, 400},
		{"zero ID", "/cart/items/0", true, nil, 400},
		{"missing or foreign line", "/cart/items/11", true, service.ErrCartItemNotFound, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &removalService{err: tc.err}
			h := NewCartHandler(s)
			router := mux.NewRouter()
			router.HandleFunc("/cart", h.RemoveItems).Methods(http.MethodDelete)
			router.HandleFunc("/cart/items/{id}", h.RemoveItems).Methods(http.MethodDelete)
			req := httptest.NewRequest(http.MethodDelete, tc.path, nil)
			if tc.authenticated {
				req = req.WithContext(auth.WithClaims(req.Context(), &auth.Claims{UserID: 7}))
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status %d, body %s", w.Code, w.Body)
			}
			if tc.status == 200 || tc.status == 404 {
				if s.user != 7 {
					t.Fatal("wrong customer")
				}
				if tc.path == "/cart" {
					if s.item != nil {
						t.Fatal("clear was scoped to a line")
					}
				} else if s.item == nil || *s.item != 11 {
					t.Fatal("wrong cart line")
				}
			} else if s.calls != 0 {
				t.Fatal("invalid request reached service")
			}
		})
	}
}
