package models

import "time"

// ORD-01: Daily counter for ORD-YYYYMMDD-NNN
// This makes it safe when 2 orders are placed at same time
type OrderDailySequence struct {
	OrderDate  time.Time `gorm:"type:date;primaryKey" json:"order_date"`
	LastNumber int       `gorm:"not null" json:"last_number"`
}
