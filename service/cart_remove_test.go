package service

import (
	"alpha-dev/models"
	"alpha-dev/repository"
	"context"
	"errors"
	"gorm.io/gorm"
	"testing"
)

type removalRepo struct {
	repository.CartRepository
	cart    models.Cart
	failure error
	calls   int
}

func (r *removalRepo) FindCartByUserID(_ context.Context, user uint) (*models.Cart, error) {
	if user != r.cart.UserID {
		return nil, gorm.ErrRecordNotFound
	}
	return &r.cart, nil
}
func (r *removalRepo) RemoveItems(_ context.Context, user uint, item *uint) error {
	r.calls++
	if r.failure != nil {
		return r.failure
	}
	if user != r.cart.UserID {
		return gorm.ErrRecordNotFound
	}
	if item == nil {
		r.cart.Items = nil
		r.cart.RestaurantID = nil
		return nil
	}
	for i, line := range r.cart.Items {
		if line.ID == *item {
			r.cart.Items = append(r.cart.Items[:i], r.cart.Items[i+1:]...)
			if len(r.cart.Items) == 0 {
				r.cart.RestaurantID = nil
			}
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}
func TestRemoveCartItems(t *testing.T) {
	for _, tc := range []struct {
		name      string
		user      uint
		item      *uint
		failure   error
		want      error
		remaining int
	}{
		{"one line", 7, new(uint(11)), nil, nil, 1},
		{"clear", 7, nil, nil, nil, 0},
		{"other customer", 8, new(uint(11)), nil, ErrCartItemNotFound, 2},
		{"missing line", 7, new(uint(99)), nil, ErrCartItemNotFound, 2},
		{"invalid line", 7, new(uint(0)), nil, ErrCartItemNotFound, 2},
		{"unauthenticated", 0, nil, nil, ErrInvalidUser, 2},
		{"storage failure", 7, nil, errors.New("offline"), errors.New("offline"), 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restaurantID := uint(3)
			r := &removalRepo{cart: models.Cart{UserID: 7, RestaurantID: &restaurantID, Items: []models.CartItem{{Model: gorm.Model{ID: 11}}, {Model: gorm.Model{ID: 12}}}}, failure: tc.failure}
			view, err := NewCartService(r, nil).RemoveItems(context.Background(), tc.user, tc.item)
			if tc.want != nil {
				if err == nil || err.Error() != tc.want.Error() {
					t.Fatalf("error %v, want %v", err, tc.want)
				}
			} else if err != nil || len(view.Items) != tc.remaining {
				t.Fatalf("view %+v, error %v", view, err)
			}
			if len(r.cart.Items) != tc.remaining {
				t.Fatal("unexpected cart mutation")
			}
			if tc.user == 0 || (tc.item != nil && *tc.item == 0) {
				if r.calls != 0 {
					t.Fatal("invalid request reached storage")
				}
			}
			if tc.remaining == 0 && r.cart.RestaurantID != nil {
				t.Fatal("empty cart retains restaurant")
			}
		})
	}
}
