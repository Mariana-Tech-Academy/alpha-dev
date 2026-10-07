package models

import "time"

// Record supplies the primary key and GORM-managed timestamps.
type Record struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// Money stores minor currency units (100 = $1.00), avoiding floating point rounding.
type Money int64
