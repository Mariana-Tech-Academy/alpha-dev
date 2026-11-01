// Offline, additive promotion after visual review. No database access.
import { readFile, writeFile, mkdir } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { createHash } from "node:crypto";
import sharp from "sharp";
const root = fileURLToPath(new URL("../../", import.meta.url));
const dir = resolve(root, "frontend/public/demo-images/photos");
const candidates = JSON.parse(
  await readFile(resolve(dir, "tokyo-candidates.json"), "utf8"),
);
const sourcesFile = resolve(root, "tools/seed-foodhub/photo-sources.json");
const sources = JSON.parse(await readFile(sourcesFile, "utf8"));
const records = JSON.parse(
  await readFile(resolve(dir, "manifest.json"), "utf8"),
).filter(
  (r) => !r.key.startsWith("tokyo-bento-kitchen/") && r.key !== "homepage/hero",
);
const catalog = JSON.parse(
  await readFile(resolve(root, "tools/seed-foodhub/catalog.json"), "utf8"),
);
for (const receipt of candidates) {
  const source = sources.find(
    (s) => s.restaurant === receipt.restaurant && s.item === receipt.item,
  );
  if (!source?.visual_approved)
    throw Error("Visual review required: " + receipt.item);
  const item = catalog
    .find((r) => r.key === "tokyo-bento-kitchen")
    .items.find((d) => d.key === receipt.item);
  let bytes = await readFile(
    resolve(root, "frontend/public", "." + receipt.path),
  );
  let processing = receipt.processing;
  if (receipt.restaurant === "homepage") {
    const original = await sharp(
      resolve("/private/tmp", `foodhub-tokyo-${receipt.id}.jpg`),
    )
      .rotate()
      .toBuffer();
    const meta = await sharp(original).metadata();
    const cropHeight = Math.min(
      meta.height,
      Math.round((meta.width * 760) / 1100),
    );
    const top = Math.min(
      meta.height - cropHeight,
      Math.round(meta.height * 0.38),
    );
    bytes = await sharp(original)
      .extract({ left: 0, top, width: meta.width, height: cropHeight })
      .resize(1100, 760)
      .jpeg({ quality: 82, mozjpeg: true })
      .toBuffer();
    await writeFile(
      resolve(root, "frontend/public", "." + receipt.path),
      bytes,
    );
    processing =
      "Auto-orient; Manually reviewed sushi-centred 1100 × 760 crop; JPEG quality 82; metadata stripped; no generated content.";
  }
  const info = await sharp(bytes).metadata();
  const row = {
    ...receipt,
    ...source,
    key:
      receipt.restaurant === "homepage"
        ? "homepage/hero"
        : `tokyo-bento-kitchen/${receipt.item}`,
    label: item?.name || "Homepage sushi platter",
    status: "photograph-verified",
    fallback_path: null,
    processing,
    bytes: bytes.length,
    width: info.width,
    height: info.height,
    sha256: createHash("sha256").update(bytes).digest("hex"),
  };
  records.push(row);
  if (receipt.restaurant === "homepage") {
    const path = "/demo-images/photos/covers/tokyo-bento-kitchen.jpg";
    await mkdir(resolve(dir, "covers"), { recursive: true });
    await writeFile(resolve(root, "frontend/public", "." + path), bytes);
    records.push({
      ...row,
      key: "tokyo-bento-kitchen/cover",
      label: "Tokyo Bento Kitchen cover",
      path,
    });
  }
}
await writeFile(
  resolve(dir, "manifest.json"),
  JSON.stringify(records, null, 2) + "\n",
);
console.log(
  "Promoted eight menu photographs, one cover and the homepage hero; existing assets unchanged.",
);
