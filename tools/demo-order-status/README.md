# Local order-status presentation workflow

The My Orders page now includes an opt-in **LOCAL-DEMO-ONLY — Advance Status**
control. It uses the previously implemented staff-authenticated demo route; it never
uses the customer JWT for status writes. Customer-visible labels are:

| Stored status                  | Customer label   |
| ------------------------------ | ---------------- |
| pending / confirmed            | Pending          |
| preparing / ready_for_delivery | In Progress      |
| out_for_delivery               | Out for Delivery |
| delivered                      | Complete         |
| cancelled                      | Cancelled        |

The demo route advances `pending → preparing → out_for_delivery → delivered`.
It also maps legacy confirmed/ready_for_delivery orders to the next visible stage.
The normal manager PATCH endpoint retains its original six-step flow. No schema or
data migration is needed; no existing statuses are rewritten. Checkout still creates
Pending orders. Neither workflow supports new cancellation transitions.

## Enable after review

The owner has confirmed that the existing local FoodHub setup and manager login work.
This is local setup history, not a claim that a fresh checkout contains database data
or credentials. No seeds, restarts or order writes are performed by documentation or
offline tests. Use an existing active manager who owns the selected restaurant; enter
credentials privately in the browser. No passwords or JWTs are distributed with Git.
The browser panel accepts managers only. The backend retains its existing staff policy:
managers are restricted to their restaurant; authorised admins can use the staff API.
Customer credentials never authorise status writes.

For a fresh environment, prepare the matching backend/database separately using the
[seed guide](../seed-foodhub/README.md). The demo flags below are disabled by default.
These are setup instructions, not commands to rerun against the working installation.

For a later approved restart, set these backend process variables:

```sh
APP_ENV=development FOODHUB_DEMO_SAFE_START=1 FOODHUB_DEMO_STATUS=1 go run .
```

**Safe startup:** `FOODHUB_DEMO_SAFE_START` is disabled by default (unset or `0`).
The explicit value `1` requires APP_ENV exactly `development`, DB_NAME exactly
`foodhub_new_structure`, a loopback configured DB_HOST, and no conflicting non-development
environment labels. Invalid flag values/configuration abort startup before opening the
server. After connecting, a five-second read-only query checks `current_database()`,
`inet_server_addr()` and `inet_client_addr()`; both reported IPs must be loopback.
Missing identity, a different database, a remote IP or a query failure aborts startup;
there is no fallback to migration/cleanup.

Only after successful verification does safe-start skip **both** startup AutoMigrate
and the expired-token cleanup goroutine (including its first immediate delete).
JWT validation and token-revocation lookup/recording remain enabled. Authenticated
application actions can still write normally; this flag is not a database read-only
mode and does not grant approval to change passwords or orders.

With the switch unset/`0`, normal startup still runs AutoMigrate and cleanup exactly
as before. No production defaults, schema, accounts or order statuses are changed.
The command above is for a deliberate local startup after configuration review. Use `go run .`, so the new startup helper is
compiled; running only `main.go` omits other files in the main package.

The backend additionally requires `.env` DB_NAME exactly `foodhub_new_structure`,
a loopback DB_HOST, and no production/non-development environment markers. Requests
must arrive from loopback with a loopback Host; forwarded requests are rejected.
The route returns 404 when disabled or outside these local conditions.

For a fresh local development session, start Vite in another terminal:

```sh
cd frontend
VITE_FOODHUB_DEMO_STATUS=1 npm run dev
```

The UI exists only with Vite development mode, explicit frontend opt-in, and loopback
frontend/API hosts. A production build never enables this control, regardless of the
frontend flag. No real `.env` file is modified automatically.

1. Remain logged in as a customer and open My Orders on localhost. Select an eligible order.
2. In the LOCAL-DEMO-ONLY panel, enter the restaurant manager’s email and current
   password. The email is visible; the password is masked. Click **Sign in as manager**.
3. The panel calls `/auth/login` without customer authorisation, then `/auth/me` with
   the returned manager token. It requires matching active manager identities and an
   active restaurant whose `manager_id` matches. Success displays **Manager authenticated**.
   There is no JWT paste field. Credentials clear on submission; the manager JWT stays
   in component-owned memory and never replaces the customer session in browser storage.
4. A login error reports only the failing stage (login request, profile request,
   identity check, restaurant request or ownership check) and a safe HTTP status when
   available. Do not copy passwords, headers, response bodies or network archives.
   Refresh orders to confirm the customer session remains active after manager login.
5. Only when deliberately authorised to change that order, click **Advance Status**.
   Each click rechecks ownership; the backend checks JWT, role, ownership and expected
   status before a conditional update. History refreshes on success. Login alone does
   not advance an order. Terminal orders remain in history but leave the selector.
6. **Manager logout**, cancellation, a full page reload, unmount or a changed customer
   session discards the manager session. Staff 401/403 errors clear only staff credentials.
   Logout does not revoke the in-memory JWT on the server; normal expiry/revocation
   rules continue to apply. Cancellation cannot undo a write already accepted by the API.

`POST /api/demo/orders/{id}/advance` accepts only
`{"expected_status":"pending"}` (or the currently displayed native status).
It rejects a desired target, stale status, invalid status, customer JWT, missing,
expired/revoked JWT, and a different restaurant's manager. A conditional update
prevents concurrent requests from advancing the same observed stage twice. No timer
or checkout hook advances orders. Remove both opt-in flags after the presentation.

## Verification without database writes

Automated tests cover real JWT middleware/route/service flow with an isolated in-memory
repository, customer history readback, ownership, terminal/invalid/stale transitions,
unchanged snapshots, and the production/local gates. Scripted SQL tests exercise the
real GORM conditional UPDATE/readback for each persisted stage, without PostgreSQL.
The existing PostgreSQL-writing snapshot integration test must not run without
separate approval; all other Go tests and vet can run safely.

```sh
go test ./... -skip '^TestOrderSnapshot$'
go vet ./...
cd frontend
npm run lint
npm run build
npm test
```

## Optional original terminal workflow

This optional terminal control reuses `GET /api/auth/me` and the existing authenticated
`PATCH /api/orders/{id}/status` route. It adds no server endpoint and does not connect
to PostgreSQL directly. It is disabled by default, rejects non-development environment
labels (including those in the root `.env`), and contacts only `127.0.0.1` without
following redirects. Do not enable or distribute this utility in production.
The existing staff API remains protected by JWT, manager/admin roles, restaurant
ownership for managers, next-step validation and a conditional status update.

**Legacy terminal alternative:** this script still accepts a privately obtained staff
JWT; the browser workflow above uses email/password and is the current demo interface.
Using this script changes a real local order and requires deliberate authorisation.
An existing manager/admin account and valid JWT are required. It provides no accounts,
role changes or token minting. An ordinary customer token is rejected.

After approval, from the repository root, with the local API running:

```sh
APP_ENV=development FOODHUB_DEMO_STATUS=1 python3 tools/demo-order-status/advance.py ORDER_ID confirmed
```

Replace `ORDER_ID` with a real order ID. Paste a staff JWT at the hidden prompt;
never put tokens in shell arguments or commit them. Confirm with `ADVANCE`.
The command contacts only the local API on port 8080 (`--port` can override it).

Advance one native step at a time:

`pending → confirmed → preparing → ready_for_delivery → out_for_delivery → delivered`

Use the desired next status as the second argument, then refresh My Orders.
Skipping/reversing steps fails. Checkout continues to create Pending orders.
`cancelled` is a valid backend enum and displays correctly in history, but the current
service does not allow transitioning to it. This utility does not enable cancellation.
No schema changes, migration, scheduler, automatic progression or fake delivery state
are required. The environment checks are an operator safeguard; server JWT/ownership
checks are the authorization boundary. Never forward a production API through this local port.

Offline tests (mock HTTP only, no database writes):

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tools/demo-order-status -p 'test_*.py'
```
