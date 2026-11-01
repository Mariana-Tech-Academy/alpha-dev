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
- Cart uses only `GET /cart`, `POST /cart/items` and `PUT /cart/items/{cartLineId}`. Repeated adds increment quantity. Quantities are 1–99. There is no removal or clear-cart endpoint, so there is no remove button or restaurant-switch action. An unavailable line can block the cart until its availability is restored externally.
- Cart display distinguishes live and locked prices. Checkout uses backend-returned `subtotal`, `delivery_fee` and `total` from `GET /checkout/summary`. Current backend delivery fee is zero.
- Delivery county/region maps to the backend's required `delivery_state` field. There is no separate postcode field in the API.
- `POST /orders` takes delivery details only and clears the cart on success. No payment processing is implemented. Duplicate clicks are blocked; after a failed request the user is prompted to review history/basket before retrying because the backend has no idempotency key.
- Confirmation uses the actual placement response in router history state. Without that state, the confirmation page links to history. `GET /orders` gives paginated summaries; the API has no customer order-detail endpoint.
- Restaurant images come from API `image_url` values. Missing or broken images show a decorative placeholder. The hero illustration is CSS artwork, not a restaurant listing. There are no invented ratings, delivery estimates or restaurants.
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
- [ ] With externally changed menu prices/availability, verify saved/current GBP prices, notices and checkout blocking. No removal button should appear.
- [ ] Review checkout totals; complete all delivery fields; submit once and check the real order number, status, item lines and totals on confirmation.
- [ ] Revisit basket after ordering; verify it is empty. Open history and check the new order, timestamps, pagination and Refresh.
- [ ] Interrupt API connectivity to check retry feedback, especially after order submission. Check history before retrying a potentially completed order.
- [ ] Refresh/deep-link each page; an empty confirmation state should lead to order history. Test unknown URLs.

## Verification scope

Lint, production build and automated client tests can run without the backend. Full browser/API integration still requires a running backend and the manual checks above; no backend or database is launched by the frontend tooling.
