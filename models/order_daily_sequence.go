package models

import "time"

// OrderDailySequence is the per-day counter behind ORD-YYYYMMDD-NNN.
type OrderDailySequence struct {
	OrderDate  time.Time `gorm:"type:date;primaryKey"`
	LastNumber int       `gorm:"not null"`
}
