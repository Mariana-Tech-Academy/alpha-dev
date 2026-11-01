# FoodHub stock photographs

The current manifest contains 42 unique Pexels source photographs across 48 photo
slots: six restaurant covers, 41 menu dishes and the homepage sushi hero. Seven
additional menu slots have no approved photograph. Their backend catalogue image
paths are blank and the customer menus hide them. Original CC0 artwork remains
archived; customer cards do not use it as a fallback. A configured image that fails
to load shows a neutral “Photo unavailable” placeholder without hiding its card.

The owner has confirmed the existing local demo works. Earlier asset-preparation
notes saying that no seed had run described an earlier preparation stage, not the
current installation. These files do not populate a new database; follow the
[seed guide](../../../../tools/seed-foodhub/README.md) for a fresh environment.

## Sources, licence and scope

Photo source decisions are recorded in `tools/seed-foodhub/photo-sources.json` and
`manifest.json`. Credits and source/licence links are in `ATTRIBUTION.md`.
The Pexels License applies to photographs; the original illustrations' CC0 dedication
does not. Photos are representative food servings for fictional restaurants, not
photographs of actual FoodHub establishments. They imply no endorsement, exact
portions, recipes, allergens or dietary suitability. Do not redistribute them as a
stock-photo service or sell unaltered copies. Licence: https://www.pexels.com/license/ .

- `gallery.html` previews photo slots and neutral placeholders offline, without API calls.
- `manifest.json` records mappings, approval status, photographer credits, processing,
  dimensions, byte sizes and SHA-256 hashes.
- `ATTRIBUTION.md` provides photographer credits, source links and placeholder reasons.
- `candidates.json` and `tokyo-candidates.json` are public photo provenance metadata,
  not database backup receipts or account data. Rejected candidates are not approvals.

## Runtime presentation

Approved legacy illustration URLs can map to reviewed photos through
`frontend/src/utils/demo-photo-map.json`, without a database update. Unknown or blank
API values do not receive arbitrary stock photos. Restaurant visibility requires a
qualifying configured cover; menu visibility requires a configured photo. Historical
orders remain visible regardless of current image metadata.

All runtime demo photos are local JPEG assets copied by Vite to the build. Menu frames
retain the whole serving; covers and the hero use cropped frames. Vite's base path is
honoured. There are no photo-provider credentials or runtime API downloads. `sharp`
is development tooling for decoding, proportional resizing, optimisation and metadata
removal. Original image sources and licence manifests remain reviewable.

## Optional asset preparation

These scripts are separate from tests/builds. Download scripts access reviewed Pexels
sources; promotion and gallery generation operate on local files. They do not seed
PostgreSQL. Review a changed photo and its licence before marking it approved.

From the repository root:

```sh
node frontend/scripts/download-demo-photos.mjs
node frontend/scripts/promote-demo-photos.mjs
node frontend/scripts/download-tokyo-photos.mjs
node frontend/scripts/finalise-tokyo-photos.mjs
node frontend/scripts/build-photo-gallery.mjs
```

Do not run these merely to review the checked-in assets. Frontend tests verify
manifest mappings, file hashes, archived artwork and photo visibility. The optional
`node frontend/scripts/check-photo-gallery.mjs` checks the gallery with an isolated
Chrome profile on macOS. Asset tests are not fresh-database integration tests.
