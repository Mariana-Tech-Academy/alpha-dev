# FoodHub customer frontend

React + TypeScript + Vite customer demo connected to the existing Go API. No mock restaurant or order data is used.

## Run

Use Node.js 22.18+ (Node 24 recommended).

```sh
cd frontend
npm install
npm run dev
```

Open http://localhost:5173. The backend must already be running separately at http://localhost:8080; these commands do not start it or change its database. Vite uses port 5173 strictly to match the backend's default CORS origin.

The API defaults to `http://localhost:8080/api`. To override, copy `.env.example` to `.env.local`, change `VITE_API_BASE_URL`, and restart Vite. This is a public browser configuration value; do not put secrets in Vite variables. The backend must allow the frontend origin in its CORS configuration. Use localhost rather than 127.0.0.1 with default backend CORS settings.

```sh
npm run lint
npm test
npm run build
npm run preview
npm run format
```

For production hosting, configure an SPA fallback to `index.html` for deep links. A non-local frontend needs an API URL reachable by its visitors and an allowed CORS origin. Vite preview uses 5173 too; stop the dev server before previewing.

## Structure

- `src/api/`: backend request/response types, HTTP client and endpoint functions.
- `src/auth/`: session restoration, context and customer route guard.
- `src/components/`: layout and shared feedback, image and totals components.
- `src/hooks/`: abortable resource loading with stale-response protection.
- `src/pages/`: restaurant search/menu, login/register, cart, checkout, confirmation and history.
- `src/utils/`: GBP minor-unit formatting and dates.
- `tests/`: Node tests for HTTP envelopes, bearer tokens, expiry, 204s, errors and currency.

Tokens live in sessionStorage for the current browser tab; reloads restore the user through `/auth/me`. Protected 401s clear the session and send customers back through login. Logout revokes the token through the API before clearing the session. A network failure during logout is shown so revocation can be retried.

## API and demo boundaries

- Public restaurants are fetched from `GET /restaurants`, `/restaurants/search?q=...` and `/restaurants/{id}`. Restaurant details include the available menu items.
- Registration creates a customer and then asks the user to log in. Customer-only pages reject manager/admin accounts.
- Cart uses `GET /cart`, `POST /cart/items`, `PUT /cart/items/{cartLineId}`, `DELETE /cart/items/{cartLineId}` and `DELETE /cart`. Repeated adds increment quantity; quantities are 1–99. Customers can remove lines or clear their basket without changing placed orders. A basket holds items from one restaurant at a time.
- Cart display distinguishes live and locked prices. Basket and checkout use backend-returned totals. The current backend charges one £2.99 delivery fee per nonempty basket/new order, calculated in integer pence. History displays saved totals, including older zero-fee orders. With an older backend response, the basket shows only its supplied subtotal rather than inventing a fee.
- Delivery county/region maps to the backend's required `delivery_state` field. There is no separate postcode field in the API.
- `POST /orders` takes delivery details only and clears the cart on success. No payment processing is implemented. Duplicate clicks are blocked; after a failed request the user is prompted to review history/basket before retrying because the backend has no idempotency key.
- Confirmation uses the actual placement response in router history state. Without that state, the confirmation page links to history. `GET /orders` gives paginated summaries; the API has no customer order-detail endpoint.
- Restaurant and dish images come from API `image_url` values, with reviewed legacy image paths mapped to local photos. The catalogue has six fictional restaurants and 48 available dishes: six covers and 41 dishes have photos, while seven dish image paths are blank. Explore, search, and menu pages list every restaurant and dish the API returns. A missing or unapproved image shows a neutral “Photo unavailable” placeholder and does not remove the card. Image-load failures use the same placeholder. Original illustrations remain archived and are not runtime fallbacks.
- The photo manifest contains 42 unique Pexels sources across 48 photographed slots (six covers, 41 dishes and the homepage sushi hero). Photos are representative servings, not actual restaurant products. Credits and licence information are in `public/demo-images/photos/`; Pexels photos are not covered by the original artwork’s CC0 dedication. No invented ratings, delivery estimates or restaurant data are added by the UI.
- Pagination for restaurant listing/search has no total count; Next is enabled when a full page was returned, so the final next page can be empty. Order history uses the API's total.
- Google Fonts are optional; local system sans-serif fallbacks apply if unavailable.

## Manual integration checklist

Use an existing running backend and database with at least one active restaurant and available menu items. This frontend does not seed or start the database. Actual registration/order tests write through normal application API requests.

- [ ] Browse at desktop and mobile widths; check navigation, keyboard focus, readable cards and image fallbacks.
- [ ] List restaurants; search by name/location; test no matches and pagination. Open a restaurant directly by URL and filter menu categories.
- [ ] Register a customer; check missing fields, mismatched passwords, duplicate email and backend validation messages. Log in after registration.
- [ ] Test wrong credentials, session restoration after reload, logout's 204 response and protected routes while logged out.
- [ ] With an expired/revoked token, open cart/history and confirm login is required with a useful message. Login should return to the requested page.
- [ ] Add an item, add it again, open basket and adjust quantities. At 1 and 99 the corresponding controls must be disabled. Confirm cart-line IDs are used for updates.
- [ ] Add from a second restaurant while the basket is occupied; confirm the API conflict message appears and the existing basket remains.
- [ ] With externally changed menu prices/availability, verify saved/current GBP prices, notices and checkout blocking. Remove an unavailable line and verify checkout becomes available when all remaining lines are eligible.
- [ ] Review checkout totals; complete all delivery fields; submit once and check the real order number, status, item lines and totals on confirmation.
- [ ] Revisit basket after ordering; verify it is empty. Open history and check the new order, timestamps, pagination and Refresh.
- [ ] Interrupt API connectivity to check retry feedback, especially after order submission. Check history before retrying a potentially completed order.
- [ ] Refresh/deep-link each page; an empty confirmation state should lead to order history. Test unknown URLs.

## Verification scope

Lint, production build and automated client tests can run without the backend. Full browser/API integration still requires a running backend and the manual checks above; no backend or database is launched by the frontend tooling.

Customer Explore/search results require a configured photographic cover. Visibility
uses image paths, so temporary image-load errors do not hide restaurants. Orders
from hidden restaurants remain in My Orders, with expandable purchased-item and
saved delivery snapshots from `GET /api/orders`. Restart the Go API after pulling
these response changes; no migration is required.

The current form-based LOCAL-DEMO-ONLY workflow is described below. The manager workflow and optional terminal utility are documented in
[`tools/demo-order-status/README.md`](../tools/demo-order-status/README.md).
The panel now signs in an existing manager directly; no manual JWT generation or
paste is needed. Separate approval is required before changing actual order records.
Customer credentials cannot change status. No automatic status changes occur.

## In-memory local-demo manager login (frontend only)

The LOCAL-DEMO-ONLY panel in My Orders has a normal visible email input and a masked
password input, with no reveal button. Both fields clear immediately on submission
and after the attempt. Credentials are transient request data, never React state or
browser storage. No real credentials or JWTs are embedded in source or logged.

With the existing local development flags enabled, the panel calls `/auth/login`
without modifying the customer session, then `/auth/me` using the returned staff JWT.
It requires an active manager, matching login/profile identity, and an active selected
restaurant whose API `manager_id` matches that manager. It displays **Manager
authenticated**. JWTs remain only in the component-owned memory closure; only a safe
manager/restaurant identity reaches rendered UI. Each Advance Status click rechecks
restaurant ownership and uses that staff token. Backend JWT, role, ownership, local-only
and status-transition checks are unchanged and remain authoritative.

Staff 401/403 errors clear only the staff token; they never clear customer storage or
trigger customer-expiry events. Passwords, tokens and server error payloads are not
shown in error messages. No customer `acceptLogin`/logout helper is called for staff
operations. Manager logout clears local memory only and does not revoke the JWT in
PostgreSQL; issued tokens retain normal expiry/revocation rules.

Manager logout, cancellation, customer-session changes and unmount invalidate pending
requests. Abort signals plus generation guards discard late responses, even if a
request ignores abort. A full page reload discards the staff session and requires another manager login. Cancellation
cannot undo a status update already committed by the backend; refresh history if an
in-flight write was interrupted. No live update was used for automated verification.

The manager form itself uses existing authentication endpoints and needs no schema
change. The complete feature also includes backend basket, history, pricing and gated
demo-route changes; a fresh environment must run the matching backend code. The owner
has confirmed the existing local demo works. Automated login tests use synthetic
credentials and mocked responses; they do not establish fresh-database setup or live
order-write verification. See the seed documentation for fresh-environment preparation.

Verification:

```sh
npm run lint
npm run build
npm test
node scripts/check-demo-manager-login.mjs
```

The optional Chrome test requires the existing local Vite server. It creates its own
temporary browser profile, mocks every API call and blocks backend port 8080 at the
network layer. It verifies field types/clearing, success feedback, token reuse,
status/history refresh using mocked orders, logout, 401/role/ownership rejection,
late cancellation and refresh/customer-session cleanup. It makes zero real backend
requests and never touches the customer's browser profile or database.
