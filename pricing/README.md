# New-order delivery pricing — code prepared, activation not performed

The authoritative policy is `Calculate` in `pricing.go`: a nonempty basket/order has
one 299-pence delivery fee and total = locked cart subtotal + fee. An empty basket has
zero subtotal, fee and total. Zero-priced food still attracts the flat delivery fee.

Cart preview, validated checkout summary and transactional order placement all use
this policy. The subtotal continues to use server-side locked cart-line prices and
quantities; availability and single-restaurant checks are unchanged. Nonempty basket
previews retain existing unavailable-item notices and checkout blocking.

Only newly placed orders save the current fee. History/confirmation read saved values
and never recalculate fees. Historical zero-fee orders are not updated. No database
model/schema/default changed; existing orders.delivery_fee/subtotal/total fields are
reused. Client pricing fields remain rejected by the order handler's strict decoder.

The React Basket uses the existing Totals component and backend-returned values.
Cart pricing fields are optional during rollout so Vite hot reload against the old
backend keeps the previous subtotal display; the client does not invent a 299-pence
fallback. Checkout, confirmation and history keep their existing saved/API breakdowns.

## Activation only after separate approval

No services were restarted, database records written, migration run or live order
placed to validate this change. Go's existing running binary retains its old zero-fee
policy until restart; do not present the new fee as active yet.

1. Pause checkout activity during the coordinated switch.
2. Reidentify the existing FoodHub backend process and stop only that verified process.
3. From the repository root, start:

   ```sh
   APP_ENV=development FOODHUB_DEMO_SAFE_START=1 FOODHUB_DEMO_STATUS=1 go run .
   ```

4. Require the LOCAL DEMO SAFE START banner confirming database identity verification,
   AutoMigrate skipped, cleanup skipped and JWT/revocation enforcement retained. Stop
   if any safety or port check fails. Never omit safe-start to work around a failure.
5. The existing Vite server can serve the updated code; hard-refresh Basket and Checkout
   to replace cached zero-fee responses. If a frontend restart is separately approved,
   run `VITE_FOODHUB_DEMO_STATUS=1 npm run dev` from `frontend` after stopping only its
   verified listener. Vite must not silently move to another port.
6. With the customer's existing session, inspect GET /api/cart and GET
   /api/checkout/summary read-only: a nonempty valid basket must report delivery_fee=299
   and total=subtotal+299; empty cart values must all be zero. Do not place an order
   until separately authorised. Check existing history still shows saved zero fees.

## Verification

Go tests (excluding the known PostgreSQL-writing TestOrderSnapshot) and vet, plus
frontend lint/build/tests, cover multi-item quantities, one fee, empty baskets, free
food, rejected client pricing, actual GORM INSERT arguments and transaction semantics,
cart clearance/Pending state, historical compatibility and GBP display. Tests use
scripted SQL connections; no live PostgreSQL writes are required. The excluded test's
expected values were updated for future approved execution, but it was not run.
