package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"alpha-dev/models"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsKey contextKey = "claims"

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrTokenRevoked = errors.New("token has been revoked")
)

type Claims struct {
	UserID uint        `json:"user_id"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

// RevocationStore records tokens that were revoked before they expired.
type RevocationStore interface {
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
	Revoke(ctx context.Context, tokenID string, expiresAt time.Time) error
}

type JWTManager struct {
	secret      []byte
	expiry      time.Duration
	revocations RevocationStore
}

func NewJWTManager(secret string, expiry time.Duration, revocations RevocationStore) *JWTManager {
	return &JWTManager{secret: []byte(secret), expiry: expiry, revocations: revocations}
}

func newTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (m *JWTManager) Generate(userID uint, role models.Role) (string, error) {
	id, err := newTokenID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Verify returns ErrInvalidToken or ErrTokenRevoked for tokens the client should not use;
// any other error means the revocation check itself failed.
func (m *JWTManager) Verify(ctx context.Context, tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(*jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if claims.ID == "" {
		return nil, fmt.Errorf("%w: missing token id", ErrInvalidToken)
	}

	revoked, err := m.revocations.IsRevoked(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("check token revocation: %w", err)
	}
	if revoked {
		return nil, ErrTokenRevoked
	}
	return claims, nil
}

func (m *JWTManager) Revoke(ctx context.Context, claims *Claims) error {
	if claims == nil || claims.ID == "" || claims.ExpiresAt == nil {
		return ErrInvalidToken
	}
	return m.revocations.Revoke(ctx, claims.ID, claims.ExpiresAt.Time)
}

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext returns the authenticated user's claims set by middleware.Authenticate.
func ClaimsFromContext(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("no auth claims in context")
	}
	return claims, nil
}
