package models

// Categories are shared platform-wide (for example Pizza, Drinks, Desserts).
type Category struct {
	Record
	Name      string     `gorm:"size:100;not null;uniqueIndex" json:"name"`
	MenuItems []MenuItem `gorm:"foreignKey:CategoryID" json:"-"`
}
