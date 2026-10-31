package repository

import (
	"gorm.io/gorm"
	"context"

	"alpha-dev/models"
)


// Checkout must create the order, create its items and clear the cart inside one
// r.db.Transaction(func(tx *gorm.DB) error { ... }) call.
//The service depends on this interface
type OrderRepository interface {
	FindByID(ctx context.Context, id uint) (models.Order, error)
	UpdateStatus(ctx context.Context, id uint, current, next models.OrderStatus) (bool, error)
}



type PostgresOrderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) FindByID(ctx context.Context, id uint) (models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Preload("Restaurant"). //runs second query to fill order.restaurant

		//WHERE id = 7 finds the row with that ID.
       // ORDER BY id LIMIT 1 means “give me only one row”.
       // &order is a pointer to your empty order variable. GORM fills that variable with the row’s data (status, restaurant ID, and so on). Without the &, GORM would fill a copy, and your original would stay empty.
		First(&order, id).Error
	return order, err

	//First(&order, id) means “find the first order whose ID is id, and put it into order”.

	//It looks up one order by its ID in the database, puts the result into your order variable, and gives you back any error that happened.
}

func (r *PostgresOrderRepo) UpdateStatus(ctx context.Context, id uint, current, next models.OrderStatus) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Order{}). //what table to look into
		Where("id = ? AND status = ?", id, current). //uses ? placeholders, so values are sent as parameters.
		Update("status", next) //"status" is the column name.
		// next is the new value, for example confirmed.

	return result.RowsAffected == 1, result.Error
	//So Error and RowsAffected are just fields of that struct, the same as order.Status is a field of your Order, the fields come from the db struct and we store them in result variable
}