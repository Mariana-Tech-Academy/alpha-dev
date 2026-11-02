package main

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"
)

// Only repository storage is in-memory. These tests call the existing production
// cart and checkout services; they do not execute the seed, SQL or HTTP writes.
type fixtureMenu struct {
	restaurant models.Restaurant
	item       models.MenuItem
}

func (f *fixtureMenu) FindRestaurantByID(_ context.Context, id uint) (models.Restaurant, error) {
	if id != f.restaurant.ID {
		return models.Restaurant{}, errors.New("wrong restaurant")
	}
	return f.restaurant, nil
}
func (f *fixtureMenu) FindMenuItemByID(_ context.Context, id uint) (models.MenuItem, error) {
	if id != f.item.ID {
		return models.MenuItem{}, errors.New("wrong menu item")
	}
	return f.item, nil
}
func (f *fixtureMenu) FindCategoryByID(context.Context, uint) (models.Category, error) {
	return models.Category{}, errors.New("unexpected call")
}
func (f *fixtureMenu) CreateMenuItem(context.Context, *models.MenuItem) error {
	return errors.New("unexpected call")
}
func (f *fixtureMenu) UpdateMenuItemAvailability(context.Context, uint, bool) error {
	return errors.New("unexpected call")
}
func (f *fixtureMenu) UpdateMenuItem(context.Context, uint, map[string]any) error {
	return errors.New("unexpected call")
}
func (f *fixtureMenu) DeleteMenuItem(context.Context, uint) error {
	return errors.New("unexpected call")
}
func (f *fixtureMenu) SearchMenuItems(context.Context, repository.MenuSearchQuery) ([]models.MenuSearchResult, error) {
	return nil, errors.New("unexpected call")
}
func (f *fixtureMenu) FindMenuByRestaurant(context.Context, uint, uint) ([]models.MenuItem, error) {
	return nil, errors.New("unexpected call")
}

type fixtureCart struct {
	menu *fixtureMenu
	cart *models.Cart
}

func (f *fixtureCart) AddItem(_ context.Context, userID, menuItemID, restaurantID uint, quantity int, price int64) (*models.CartItem, error) {
	if menuItemID != f.menu.item.ID || restaurantID != f.menu.restaurant.ID {
		return nil, errors.New("wrong relationship")
	}
	line := models.CartItem{CartID: 1, MenuItemID: menuItemID, MenuItem: f.menu.item, Quantity: quantity, Price: price}
	line.ID = 1
	f.cart = &models.Cart{UserID: userID, RestaurantID: &restaurantID, Restaurant: &f.menu.restaurant, Items: []models.CartItem{line}}
	f.cart.ID = 1
	return &line, nil
}
func (f *fixtureCart) FindCartByUserID(_ context.Context, id uint) (*models.Cart, error) {
	if f.cart == nil || f.cart.UserID != id {
		return nil, errors.New("wrong customer")
	}
	return f.cart, nil
}
func (f *fixtureCart) UpdateQuantity(context.Context, uint, uint, int) error {
	return errors.New("unexpected call")
}

func TestEveryPresentationDishSupportsRealCartAndCheckout(t *testing.T) {
	rows, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for ri, r := range rows {
		for di, d := range r.Items {
			t.Run(r.Key+"/"+d.Key, func(t *testing.T) {
				restaurantID, categoryID := uint(ri+1), uint(di+1)
				restaurant := restaurantRecord(r, 99)
				restaurant.ID = restaurantID
				item := menuRecord(d, restaurantID, categoryID)
				item.ID = uint(ri*8 + di + 1)
				item.Category = models.Category{Name: d.Category}
				item.Category.ID = categoryID
				if restaurant.Status != models.RestaurantStatusActive || restaurant.DeletedAt.Valid {
					t.Fatal("restaurant is not public/checkout eligible")
				}
				if !item.Available || item.DeletedAt.Valid || item.RestaurantID != restaurant.ID || item.CategoryID != item.Category.ID {
					t.Fatal("invalid menu availability/relationships")
				}
				if item.Name != d.Name || item.Description != d.Description || item.Price != d.Price || item.ImageURL != d.ImageURL {
					t.Fatal("seed changed dish details")
				}
				menu := &fixtureMenu{restaurant: restaurant, item: item}
				cart := &fixtureCart{menu: menu}
				cartService := service.NewCartService(cart, menu)
				added, err := cartService.AddItem(context.Background(), 42, item.ID, 2)
				if err != nil {
					t.Fatal(err)
				}
				if added.Price != d.Price {
					t.Fatal("cart price is not the catalogue price in pence")
				}
				summary, err := service.NewCheckoutService(cart).Summary(context.Background(), 42)
				if err != nil {
					t.Fatal(err)
				}
				if summary.Restaurant.ID != restaurantID || len(summary.Items) != 1 || summary.Items[0].MenuItemID != item.ID || summary.Items[0].Name != d.Name || summary.Subtotal != 2*d.Price || summary.DeliveryFee != 299 || summary.Total != 2*d.Price+299 {
					t.Fatal("checkout identity or server totals do not match dish")
				}
			})
		}
	}
}

func (f *fixtureCart) RemoveItems(context.Context, uint, *uint) error { return nil }
