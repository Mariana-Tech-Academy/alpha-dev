package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"alpha-dev/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	if err := godotenv.Load("../.env"); err != nil {
		t.Skipf("order snapshot test needs ../.env: %v", err)
	}

	const dbName = "foodhub_test"
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		dbName,
		os.Getenv("SSL_MODE"),
	)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skipf("foodhub_test is not available: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Skipf("foodhub_test is not available: %v", err)
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		t.Skipf("foodhub_test is not available: %v", err)
	}

	var connectedDB string
	if err := db.Raw("SELECT current_database()").Scan(&connectedDB).Error; err != nil {
		t.Fatalf("failed to verify database: %v", err)
	}
	if connectedDB != dbName {
		t.Fatalf("unsafe database connection: %s", connectedDB)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Restaurant{},
		&models.Category{},
		&models.MenuItem{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderDailySequence{},
		&models.RevokedToken{},
	); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

func TestOrderSnapshot(t *testing.T) {
	db := setupOrderTestDB(t)

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to start test transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		if err := tx.Rollback().Error; err != nil {
			t.Errorf("failed to roll back test transaction: %v", err)
		}
	})
	db = tx

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	customer := models.User{
		FirstName: "Test",
		LastName:  "Customer",
		Email:     "snapshot.customer." + suffix + "@test.local",
		Phone:     "07123456789",
		Password:  "test-password",
		Role:      models.RoleCustomer,
	}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("failed to create test customer: %v", err)
	}

	manager := models.User{
		FirstName: "Test",
		LastName:  "Manager",
		Email:     "snapshot.manager." + suffix + "@test.local",
		Phone:     "07987654321",
		Password:  "test-password",
		Role:      models.RoleManager,
	}
	if err := db.Create(&manager).Error; err != nil {
		t.Fatalf("failed to create test manager: %v", err)
	}

	restaurant := models.Restaurant{
		ManagerID: manager.ID,
		Name:      "Snapshot Test Restaurant",
		Address:   "123 Test Street",
		City:      "Manchester",
		Phone:     "01611234567",
		Email:     "restaurant." + suffix + "@test.local",
		Status:    models.RestaurantStatusActive,
	}
	if err := db.Create(&restaurant).Error; err != nil {
		t.Fatalf("failed to create test restaurant: %v", err)
	}

	category := models.Category{Name: "Burgers " + suffix}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}

	menuItem := models.MenuItem{
		RestaurantID: restaurant.ID,
		CategoryID:   category.ID,
		Name:         "Chicken Burger",
		Description:  "Grilled chicken burger",
		Price:        850,
		Available:    true,
	}
	if err := db.Create(&menuItem).Error; err != nil {
		t.Fatalf("failed to create test menu item: %v", err)
	}

	restaurantID := restaurant.ID
	cart := models.Cart{UserID: customer.ID, RestaurantID: &restaurantID}
	if err := db.Create(&cart).Error; err != nil {
		t.Fatalf("failed to create test cart: %v", err)
	}

	cartItem := models.CartItem{
		CartID:     cart.ID,
		MenuItemID: menuItem.ID,
		Quantity:   2,
		Price:      850,
	}
	if err := db.Create(&cartItem).Error; err != nil {
		t.Fatalf("failed to add item to cart: %v", err)
	}

	repo := NewOrderRepository(db)
	placedOrder, err := repo.PlaceOrder(context.Background(), customer.ID, &models.Order{
		DeliveryName:    "Test Customer",
		DeliveryPhone:   "07123456789",
		DeliveryAddress: "123 Original Street",
		DeliveryCity:    "Manchester",
		DeliveryState:   "England",
		DeliveryFee:     250,
	})
	if err != nil {
		t.Fatalf("failed to place test order: %v", err)
	}
	if placedOrder.Subtotal != 1700 || placedOrder.DeliveryFee != 0 || placedOrder.Total != 1700 {
		t.Fatalf("expected subtotal 1700, fee 0, total 1700; got %d, %d, %d", placedOrder.Subtotal, placedOrder.DeliveryFee, placedOrder.Total)
	}

	if err := db.Model(&customer).Updates(map[string]any{
		"first_name": "Updated",
		"last_name":  "Customer",
		"phone":      "07999999999",
	}).Error; err != nil {
		t.Fatalf("failed to update customer profile: %v", err)
	}

	saved := reloadOrder(t, db, placedOrder.ID)
	if saved.DeliveryName != "Test Customer" || saved.DeliveryPhone != "07123456789" || saved.DeliveryAddress != "123 Original Street" || saved.DeliveryCity != "Manchester" || saved.DeliveryState != "England" {
		t.Fatalf("delivery snapshot changed after profile update: %+v", saved)
	}
	assertSnapshotItem(t, saved, "Chicken Burger", 850, 2, 1700)
	if saved.Total != 1700 {
		t.Fatalf("order total changed after profile update: %d", saved.Total)
	}

	if err := db.Model(&menuItem).Updates(map[string]any{
		"name":  "Premium Chicken Burger",
		"price": int64(1000),
	}).Error; err != nil {
		t.Fatalf("failed to update menu item: %v", err)
	}
	saved = reloadOrder(t, db, placedOrder.ID)
	assertSnapshotItem(t, saved, "Chicken Burger", 850, 2, 1700)
	if saved.Total != 1700 {
		t.Fatalf("order total changed after menu update: %d", saved.Total)
	}

	if err := db.Delete(&menuItem).Error; err != nil {
		t.Fatalf("failed to delete menu item: %v", err)
	}
	var deleted models.MenuItem
	err = db.First(&deleted, menuItem.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected menu item to be soft-deleted, got: %v", err)
	}
	saved = reloadOrder(t, db, placedOrder.ID)
	assertSnapshotItem(t, saved, "Chicken Burger", 850, 2, 1700)
	if saved.Total != 1700 || saved.DeliveryFee != 0 {
		t.Fatalf("order totals changed after menu deletion: fee %d total %d", saved.DeliveryFee, saved.Total)
	}
}

func reloadOrder(t *testing.T, db *gorm.DB, id uint) models.Order {
	t.Helper()
	var order models.Order
	if err := db.Preload("Items").First(&order, id).Error; err != nil {
		t.Fatalf("failed to retrieve order: %v", err)
	}
	return order
}

func assertSnapshotItem(t *testing.T, order models.Order, name string, unitPrice int64, quantity int, lineTotal int64) {
	t.Helper()
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(order.Items))
	}
	item := order.Items[0]
	if item.Name != name || item.UnitPrice != unitPrice || item.Quantity != quantity || item.LineTotal != lineTotal {
		t.Fatalf("order item snapshot = name %q price %d qty %d line %d", item.Name, item.UnitPrice, item.Quantity, item.LineTotal)
	}
}
