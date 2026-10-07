package models

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

// GORM parses every relationship without needing a running database.
func TestModelRelationships(t *testing.T) {
	checks := []struct {
		model     interface{}
		table     string
		relations map[string]schema.RelationshipType
	}{
		{&User{}, "users", map[string]schema.RelationshipType{"Cart": schema.HasOne, "Orders": schema.HasMany, "ManagedRestaurants": schema.HasMany}},
		{&Restaurant{}, "restaurants", map[string]schema.RelationshipType{"Manager": schema.BelongsTo, "MenuItems": schema.HasMany, "Orders": schema.HasMany}},
		{&Category{}, "categories", map[string]schema.RelationshipType{"MenuItems": schema.HasMany}},
		{&MenuItem{}, "menu_items", map[string]schema.RelationshipType{"Restaurant": schema.BelongsTo, "Category": schema.BelongsTo}},
		{&Cart{}, "carts", map[string]schema.RelationshipType{"User": schema.BelongsTo, "Restaurant": schema.BelongsTo, "Items": schema.HasMany}},
		{&CartItem{}, "cart_items", map[string]schema.RelationshipType{"Cart": schema.BelongsTo, "MenuItem": schema.BelongsTo}},
		{&Order{}, "orders", map[string]schema.RelationshipType{"Customer": schema.BelongsTo, "Restaurant": schema.BelongsTo, "Items": schema.HasMany}},
		{&OrderItem{}, "order_items", map[string]schema.RelationshipType{"Order": schema.BelongsTo, "MenuItem": schema.BelongsTo}},
	}
	cache := &sync.Map{}
	for _, check := range checks {
		t.Run(check.table, func(t *testing.T) {
			s, err := schema.Parse(check.model, cache, schema.NamingStrategy{})
			if err != nil {
				t.Fatal(err)
			}
			if s.Table != check.table {
				t.Fatalf("table: got %s, want %s", s.Table, check.table)
			}
			for name, want := range check.relations {
				r := s.Relationships.Relations[name]
				if r == nil {
					t.Fatalf("missing relationship %s", name)
				}
				if r.Type != want {
					t.Errorf("%s: got %s, want %s", name, r.Type, want)
				}
			}
		})
	}
}

func TestPasswordHashNeverAppearsInJSON(t *testing.T) {
	data, err := json.Marshal(User{Email: "user@example.com", PasswordHash: "private-hash"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private-hash") || strings.Contains(string(data), "password") {
		t.Fatal("user JSON exposed password information")
	}
}

func TestOrderDeliveryIsEmbeddedSnapshot(t *testing.T) {
	s, err := schema.Parse(&Order{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{"delivery_recipient_name", "delivery_phone_number", "delivery_address", "delivery_city", "delivery_state", "delivery_instructions"} {
		if s.FieldsByDBName[column] == nil {
			t.Errorf("missing snapshot column %s", column)
		}
	}
}
