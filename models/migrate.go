package models

import "gorm.io/gorm"

// Migrate creates the model tables on an already-open PostgreSQL connection.
// It does not connect to a database or load secrets.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{}, &Restaurant{}, &Category{}, &MenuItem{},
		&Cart{}, &CartItem{}, &Order{}, &OrderItem{},
	)
}
