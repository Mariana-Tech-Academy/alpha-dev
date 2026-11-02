import { test } from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { createHash } from "node:crypto";
import { fileURLToPath } from "node:url";
import { resolve } from "node:path";
import sharp from "sharp";
import { imageUrl, imageCandidates } from "../src/utils/images.ts";
const root = fileURLToPath(new URL("../../", import.meta.url));
const catalog = JSON.parse(
  await readFile(resolve(root, "tools/seed-foodhub/catalog.json"), "utf8"),
);
const manifest = JSON.parse(
  await readFile(
    resolve(root, "frontend/public/demo-images/manifest.json"),
    "utf8",
  ),
);
test("all 45 original CC0 fallback assets remain intact", async () => {
  const paths = catalog.flatMap((r) => [
    r.image_url,
    ...r.items.map((d) => d.image_url),
  ]);
  assert.equal(paths.length, 54);
  assert.equal(paths.filter(Boolean).length, 47);
  assert.equal(paths.filter((path) => !path).length, 7);
  assert.equal(new Set(paths.filter(Boolean)).size, 47);
  assert.equal(manifest.length, 45);
  assert.ok(manifest.reduce((sum, item) => sum + item.bytes, 0) < 1024 * 1024);
  for (const entry of manifest) {
    const path = entry.path;
    assert.equal(entry.license, "CC0-1.0");
    const png = await readFile(resolve(root, "frontend/public", "." + path));
    assert.equal(createHash("sha256").update(png).digest("hex"), entry.sha256);
    assert.equal(png.length, entry.bytes);
    const info = await sharp(png).metadata();
    assert.equal(info.format, "png");
    assert.equal(info.width, path.includes("/covers/") ? 960 : 480);
    assert.equal(info.height, path.includes("/covers/") ? 720 : 360);
    const svg = await readFile(resolve(root, entry.source), "utf8");
    assert.ok(svg.includes("CC0-1.0"));
    assert.ok(
      !/<script|<image|https?:\/\//i.test(
        svg.replace("http://www.w3.org/2000/svg", ""),
      ),
    );
  }
});
test("local image paths respect Vite base without changing remote URLs", () => {
  assert.equal(
    imageUrl("/demo-images/covers/burger-district.png", "/foodhub/"),
    "/foodhub/demo-images/covers/burger-district.png",
  );
  assert.equal(
    imageUrl("https://images.example.test/dish.jpg"),
    "https://images.example.test/dish.jpg",
  );
  assert.equal(imageUrl(""), undefined);
  assert.equal(imageUrl(undefined), undefined);
  assert.equal(imageUrl("javascript:alert(1)"), undefined);
  assert.equal(imageUrl("//unexpected.example.test/image.png"), undefined);
});

const photoManifest = JSON.parse(
  await readFile(
    resolve(root, "frontend/public/demo-images/photos/manifest.json"),
    "utf8",
  ),
);
test("42 verified photo sources cover 48 slots and seven honest fallbacks remain", async () => {
  const photos = photoManifest.filter(
    (entry) => entry.status === "photograph-verified",
  );
  assert.equal(photoManifest.length, 55);
  assert.equal(photos.length, 48);
  assert.equal(new Set(photos.map((entry) => entry.source_url)).size, 42);
  assert.equal(
    photoManifest.filter((entry) => entry.status === "neutral-placeholder")
      .length,
    7,
  );
  for (const restaurant of catalog) {
    const refs = [
      restaurant.image_url,
      ...restaurant.items.map((item) => item.image_url),
    ];
    for (const path of refs)
      assert.ok(photoManifest.some((entry) => entry.path === path));
  }
  for (const entry of photoManifest) {
    if (entry.fallback_path)
      assert.ok(
        manifest.some((fallback) => fallback.path === entry.fallback_path),
      );
    if (entry.status !== "photograph-verified") {
      assert.ok(entry.reason);
      assert.equal(entry.path, "");
      continue;
    }
    assert.equal(new URL(entry.source_url).hostname, "www.pexels.com");
    assert.equal(new URL(entry.download_url).hostname, "images.pexels.com");
    assert.equal(entry.licence, "Pexels License");
    assert.equal(entry.licence_url, "https://www.pexels.com/license/");
    assert.ok(
      entry.photographer &&
        entry.source_verified_on &&
        entry.visual_reviewed_on,
    );
    const image = await readFile(
      resolve(root, "frontend/public", "." + entry.path),
    );
    assert.equal(
      createHash("sha256").update(image).digest("hex"),
      entry.sha256,
    );
    const info = await sharp(image).metadata();
    assert.equal(info.format, "jpeg");
    assert.equal(info.width, entry.width);
    assert.equal(info.height, entry.height);
    assert.ok(info.width <= 1100 && info.height <= 1100);
  }
});
test("cards display only photographs; legacy art and missing photos use neutral placeholders", () => {
  const photo = photoManifest.find(
    (entry) => entry.status === "photograph-verified",
  );
  assert.deepEqual(imageCandidates(photo.path), [photo.path]);
  assert.deepEqual(imageCandidates(photo.fallback_path), [photo.path]);
  assert.deepEqual(imageCandidates(photo.path, "/foodhub/"), [
    "/foodhub" + photo.path,
  ]);
  const missing = photoManifest.find(
    (entry) => entry.status === "neutral-placeholder",
  );
  assert.deepEqual(imageCandidates(missing.path), []);
  assert.deepEqual(imageCandidates(missing.fallback_path), []);
  assert.deepEqual(imageCandidates(undefined), []);
  assert.ok(
    catalog.every(
      (r) => r.items.length === 8 && r.items.every((d) => d.available),
    ),
  );
});

test("photo detection marks 41 dishes and leaves seven blanks for a placeholder", async () => {
  const { hasMenuPhoto } = await import("../src/utils/images.ts");
  const dishes = catalog.flatMap((restaurant) => restaurant.items);
  assert.equal(
    dishes.filter((dish) => hasMenuPhoto(dish.image_url)).length,
    41,
  );
  assert.equal(
    dishes.filter((dish) => !hasMenuPhoto(dish.image_url)).length,
    7,
  );
  assert.equal(dishes.length, 48);
  assert.ok(dishes.every((dish) => dish.available));
  assert.ok(catalog.every((restaurant) => hasMenuPhoto(restaurant.image_url)));
  // Visibility is independent of fetch success, URL extension and Vite base.
  assert.equal(hasMenuPhoto("https://images.example.test/photo?id=12"), true);
  assert.equal(hasMenuPhoto("/demo-images/photos/menu/example.jpg"), true);
  assert.equal(hasMenuPhoto("   "), false);
  assert.equal(hasMenuPhoto("/demo-images/menu/unapproved.png"), false);
});

test("cover detection accepts the six photographs and rejects missing or placeholder covers", async () => {
  const { hasRestaurantPhoto } = await import("../src/utils/images.ts");
  assert.equal(
    catalog.filter((r) => hasRestaurantPhoto(r.image_url)).length,
    6,
  );
  for (const path of [
    "",
    undefined,
    " ",
    "/images/placeholder.jpg",
    "/images/illustration.svg",
    "/demo-images/covers/unapproved.png",
  ])
    assert.equal(hasRestaurantPhoto(path), false, String(path));
  assert.equal(
    hasRestaurantPhoto("https://photos.example.test/cover?id=9"),
    true,
  );
  for (const cover of catalog.map((r) => r.image_url))
    assert.equal(hasRestaurantPhoto(cover), true);
});
