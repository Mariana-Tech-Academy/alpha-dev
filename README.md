# FoodHub — alpha-dev

Food ordering platform for the Mariana Tech Academy four-week internship.
Required stack: Go, PostgreSQL, GORM, JWT and React.

## Current contribution

The `models` package defines the eight core database entities, relationships,
indexes, unique constraints, timestamps and basic database check constraints.
It also includes order status constants, a transition helper and a migration
helper. See [database design](docs/database.md) for the ERD and implementation
rules. The API, authentication, frontend and database connection are still to
be implemented. `main.go` is a buildable entry point placeholder.

## Local setup and checks

Install Go 1.22 or newer, then run from the project folder:

```bash
go mod tidy
gofmt -w main.go models
go test ./...
go vet ./...
go build ./...
```

Commit the resulting `go.sum` with the model changes. Tests cover model
relationship parsing, delivery snapshot columns, hidden password hashes and
all order status pairs. These tests do not validate PostgreSQL migrations;
database integration tests are still required.

This contribution was prepared in an environment without Go or PostgreSQL,
so the Go checks and live migrations have not been run here.

## Database setup (next backend task)

The backend team should add `gorm.io/driver/postgres`, open a connection using
an environment variable such as `DATABASE_URL`, and call `models.Migrate(db)`
from explicit development setup code. Never commit credentials. Production
schema changes should use reviewed, versioned migrations. The model helper
does not run automatically.

## Contributing

Use a feature branch such as `feature/database-models`. Push that branch and
open a Pull Request for team review; never commit directly to `main`.
Use request DTOs for input rather than binding requests directly to models:
customers must not be able to assign roles, approval statuses or password hashes.

## Project documentation to complete

Team members, API documentation, backend/frontend run commands, environment
variables and deployment URLs will be added as those features are implemented.
