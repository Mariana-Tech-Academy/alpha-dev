# Current demo photography

The current catalogue uses six restaurant covers and 41 photographed menu dishes,
plus a photographic homepage hero. The photo manifest has 42 unique Pexels sources
across those 48 photo slots. Seven backend dishes have blank image paths and are
hidden from customer menus. Broken configured images show neutral placeholders;
archived illustrations are not runtime error fallbacks.

Start with [the photo gallery](photos/gallery.html),
[photo credits](photos/ATTRIBUTION.md) and [photo documentation](photos/README.md).
Pexels photographs are not covered by the original artwork's CC0 dedication.
The owner has confirmed the existing local demo works; fresh environment setup is
separate and is described in the [seed guide](../../../tools/seed-foodhub/README.md).

# Original FoodHub local fallback image assets

45 original vector-derived PNG illustrations: five restaurant food-spread covers
and one image for each of the 40 demo dishes, from the original five-restaurant catalogue.
These are illustrations of fictional servings, not photographs of actual meals.
These original illustrations use no third-party photographs, logos or fonts. PNG exports and editable SVG sources are provided under the CC0-1.0
dedication in `LICENSE.txt`. No attribution is required; do not imply endorsement.

Sources: `tools/seed-foodhub/generate_images.py` and the generated SVG files under
`tools/seed-foodhub/image-sources/`. `manifest.json` records every asset's source,
licence, byte count and SHA-256. This makes the origin and file integrity reviewable.
The CC0 reference is https://creativecommons.org/publicdomain/zero/1.0/ .

## Preview without database access

Open `gallery.html` from this directory in a browser. It directly loads the local
files and contains no API calls, seed execution or simulated restaurant records.
If the frontend is already running, use `/demo-images/gallery.html`.

## Technical details

- PNG covers: 960×720; menu images: 480×360.
- Exports use a 128-colour palette and maximum PNG compression. All 45 image files
  together are under 600 KB. Sharp is development tooling only; it is not shipped
  to or run in the browser.
- `frontend/public` files are copied to `dist` by Vite. Screensharing needs no
  external image host and no internet connection for these images.
- Root-relative seed values such as `/demo-images/covers/burger-district.png`
  are served by the frontend origin, not the Go API. The frontend handles a Vite
  base-path prefix. Other consumers of the API must resolve these paths against
  the frontend origin or arrange their own asset hosting.
- The backend already stores/returns `image_url`; PNG paths also meet its current
  extension restriction. No backend model, HTTP route or migration change is
  needed. Direct file upload remains unsupported by the backend.
- The shared FoodImage component displays these images on restaurant cards,
  restaurant headers and menu cards. It uses responsive fixed frames, lazy
  loading, asynchronous decoding and alt text. Missing/broken URLs render an
  accessible bowl illustration. A failed old URL does not hide a new URL.
- Existing database rows have not been changed. Image paths have been added to
  the proposed local seed catalogue and insert logic only. The seed still refuses
  to modify existing rows and still requires separate execution approval. An
  older seed already inserted with blank images would be reported as drift, not
  automatically updated.

## Regenerate (files only; never connects to a database)

After `cd frontend && npm install`, from the repository root run:

```sh
python3 tools/seed-foodhub/generate_images.py
```

The Python script generates SVGs and invokes
`frontend/scripts/export-demo-images.mjs` to export PNGs and the manifest. The
supplied assets do not need Python or Sharp to display. The script does not run
`tools/seed-foodhub/main.go`, read `.env` or access PostgreSQL.

## Limits and manual checks

Visual review sampled the burger cover and strawberry waffle. No browser-driven
end-to-end screenshots were taken. After the seed is separately approved, verify
restaurant and menu cards at desktop/mobile widths, and temporarily block a PNG
request in browser DevTools to check the fallback. Loading restaurant records
still requires the real backend; the offline gallery is purely an artwork review.
For a future change to real photography, separately verify each photograph's
source/licence, subject suitability and redistribution rights rather than relying
on search-engine thumbnails or hotlinked URLs.
