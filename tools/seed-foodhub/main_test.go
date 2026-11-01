package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Offline tests only: these never call applySeed, open PostgreSQL or run main.
func TestPresentationCatalog(t *testing.T) {
	rows, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"Burger District", "Bella Italia Kitchen", "Spice Route", "The Shawarma House", "Sweet Treats Café", "Tokyo Bento Kitchen"}
	total, available := 0, 0
	categories := map[string]bool{}
	for i, r := range rows {
		if r.Name != expected[i] || len(r.Items) != 8 {
			t.Fatalf("unexpected restaurant/menu: %s", r.Name)
		}
		perRestaurant := 0
		for _, d := range r.Items {
			total++
			categories[d.Category] = true
			if d.Available {
				available++
				perRestaurant++
			}
		}
		if perRestaurant != 8 {
			t.Fatalf("expected eight orderable dishes for %s", r.Name)
		}
	}
	if total != 48 || available != 48 {
		t.Fatalf("unexpected totals: %d/%d", total, available)
	}
	t.Logf("6 restaurants, %d dishes, %d available, %d unavailable, %d categories", total, available, total-available, len(categories))
}
func TestRefusesWrongTarget(t *testing.T) {
	good := connection{Host: "localhost", Port: "5432", Name: databaseName, User: "local-role", SSLMode: "disable"}
	if err := validateConnection(good); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"host", "name", "port"} {
		bad := good
		switch field {
		case "host":
			bad.Host = "db.production.example"
		case "name":
			bad.Name = "postgres"
		case "port":
			bad.Port = "not-a-port"
		}
		if validateConnection(bad) == nil {
			t.Fatalf("accepted unsafe %s", field)
		}
	}
}
func TestRejectsInvalidCatalog(t *testing.T) {
	for _, scenario := range []string{"duplicate-key", "zero-price", "no-description", "no-orderable-items"} {
		var rows []restaurant
		if err := json.Unmarshal(catalogJSON, &rows); err != nil {
			t.Fatal(err)
		}
		switch scenario {
		case "duplicate-key":
			rows[1].Key = rows[0].Key
		case "zero-price":
			rows[0].Items[0].Price = 0
		case "no-description":
			rows[0].Items[0].Description = ""
		case "no-orderable-items":
			for i := range rows[0].Items {
				rows[0].Items[i].Available = false
			}
		}
		if validateCatalog(rows) == nil {
			t.Fatalf("accepted %s", scenario)
		}
	}
}

// Validate the manifest-to-file contract without invoking any database code.
func TestLocalImagesExist(t *testing.T) {
	rows, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, r := range rows {
		paths := []string{r.ImageURL}
		for _, d := range r.Items {
			paths = append(paths, d.ImageURL)
		}
		for _, p := range paths {
			if p == "" {
				continue
			}
			if !validImagePath(p) || seen[p] {
				t.Fatalf("invalid/duplicate local asset: %s", p)
			}
			seen[p] = true
			info, err := os.Stat(filepath.Join("../../frontend/public", strings.TrimPrefix(p, "/")))
			if err != nil || info.Size() == 0 {
				t.Fatalf("missing asset: %s", p)
			}
		}
	}
	if len(seen) != 47 {
		t.Fatalf("expected 47 photographed restaurant/menu image slots, got %d", len(seen))
	}
}

// The additive release must never edit the previously seeded five restaurants.
func TestOriginalFiveCatalogUnchanged(t *testing.T) {
	var rows []json.RawMessage
	if err := json.Unmarshal(catalogJSON, &rows); err != nil {
		t.Fatal(err)
	}
	raw := append([]byte("["), bytes.Join(func() [][]byte {
		r := [][]byte{}
		for _, v := range rows[:5] {
			r = append(r, v)
		}
		return r
	}(), []byte(","))...)
	raw = append(raw, ']')
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(compact.Bytes())) != "cd6ec8fb9f7a41c81cc5e13da74a6fa7fe532efa51ec79ead87574f2e0bc9221" {
		t.Fatal("original five catalogue entries changed; seed would refuse drift")
	}
}
