package models

type Restaurant struct {
	Record
	ManagerID   uint             `gorm:"not null;index" json:"manager_id"`
	Manager     User             `gorm:"foreignKey:ManagerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Name        string           `gorm:"size:150;not null;index" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	Address     string           `gorm:"type:text;not null" json:"address"`
	City        string           `gorm:"size:100;not null;index" json:"city"`
	State       string           `gorm:"size:100;index" json:"state"`
	Phone       string           `gorm:"size:30;not null" json:"phone"`
	Email       string           `gorm:"size:254;not null" json:"email"`
	ImageURL    string           `gorm:"type:text" json:"image_url"`
	OpeningTime string           `gorm:"type:time;not null" json:"opening_time"`
	ClosingTime string           `gorm:"type:time;not null" json:"closing_time"`
	Status      RestaurantStatus `gorm:"size:30;not null;default:pending_approval;index;check:chk_restaurant_status,status IN ('pending_approval','active','suspended')" json:"status"`
	// A manager can toggle this flag, but cannot override approval/suspension.
	IsAcceptingOrders bool       `gorm:"not null;default:false" json:"is_accepting_orders"`
	MenuItems         []MenuItem `gorm:"foreignKey:RestaurantID" json:"-"`
	Orders            []Order    `gorm:"foreignKey:RestaurantID" json:"-"`
}
