// Offline promotion of manually reviewed photographs. Never opens a database.
import { readFile, writeFile, mkdir, unlink } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { resolve, dirname } from "node:path";
import { createHash } from "node:crypto";
const root = fileURLToPath(new URL("../../", import.meta.url));
const dir = resolve(root, "frontend/public/demo-images/photos");
const catalogueFile = resolve(root, "tools/seed-foodhub/catalog.json");
const catalogue = JSON.parse(await readFile(catalogueFile, "utf8"));
const sources = JSON.parse(
  await readFile(
    resolve(root, "tools/seed-foodhub/photo-sources.json"),
    "utf8",
  ),
);
const downloads = JSON.parse(
  await readFile(resolve(dir, "candidates.json"), "utf8"),
);
const fallbackManifest = JSON.parse(
  await readFile(
    resolve(root, "frontend/public/demo-images/manifest.json"),
    "utf8",
  ),
);
const coverDish = {
  "burger-district": "classic-smash",
  "bella-italia-kitchen": "margherita",
  "spice-route": "butter-chicken",
  "shawarma-house": "chicken-wrap",
  "sweet-treats-cafe": "strawberry-waffle",
};
const missing = {
  "burger-district/plant-burger":
    "No verified professional photograph of the specified bean-and-vegetable patty; generic vegan and branded chain burgers were not substituted.",
  "bella-italia-kitchen/alfredo":
    "Closest free sources identify fettuccine or tomato-based tagliatelle; the specified creamy garlic chicken-and-mushroom tagliatelle was not verified.",
  "spice-route/lamb-rogan-josh":
    "No directly verified free source for lamb Rogan Josh; generic meat curries and recipe-site images were rejected.",
  "shawarma-house/lamb-wrap":
    "Free sources identified chicken or unspecified shawarma; lamb wrap could not be verified.",
};
const records = [],
  map = {};
async function promote(restaurant, item, cover = false) {
  const key = restaurant.key + (cover ? "/cover" : "/" + item.key);
  const fallback = cover
    ? `/demo-images/covers/${restaurant.key}.png`
    : `/demo-images/menu/${restaurant.key}/${item.key}.png`;
  if (!fallbackManifest.some((entry) => entry.path === fallback))
    throw new Error("Missing fallback for " + key);
  const targetItem = cover ? coverDish[restaurant.key] : item.key;
  const source = sources.find(
    (entry) => entry.restaurant === restaurant.key && entry.item === targetItem,
  );
  const receipt = source && downloads.find((entry) => entry.id === source.id);
  let record = {
    key,
    label: cover ? restaurant.name + " cover" : item.name,
    fallback_path: fallback,
    status: "neutral-placeholder",
    path: "",
  };
  if (
    source?.visual_approved &&
    source.source_verified_on &&
    receipt?.download_status === "verified-image"
  ) {
    const sourceFile = resolve(dir, `pexels-${source.id}.jpg`);
    const path = cover
      ? `/demo-images/photos/covers/${restaurant.key}.jpg`
      : `/demo-images/photos/menu/${restaurant.key}/${item.key}.jpg`;
    const target = resolve(root, "frontend/public", "." + path);
    // Allow rerunning promotion from existing reviewed output after candidate cleanup.
    let bytes;
    try {
      bytes = await readFile(sourceFile);
    } catch {
      bytes = await readFile(target);
    }
    if (createHash("sha256").update(bytes).digest("hex") !== receipt.sha256)
      throw new Error("Candidate checksum mismatch: " + key);
    await mkdir(dirname(target), { recursive: true });
    await writeFile(target, bytes);
    record = {
      ...record,
      status: "photograph-verified",
      path,
      source_url: source.source_url,
      download_url: source.download_url,
      photographer: source.photographer,
      licence: source.licence,
      licence_url: source.licence_url,
      source_verified_on: source.source_verified_on,
      visual_reviewed_on: source.visual_reviewed_on,
      match_notes: source.match_notes,
      bytes: bytes.length,
      width: receipt.width,
      height: receipt.height,
      sha256: receipt.sha256,
      original_sha256: receipt.original_sha256,
      processing: receipt.processing,
    };
    map[fallback] = path;
  } else {
    record.reason =
      source?.visual_review_notes ||
      missing[key] ||
      "No source/licence/download fully verified.";
  }
  if (cover) restaurant.image_url = record.path;
  else item.image_url = record.path;
  records.push(record);
}
for (const restaurant of catalogue.filter(
  (r) => r.key !== "tokyo-bento-kitchen",
)) {
  await promote(restaurant, null, true);
  for (const item of restaurant.items) await promote(restaurant, item);
}
// Keep the additive, separately reviewed Tokyo/hero records intact.
const additive = JSON.parse(
  await readFile(resolve(dir, "manifest.json"), "utf8"),
).filter(
  (row) =>
    row.key.startsWith("tokyo-bento-kitchen/") || row.key === "homepage/hero",
);
records.push(...additive);
await writeFile(catalogueFile, JSON.stringify(catalogue, null, 2) + "\n");
await writeFile(
  resolve(dir, "manifest.json"),
  JSON.stringify(records, null, 2) + "\n",
);
await writeFile(
  resolve(root, "frontend/src/utils/demo-photo-map.json"),
  JSON.stringify(map, null, 2) + "\n",
);
// Do not ship rejected or redundant candidate files in the public directory.
for (const row of downloads)
  if (row.path?.match(/^\/demo-images\/photos\/pexels-\d+\.jpg$/))
    await unlink(resolve(root, "frontend/public", "." + row.path)).catch(
      (error) => {
        if (error.code !== "ENOENT") throw error;
      },
    );
const approved = records.filter(
  (record) => record.status === "photograph-verified",
);
console.log(
  `${new Set(approved.map((record) => record.source_url)).size} unique verified photographs; ${approved.length}/${records.length} image slots replaced; ${records.length - approved.length} neutral placeholders.`,
);
