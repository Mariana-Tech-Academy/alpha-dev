package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"alpha-dev/models"
)

type fakeRevocationStore struct {
	revoked map[string]time.Time
}

func (f *fakeRevocationStore) IsRevoked(_ context.Context, tokenID string) (bool, error) {
	_, ok := f.revoked[tokenID]
	return ok, nil
}

func (f *fakeRevocationStore) Revoke(_ context.Context, tokenID string, expiresAt time.Time) error {
	f.revoked[tokenID] = expiresAt
	return nil
}

func newTestManager() *JWTManager {
	return NewJWTManager("test-secret", time.Hour, &fakeRevocationStore{revoked: map[string]time.Time{}})
}

func TestRevokedTokenIsRejected(t *testing.T) {
	ctx := context.Background()
	m := newTestManager()

	token, err := m.Generate(1, models.RoleCustomer)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Verify(ctx, token)
	if err != nil {
		t.Fatalf("fresh token should verify: %v", err)
	}

	if err := m.Revoke(ctx, claims); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Verify(ctx, token); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("expected ErrTokenRevoked, got %v", err)
	}
}

func TestRevokingOneTokenKeepsOthersValid(t *testing.T) {
	ctx := context.Background()
	m := newTestManager()

	// Issued in the same second for the same user, as with two devices logging in together.
	first, _ := m.Generate(1, models.RoleCustomer)
	second, _ := m.Generate(1, models.RoleCustomer)
	if first == second {
		t.Fatal("tokens issued together must differ")
	}

	claims, err := m.Verify(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Revoke(ctx, claims); err != nil {
		t.Fatal(err)
	}

	if _, err := m.Verify(ctx, second); err != nil {
		t.Fatalf("second token should still be valid: %v", err)
	}
}

func TestInvalidTokens(t *testing.T) {
	ctx := context.Background()
	m := newTestManager()

	other := NewJWTManager("other-secret", time.Hour, &fakeRevocationStore{revoked: map[string]time.Time{}})
	wrongSecret, _ := other.Generate(1, models.RoleCustomer)

	expired := NewJWTManager("test-secret", -time.Minute, &fakeRevocationStore{revoked: map[string]time.Time{}})
	expiredToken, _ := expired.Generate(1, models.RoleCustomer)

	for name, token := range map[string]string{
		"garbage":      "not-a-jwt",
		"wrong secret": wrongSecret,
		"expired":      expiredToken,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := m.Verify(ctx, token); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}
