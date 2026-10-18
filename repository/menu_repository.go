package repository

import( 
	"gorm.io/gorm"

	"context"

	"alpha-dev/models"
)

//ctx is short for context. It’s a small object Go passes along with a request. It carries two main things: a cancel signal and an optional deadline.
//context.Context - built in Go Package and a type
type MenuRepository interface {
FindRestaurantByID(ctx context.Context, id uint) (models.Restaurant, error)
CreateMenuItem(ctx context.Context, item *models.MenuItem) error
}

//you can’t add items to a restaurant you don’t manage. To check that, the service needs the restaurant’s ManagerID, and this function fetches it.
func (repo *PostgresMenuRepo) FindRestaurantByID(ctx context.Context, id uint) (models.Restaurant, error) {
	//repo.db is your GORM database connection. .WithContext(ctx) tells GORM: “run the next query using this context.” If the request is cancelled or times out, GORM stops the query.
	var restaurant models.Restaurant
	err := repo.db.WithContext(ctx).First(&restaurant, id).Error
	return restaurant, err
}

//.Create(item) inserts the menu item as a new row in the database.
func (repo *PostgresMenuRepo) CreateMenuItem(ctx context.Context, item *models.MenuItem) error {
	return repo.db.WithContext(ctx).Create(item).Error
}

//holds database connection
type PostgresMenuRepo struct {
	db *gorm.DB
}

//builds new repository
//repository is a piece of code that talks to database
//here it finds a restaurant and saves menu item
//essentiually it fetches and saves data
//The left db is the name of the slot inside the struct.
// The right db is the value to put in that slot. 
//It’s the database connection that was passed into the function.
func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &PostgresMenuRepo{db: db}
}
