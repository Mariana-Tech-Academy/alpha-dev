package main

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"net"
	"time"
)

// Validate the actual server response; a configured localhost hostname alone is
// insufficient (for example, it may be a forwarded connection to another host).
func verifyDemoDatabase(db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var identity struct {
		DatabaseName  string
		ServerAddress string
		ClientAddress string
	}
	result := db.WithContext(ctx).Raw(`SELECT current_database() AS database_name,
 host(inet_server_addr()) AS server_address, host(inet_client_addr()) AS client_address`).Scan(&identity)
	if result.Error != nil || result.RowsAffected != 1 {
		return errors.New("could not verify demo database identity; startup stopped")
	}
	return validateDemoDatabaseIdentity(identity.DatabaseName, identity.ServerAddress, identity.ClientAddress)
}
func validateDemoDatabaseIdentity(database, server, client string) error {
	serverIP, clientIP := net.ParseIP(server), net.ParseIP(client)
	if database != "foodhub_new_structure" || serverIP == nil || !serverIP.IsLoopback() || clientIP == nil || !clientIP.IsLoopback() {
		return errors.New("demo safe-start requires a verified loopback connection to foodhub_new_structure; startup stopped")
	}
	return nil
}

// A failed safe-start check must abort, never fall back to migrations/cleanup.
func runStartupTasks(safe bool, verify, migrate func() error, startCleanup func()) error {
	if safe {
		if verify == nil {
			return errors.New("demo safe-start has no database verifier")
		}
		return verify()
	}
	if err := migrate(); err != nil {
		return err
	}
	startCleanup()
	return nil
}
