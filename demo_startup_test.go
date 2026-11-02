package main

import (
	"alpha-dev/auth"
	"alpha-dev/models"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestDemoDatabaseIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, db, server, client string
		valid                    bool
	}{
		{"IPv4", "foodhub_new_structure", "127.0.0.1", "127.0.0.1", true},
		{"IPv6", "foodhub_new_structure", "::1", "::1", true},
		{"wrong database", "foodhub_production", "127.0.0.1", "127.0.0.1", false},
		{"remote server", "foodhub_new_structure", "192.0.2.3", "127.0.0.1", false},
		{"remote client", "foodhub_new_structure", "127.0.0.1", "192.0.2.3", false},
		{"missing identity", "", "", "", false},
		{"Unix socket has no verified IP", "foodhub_new_structure", "", "", false},
		{"hostname is not verified IP", "foodhub_new_structure", "localhost", "127.0.0.1", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateDemoDatabaseIdentity(tc.db, tc.server, tc.client); (err == nil) != tc.valid {
				t.Fatalf("error %v", err)
			}
		})
	}
}
func TestStartupTasks(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		safe, verifyFailure, migrateFailure bool
		want                                []string
		wantError                           bool
	}{
		{name: "normal startup unchanged", want: []string{"migrate", "cleanup"}},
		{name: "safe start skips both write tasks", safe: true, want: []string{"verify"}},
		{name: "failed verification never falls back", safe: true, verifyFailure: true, want: []string{"verify"}, wantError: true},
		{name: "normal migration failure stops cleanup", migrateFailure: true, want: []string{"migrate"}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := []string{}
			err := runStartupTasks(tc.safe, func() error {
				calls = append(calls, "verify")
				if tc.verifyFailure {
					return errors.New("verification failed")
				}
				return nil
			}, func() error {
				calls = append(calls, "migrate")
				if tc.migrateFailure {
					return errors.New("migration failed")
				}
				return nil
			}, func() { calls = append(calls, "cleanup") })
			if (err != nil) != tc.wantError || !reflect.DeepEqual(calls, tc.want) {
				t.Fatalf("calls %v err %v", calls, err)
			}
		})
	}
	t.Run("missing verifier fails closed", func(t *testing.T) {
		touched := false
		err := runStartupTasks(true, nil, func() error { touched = true; return nil }, func() { touched = true })
		if err == nil || touched {
			t.Fatal("missing verifier did not stop startup")
		}
	})
}

type safeStartRevocations struct {
	revoked bool
	checks  int
}

func (s *safeStartRevocations) IsRevoked(context.Context, string) (bool, error) {
	s.checks++
	return s.revoked, nil
}
func (s *safeStartRevocations) Revoke(context.Context, string, time.Time) error {
	s.revoked = true
	return nil
}
func TestSafeStartKeepsJWTAndRevocationEnforcement(t *testing.T) {
	store := &safeStartRevocations{}
	err := runStartupTasks(true, func() error { return nil }, func() error { t.Fatal("migration ran"); return nil }, func() { t.Fatal("cleanup ran") })
	if err != nil {
		t.Fatal(err)
	}
	tokens := auth.NewJWTManager("test-only-safe-start-secret", time.Hour, store)
	if _, err = tokens.Verify(context.Background(), "invalid"); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatal("invalid JWT accepted")
	}
	token, err := tokens.Generate(4, models.RoleManager)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tokens.Verify(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if err = tokens.Revoke(context.Background(), claims); err != nil {
		t.Fatal(err)
	}
	if _, err = tokens.Verify(context.Background(), token); !errors.Is(err, auth.ErrTokenRevoked) {
		t.Fatal("revoked JWT accepted")
	}
	if store.checks != 2 {
		t.Fatal("revocation checks were bypassed")
	}
}
