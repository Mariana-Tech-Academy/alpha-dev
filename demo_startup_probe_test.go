package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"strings"
	"testing"
)

// Read-only scripted connection: no real PostgreSQL access or startup side effects.
type demoProbe struct {
	database, server, client string
	fail, empty              bool
	queries                  int
}

func (p *demoProbe) Driver() driver.Driver                        { return p }
func (p *demoProbe) Open(string) (driver.Conn, error)             { return p, nil }
func (p *demoProbe) Connect(context.Context) (driver.Conn, error) { return p, nil }
func (p *demoProbe) Close() error                                 { return nil }
func (p *demoProbe) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (p *demoProbe) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (p *demoProbe) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	p.queries++
	if p.fail {
		return nil, errors.New("verification query failed")
	}
	if !strings.HasPrefix(query, "SELECT current_database()") || !strings.Contains(query, "inet_server_addr()") || !strings.Contains(query, "inet_client_addr()") || len(args) != 0 {
		return nil, errors.New("unexpected query")
	}
	return &demoProbeRows{probe: p}, nil
}

type demoProbeRows struct {
	probe *demoProbe
	read  bool
}

func (r *demoProbeRows) Columns() []string {
	return []string{"database_name", "server_address", "client_address"}
}
func (r *demoProbeRows) Close() error { return nil }
func (r *demoProbeRows) Next(values []driver.Value) error {
	if r.read || r.probe.empty {
		return io.EOF
	}
	r.read = true
	values[0], values[1], values[2] = r.probe.database, r.probe.server, r.probe.client
	return nil
}
func TestVerifyDemoDatabaseReadOnlyProbe(t *testing.T) {
	for _, tc := range []struct {
		name  string
		probe demoProbe
		valid bool
	}{
		{"verified IPv4", demoProbe{database: "foodhub_new_structure", server: "127.0.0.1", client: "127.0.0.1"}, true},
		{"verified IPv6", demoProbe{database: "foodhub_new_structure", server: "::1", client: "::1"}, true},
		{"wrong actual database", demoProbe{database: "other", server: "127.0.0.1", client: "127.0.0.1"}, false},
		{"actual remote server", demoProbe{database: "foodhub_new_structure", server: "192.0.2.3", client: "127.0.0.1"}, false},
		{"unverified response", demoProbe{empty: true}, false},
		{"query failure", demoProbe{fail: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn := sql.OpenDB(&tc.probe)
			defer conn.Close()
			db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			err = verifyDemoDatabase(db)
			if (err == nil) != tc.valid || tc.probe.queries != 1 {
				t.Fatalf("error=%v queries=%d", err, tc.probe.queries)
			}
		})
	}
}
