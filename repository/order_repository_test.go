package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"alpha-dev/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Load existing connection settings.
	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}

	// Always connect to the dedicated test database.
	dbName := "foodhub_test"

	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		dbName,
		os.Getenv("SSL_MODE"),
	)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Verify the actual database connection before any writes.
	var connectedDB string

	if err := db.Raw("SELECT current_database()").Scan(&connectedDB).Error; err != nil {
		t.Fatalf("failed to verify database: %v", err)
	}

	if connectedDB != "foodhub_test" {
		t.Fatalf("unsafe database connection: %s", connectedDB)
	}
	// Create the required tables in the test database.
	err = db.AutoMigrate(
		&models.User{},
		&models.Restaurant{},
		&models.Category{},
		&models.MenuItem{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.RevokedToken{},
	)

	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

func TestOrderDatabaseConnection(t *testing.T) {
	db := setupOrderTestDB(t)

	var databaseName string

	err := db.Raw("SELECT current_database()").Scan(&databaseName).Error
	if err != nil {
		t.Fatalf("failed to query database: %v", err)
	}

	if databaseName != "foodhub_test" {
		t.Fatalf("expected foodhub_test, got %s", databaseName)
	}

	t.Logf("Successfully connected to %s", databaseName)
}

func TestOrderSnapshot(t *testing.T) {
	db := setupOrderTestDB(t)

	// Start an isolated test transaction.
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to start test transaction: %v", tx.Error)
	}

	// Roll back all test changes when the test finishes.
	t.Cleanup(func() {
		if err := tx.Rollback().Error; err != nil {
			t.Errorf("failed to roll back test transaction: %v", err)
		}
	})

	// Use the transaction for all test database operations.
	db = tx

	// Step 1: Create test records.
	// Create a test customer.
	customer := models.User{
		FirstName: "Test",
		LastName:  "Customer",
		Email:     "snapshot.customer@test.local",
		Phone:     "07123456789",
		Password:  "test-password",
		Role:      models.Role("customer"),
	}

	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("failed to create test customer: %v", err)
	}

	// Create a test restaurant manager.
	manager := models.User{
		FirstName: "Test",
		LastName:  "Manager",
		Email:     "snapshot.manager@test.local",
		Phone:     "07987654321",
		Password:  "test-password",
		Role:      models.Role("manager"),
	}

	if err := db.Create(&manager).Error; err != nil {
		t.Fatalf("failed to create test manager: %v", err)
	}

	// Create an active test restaurant.
	restaurant := models.Restaurant{
		ManagerID: manager.ID,
		Name:      "Snapshot Test Restaurant",
		Address:   "123 Test Street",
		City:      "Manchester",
		Phone:     "01611234567",
		Email:     "restaurant@test.local",
		Status:    models.RestaurantStatusActive,
	}

	if err := db.Create(&restaurant).Error; err != nil {
		t.Fatalf("failed to create test restaurant: %v", err)
	}

	// Create a category for the test restaurant.
	category := models.Category{
		Name: "Burgers",
	}

	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}

	// Create a test menu item.
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

	// Create the customer's shopping cart.
	restaurantID := restaurant.ID

	cart := models.Cart{
		UserID:       customer.ID,
		RestaurantID: &restaurantID,
	}

	if err := db.Create(&cart).Error; err != nil {
		t.Fatalf("failed to create test cart: %v", err)
	}

	// Add two chicken burgers to the cart.
	cartItem := models.CartItem{
		CartID:     cart.ID,
		MenuItemID: menuItem.ID,
		Quantity:   2,
	}

	if err := db.Create(&cartItem).Error; err != nil {
		t.Fatalf("failed to add item to cart: %v", err)
	}

	// Step 2: Place an order.
	// Create the order with delivery details.
	order := &models.Order{
		DeliveryName:    "Test Customer",
		DeliveryPhone:   "07123456789",
		DeliveryAddress: "123 Original Street",
		DeliveryCity:    "Manchester",
		DeliveryFee:     250,
	}

	// Create the repository using our test transaction.
	repo := NewOrderRepository(db)

	// Place the customer's order.
	placedOrder, err := repo.PlaceOrder(
		context.Background(),
		customer.ID,
		order,
	)

	if err != nil {
		t.Fatalf("failed to place test order: %v", err)
	}

	// Check the order totals.
	if placedOrder.Subtotal != 1700 {
		t.Errorf("expected subtotal 1700, got %d", placedOrder.Subtotal)
	}

	if placedOrder.Total != 1950 {
		t.Errorf("expected total 1950, got %d", placedOrder.Total)
	}

	// Step 3: Change the customer's profile.
	// Update the customer's profile after checkout.
	if err := db.Model(&customer).Updates(map[string]interface{}{
		"first_name": "Updated",
		"last_name":  "Customer",
		"phone":      "07999999999",
	}).Error; err != nil {
		t.Fatalf("failed to update customer profile: %v", err)
	}

	// Retrieve the historical order again.
	var orderAfterProfileUpdate models.Order

	if err := db.First(
		&orderAfterProfileUpdate,
		placedOrder.ID,
	).Error; err != nil {
		t.Fatalf("failed to retrieve historical order: %v", err)
	}

	// Verify the original delivery information is unchanged.
	if orderAfterProfileUpdate.DeliveryName != "Test Customer" {
		t.Errorf("delivery name changed after profile update")
	}

	if orderAfterProfileUpdate.DeliveryPhone != "07123456789" {
		t.Errorf("delivery phone changed after profile update")
	}

	if orderAfterProfileUpdate.DeliveryAddress != "123 Original Street" {
		t.Errorf("delivery address changed after profile update")
	}

	if orderAfterProfileUpdate.DeliveryCity != "Manchester" {
		t.Errorf("delivery city changed after profile update")
	}
	// Retrieve the saved order and its items from PostgreSQL.
	var savedOrder models.Order

	if err := db.Preload("Items").
		First(&savedOrder, placedOrder.ID).Error; err != nil {
		t.Fatalf("failed to retrieve saved order: %v", err)
	}

	// Verify the original delivery information.
	if savedOrder.DeliveryName != "Test Customer" {
		t.Errorf("incorrect delivery name: %s", savedOrder.DeliveryName)
	}

	if savedOrder.DeliveryPhone != "07123456789" {
		t.Errorf("incorrect delivery phone: %s", savedOrder.DeliveryPhone)
	}

	if savedOrder.DeliveryAddress != "123 Original Street" {
		t.Errorf("incorrect delivery address: %s", savedOrder.DeliveryAddress)
	}

	// Verify the order contains one item.
	if len(savedOrder.Items) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(savedOrder.Items))
	}

	// Verify the original menu item snapshot.
	savedItem := savedOrder.Items[0]

	if savedItem.Name != "Chicken Burger" {
		t.Errorf("expected Chicken Burger, got %s", savedItem.Name)
	}

	if savedItem.UnitPrice != 850 {
		t.Errorf("expected unit price 850, got %d", savedItem.UnitPrice)
	}

	if savedItem.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", savedItem.Quantity)
	}

	if savedItem.LineTotal != 1700 {
		t.Errorf("expected line total 1700, got %d", savedItem.LineTotal)
	}

	// Step 4: Change the menu item's name and price.
	// Update the menu item after the customer has placed their order.
	if err := db.Model(&menuItem).Updates(map[string]interface{}{
		"name":  "Premium Chicken Burger",
		"price": int64(1000),
	}).Error; err != nil {
		t.Fatalf("failed to update menu item: %v", err)
	}

	// Retrieve the historical order and its items again.
	var orderAfterMenuUpdate models.Order

	if err := db.Preload("Items").
		First(&orderAfterMenuUpdate, placedOrder.ID).Error; err != nil {
		t.Fatalf("failed to retrieve order after menu update: %v", err)
	}

	// Verify the original item name remains unchanged.
	if len(orderAfterMenuUpdate.Items) != 1 {
		t.Fatalf("expected 1 historical order item, got %d",
			len(orderAfterMenuUpdate.Items))
	}

	historicalItem := orderAfterMenuUpdate.Items[0]

	if historicalItem.Name != "Chicken Burger" {
		t.Errorf("historical item name changed: %s", historicalItem.Name)
	}

	// Verify the original unit price remains unchanged.
	if historicalItem.UnitPrice != 850 {
		t.Errorf("historical unit price changed: %d", historicalItem.UnitPrice)
	}

	// Verify the original line total remains unchanged.
	if historicalItem.LineTotal != 1700 {
		t.Errorf("historical line total changed: %d", historicalItem.LineTotal)
	}

	// Verify the original order total remains unchanged.
	if orderAfterMenuUpdate.Total != 1950 {
		t.Errorf("historical order total changed: %d", orderAfterMenuUpdate.Total)
	}

	// Step 5: Verify the original order information is unchanged.
	// Step 6: Soft-delete the menu item and verify order history.
	// Soft-delete the menu item after checkout.
	if err := db.Delete(&menuItem).Error; err != nil {
		t.Fatalf("failed to delete menu item: %v", err)
	}

	// Verify the menu item is no longer returned by normal queries.
	var deletedMenuItem models.MenuItem

	err = db.First(&deletedMenuItem, menuItem.ID).Error

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected menu item to be soft-deleted, got: %v", err)
	}

	// Retrieve the historical order after menu item deletion.
	var orderAfterDeletion models.Order

	if err := db.Preload("Items").
		First(&orderAfterDeletion, placedOrder.ID).Error; err != nil {
		t.Fatalf("failed to retrieve order after deletion: %v", err)
	}

	// Verify the historical order still contains its item.
	if len(orderAfterDeletion.Items) != 1 {
		t.Fatalf("expected 1 historical item, got %d",
			len(orderAfterDeletion.Items))
	}

	deletedItemSnapshot := orderAfterDeletion.Items[0]

	// Verify the original item name and price remain unchanged.
	if deletedItemSnapshot.Name != "Chicken Burger" {
		t.Errorf("historical item name changed after deletion")
	}

	if deletedItemSnapshot.UnitPrice != 850 {
		t.Errorf("historical item price changed after deletion")
	}

	if deletedItemSnapshot.LineTotal != 1700 {
		t.Errorf("historical line total changed after deletion")
	}

	// Verify the historical order total remains unchanged.
	if orderAfterDeletion.Total != 1950 {
		t.Errorf("historical order total changed after deletion")
	}
}
