# Tokyo Bento Kitchen — additive catalogue and fresh-environment notes

The owner has confirmed the existing local FoodHub demo is working. This document
retains Tokyo's additive design and historical preparation findings. Earlier “not
executed” wording described that preparation stage; it is not the current setup
status. This review has not independently audited seed insert/rerun counts or fresh
installation behaviour and performs no database writes. See [README.md](README.md)
for the current catalogue and fresh-environment procedure.

## Menu and assets

Tokyo Bento Kitchen has eight available dishes with descriptions, integer-pence prices,
category names and local photo URLs in `catalog.json`:

| Dish                       | Pence | Category            |
| -------------------------- | ----: | ------------------- |
| Chicken Katsu Curry        |  1195 | Japanese Rice Bowls |
| Teriyaki Chicken Rice Bowl |  1095 | Japanese Rice Bowls |
| Salmon Sushi Rolls         |   895 | Sushi Rolls         |
| Vegetable Sushi Rolls      |   745 | Sushi Rolls         |
| Chicken Ramen              |  1245 | Japanese Noodles    |
| Pan-fried Gyoza            |   645 | Japanese Sides      |
| Chicken & Prawn Yakisoba   |  1295 | Japanese Noodles    |
| Assorted Mochi Trio        |   595 | Japanese Desserts   |

All eight downloaded photographs were visually reviewed against the source descriptions.
The salmon rolls include cucumber and cream cheese; yakisoba includes chicken and prawns.
The gyoza photograph does not verify its filling, so no protein/dietary claim is made.
The cover and hero share a manually reviewed sushi-platter photograph, cropped locally.
All nine new sources use the Pexels License, verified at https://www.pexels.com/license/.
This allows reuse and editing in an app/presentation, but does not imply endorsement or
permit stock-photo redistribution. Source links, photographer credits, processing,
dimensions and hashes are in `frontend/public/demo-images/photos/manifest.json` and
`ATTRIBUTION.md`. Images are representative servings; no allergen guarantees are inferred.

There are ten new optimised JPEG assets (eight dishes, cover, hero), from nine sources.
The full catalogue now has six restaurants, 48 available dishes, 41 photographed dishes,
six photo covers and the same seven blank menu image paths. Customer menus hide those
seven blank-image dishes. Original illustrations remain archived and are not rendered.
The static gallery at `/demo-images/photos/gallery.html` previews Tokyo assets without any database access.

## Additive registry and preservation

The existing `foodhub_demo_seed.records` v1 registry identifies records by `(kind, seed_key)`;
it does not freeze the first run's catalogue size. `recordedID` returns no ID for a new key,
and the existing guarded insert path creates/records that entity. Existing keys still
undergo exact field, availability, ownership, soft-delete and collision checks.
The original five entries are unchanged; a regression test hashes their complete original
catalogue payload. No migration, registry reset, update, delete or adoption of unrelated
records is introduced. The only seed-engine change is validating six expected restaurants
instead of five. All transaction, advisory lock, target, FK, trigger, collision and drift
guards remain in place.

For an environment containing only the original five seeded restaurants, if their
records still match and no new collisions exist, an approved
apply should create one manager, one active restaurant, eight menu items and five categories
(or reuse exact unambiguous existing category names). It should reuse all 50 existing
manager/restaurant/menu registry entries. A second approved apply should insert nothing
and reuse all 60 entries. These are expected outcomes for that starting state, not verified counts from the current review.
Existing accounts, restaurant/menu IDs, carts, orders and their items are not updated.

## Historical pre-addition read-only check

Checked `foodhub_new_structure` on loopback `::1`: expected v1 registry marker,
50 existing registry records, no Tokyo restaurant/manager collisions, correct FKs,
required privileges, no custom triggers/rules, no RLS/partition blockers, and sequences
not behind existing IDs. All five existing managers, five restaurants and 40 menu rows
match their catalogue/registry fields and relationships. The five new category names
have no existing matches. This was read-only; no seed or status writes were executed.
Those observations predate setup and are not a current inventory. Recheck the
actual target before any separately approved write.

## Instructions for a fresh environment

The hero is a static frontend change: refresh Vite (or serve the rebuilt `dist`).
A fresh database does not gain Tokyo from Git assets alone; it needs separately
approved catalogue setup. Do not repeat this step against the working local demo
merely to review the code.
Before any write: rerun the read-only preflight, check registry/data drift and take/verify
a fresh custom-format backup. Stop on collisions or drift; do not bypass checks.
After explicit approval only, from the project root:

```sh
go run ./tools/seed-foodhub --apply --confirm-local foodhub_new_structure
```

Then verify active restaurant, eight available menus, all nine restaurant/menu image URLs,
unchanged existing records, and zero inserts on a separately approved rerun. Refresh Explore.
No backend restart or schema migration is required for this catalogue/hero release.

Offline validation (these commands do not connect to PostgreSQL):

```sh
go test ./tools/seed-foodhub
go run ./tools/seed-foodhub
cd frontend
npm run lint
npm run build
npm test
```

To reproduce photo downloads/optimisation, run `node frontend/scripts/download-tokyo-photos.mjs`
from the root, then `node frontend/scripts/finalise-tokyo-photos.mjs` and
`node frontend/scripts/build-photo-gallery.mjs`. Promotion checks the manually reviewed
source flags; inspect any changed download before approving it. Download receipts preserve
original hashes; temporary original files are outside the repository. The existing five
restaurants' image files are never reprocessed by these additive scripts.
