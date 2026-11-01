# Local FoodHub presentation catalogue and seed utility

## Existing local setup and fresh checkouts

The owner has confirmed that the existing local FoodHub demo is working, including
manager login. Earlier readiness reports were written before local setup; their
“not executed” statements described that preparation stage, not the current demo.
This documentation update does not rerun seeds, inspect or modify database records,
start services, reset credentials or claim independently verified seed execution counts.

Git contains the catalogue, application code, tests and image assets. It contains no
populated database, account passwords, JWTs, database backups or private setup receipts.
A tutor's fresh checkout therefore needs its own local database/configuration and
account setup. Do not rerun setup against an existing working database just to review
this code. [Original catalogue notes](READINESS.md) and
[Tokyo additive setup notes](TOKYO_READINESS.md) distinguish historical preparation
from instructions for a new environment.

## Current catalogue and customer presentation

`catalog.json` defines six fictional restaurants with eight available dishes each:

| Restaurant           | Menu focus                           | Dishes |
| -------------------- | ------------------------------------ | -----: |
| Burger District      | Burgers, fries, wings                |      8 |
| Bella Italia Kitchen | Pizza, pasta, garlic bread           |      8 |
| Spice Route          | Biryani, curries, naan, pilau        |      8 |
| The Shawarma House   | Shawarma, falafel, grills            |      8 |
| Sweet Treats Café    | Waffles, cakes, milkshakes           |      8 |
| Tokyo Bento Kitchen  | Rice bowls, sushi, noodles and sides |      8 |

Prices are positive integer GBP pence. Addresses, sample telephone numbers and
`.example.test` emails are fictional. The backend catalogue has 48 available dishes;
41 have approved photo paths and seven have blank image paths. The frontend hides
those seven dishes from customer menus. All six covers have photographs. Explore
and search require qualifying configured covers; historical orders remain visible
regardless of a restaurant's current cover.

The asset manifest has 42 unique Pexels sources across 48 photo slots: six covers,
41 menu dishes and the homepage hero. The original 45 CC0 PNGs and editable SVGs
remain archived; they are not customer-card error fallbacks. Broken images use a
neutral placeholder. See `frontend/public/demo-images/photos/` for source credits,
licences, hashes and the offline gallery. Representative stock photos do not imply
endorsement or exact portions, ingredients, allergens or dietary suitability.

## Seed boundaries and guards

This separate development tool does not run from the application or frontend. Its
default invocation previews the catalogue offline without opening PostgreSQL.
Application tables must already exist; the tool does not migrate them.

An approved application creates missing seed-owned managers, active restaurants,
menu items and categories. Each restaurant has a dedicated manager. Manager passwords
are independently random and bcrypt-hashed; plaintext values are not saved or printed.
The seed publishes no usable manager credentials. Use separately authorised account
setup to obtain a known credential; enter it privately in the browser manager form.
No credential reset tool is required or included in this review selection.

A transactional `foodhub_demo_seed.records` registry, marked
`foodhub-local-presentation-seed:v1`, identifies rows by stable kind/seed keys.
Matching seed-owned rows are reused without updates. Drift, soft-deleted matches,
ambiguous categories, conflicting names/emails or an unrelated registry cause an
abort rather than adoption or automatic repair. Do not delete the registry or edit
keys to bypass these checks. The current six-restaurant catalogue represents 60
manager/restaurant/menu registry entries; categories are resolved separately.

Application-record writes are INSERTs within one transaction. The tool does not
update/delete accounts, change existing roles or ownership, clear carts, place orders
or change order status. It uses an advisory lock and short table locks; rejects
unexpected foreign keys, custom triggers/rules, RLS/partitions and unsafe targets;
and does not repair sequences or grant privileges. Sequence values may advance on
rolled-back inserts. An uncertain commit result needs inspection before any retry.

The configured target must be the designated loopback development database
`foodhub_new_structure`, with nonconflicting local/development environment labels.
The tool verifies the actual connected database name and server address. It reads
root `.env` settings literally through godotenv, suppresses connection details in
connection errors and redacts the configured password in seed errors. Keep real
configuration outside Git and never source it into shared logs.

## Fresh-environment procedure — explicit execution only

These are instructions for a separately authorised new local environment, not a
record of actions performed by a checkout, test run or this documentation update.

1. Prepare the application schema/configuration separately. Schema migrations are
   deferred to their own review; this seed does not repair a legacy schema.
2. Review the catalogue, image files and `preflight.sql`. Check target identity,
   schema/sequence permissions, collisions, registry state and existing data drift.
   For an existing database, prepare and verify a private backup outside the repo.
3. Preview offline if desired:

   ```sh
   go run ./tools/seed-foodhub
   ```

4. Only after explicit approval to write to that local database:

   ```sh
   go run ./tools/seed-foodhub --apply --confirm-local foodhub_new_structure
   ```

5. Inspect the inserted/reused summary and public API results. Do not assume fixed
   row IDs or insert counts. A complete matching six-restaurant setup should reuse
   all 60 registry entries on a separately approved rerun, inserting no new records.
   This expectation is not a claim of a database rerun performed during review.
6. Configure known customer/manager credentials privately through authorised account
   setup. Normal customer registration is available in the application. Seeding does
   not publish passwords or grant customers staff permissions.

Search or fetch restaurant details to inspect all eight available backend dishes per
restaurant. Customer menus deliberately show only the 41 photographed dishes. Cart
and checkout use real API data, preserve the one-restaurant basket rule, support
removal/clearing and apply one £2.99 delivery fee per nonempty new order. Historical
orders keep their saved totals. Registration, basket edits, checkout and status
advances write data; they are separate from read-only review and offline tests.

## Offline validation

```sh
go test ./... -skip '^TestOrderSnapshot$'
go vet ./...
cd frontend
npm run lint
npm run build
npm test
```

Catalogue and cart/checkout contract tests use in-memory/scripted repositories.
`TestOrderSnapshot` is excluded because it opens a real test database and migrates
its tables. Offline success does not prove fresh-database integration, live seed
application/rerun behaviour or order-write verification. Image preparation scripts
are optional tools; tests and builds do not download or seed data automatically.
