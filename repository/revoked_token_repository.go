package repository

import (
	"context"
	"time"

	"alpha-dev/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RevokedTokenRepository interface {
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
	Revoke(ctx context.Context, tokenID string, expiresAt time.Time) error
	DeleteExpired(ctx context.Context) error
}

type PostgresRevokedTokenRepo struct {
	db *gorm.DB
}

func NewRevokedTokenRepository(db *gorm.DB) RevokedTokenRepository {
	return &PostgresRevokedTokenRepo{db: db}
}

func (repo *PostgresRevokedTokenRepo) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	var count int64
	err := repo.db.WithContext(ctx).
		Model(&models.RevokedToken{}).
		Where("token_id = ?", tokenID).
		Count(&count).Error
	return count > 0, err
}

func (repo *PostgresRevokedTokenRepo) Revoke(ctx context.Context, tokenID string, expiresAt time.Time) error {
	return repo.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.RevokedToken{TokenID: tokenID, ExpiresAt: expiresAt}).Error
}

// DeleteExpired removes rows for tokens that have expired, since those are rejected on expiry alone.
func (repo *PostgresRevokedTokenRepo) DeleteExpired(ctx context.Context) error {
	return repo.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.RevokedToken{}).Error
}
