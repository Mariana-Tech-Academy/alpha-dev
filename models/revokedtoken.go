package models

import "time"

type RevokedToken struct {
	TokenHash string    `gorm:"primaryKey;size:64"`
	ExpiresAt time.Time `gorm:"not null;index"`
}
