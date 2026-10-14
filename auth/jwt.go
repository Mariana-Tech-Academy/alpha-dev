package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"alpha-dev/models"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type contextKey string

const claimsKey contextKey = "claims"

var ErrTokenRevoked = errors.New("token has been revoked")

type Claims struct {
	UserID uint        `json:"user_id"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	expiry time.Duration
	db     *gorm.DB
}

func NewJWTManager(
	secret string,
	expiry time.Duration,
	db *gorm.DB,
) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		expiry: expiry,
		db:     db,
	}
}

func (m *JWTManager) Generate(
	userID uint,
	role models.Role,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
		},
	}

	return jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	).SignedString(m.secret)
}

func tokenHash(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

func (m *JWTManager) Verify(tokenString string) (*Claims, error) {
	return m.verify(context.Background(), tokenString)
}

func (m *JWTManager) verify(
	ctx context.Context,
	tokenString string,
) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(_ *jwt.Token) (any, error) {
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	var revoked models.RevokedToken

	err = m.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash(tokenString)).
		First(&revoked).Error

	if err == nil {
		return nil, ErrTokenRevoked
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return claims, nil
}

func (m *JWTManager) Revoke(
	ctx context.Context,
	tokenString string,
) error {
	claims, err := m.verify(ctx, tokenString)
	if errors.Is(err, ErrTokenRevoked) {
		return nil
	}
	if err != nil {
		return err
	}

	revoked := models.RevokedToken{
		TokenHash: tokenHash(tokenString),
		ExpiresAt: claims.ExpiresAt.Time,
	}

	return m.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&revoked).Error
}

func WithClaims(
	ctx context.Context,
	claims *Claims,
) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext returns the authenticated user's claims.
func ClaimsFromContext(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("no auth claims in context")
	}

	return claims, nil
}
