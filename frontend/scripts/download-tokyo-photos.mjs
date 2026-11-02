// Additive photo candidates only. No DB access; approval requires visual review.
import { readFile, writeFile, mkdir } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { createHash } from "node:crypto";
import sharp from "sharp";
const run = promisify(execFile);
const root = fileURLToPath(new URL("../../", import.meta.url));
const sources = JSON.parse(
  await readFile(
    resolve(root, "tools/seed-foodhub/photo-sources.json"),
    "utf8",
  ),
).filter((s) => ["tokyo-bento-kitchen", "homepage"].includes(s.restaurant));
const receipts = [];
for (let start = 0; start < sources.length; start += 3) {
  await Promise.all(
    sources.slice(start, start + 3).map(async (s) => {
      if (
        new URL(s.source_url).hostname !== "www.pexels.com" ||
        new URL(s.download_url).hostname !== "images.pexels.com" ||
        s.licence_url !== "https://www.pexels.com/license/" ||
        !s.source_verified_on
      )
        throw Error("Unverified source");
      const raw = resolve("/private/tmp", `foodhub-tokyo-${s.id}.jpg`);
      await run("curl", [
        "--fail",
        "--silent",
        "--show-error",
        "--location",
        "--proto",
        "=https",
        "--max-time",
        "40",
        s.download_url,
        "--output",
        raw,
      ]);
      const input = await readFile(raw);
      const path =
        s.restaurant === "homepage"
          ? "/demo-images/photos/hero/sushi-platter.jpg"
          : `/demo-images/photos/menu/tokyo-bento-kitchen/${s.item}.jpg`;
      const output = await sharp(input)
        .rotate()
        .resize({
          width: 1100,
          height: 1100,
          fit: "inside",
          withoutEnlargement: true,
        })
        .jpeg({ quality: 82, mozjpeg: true })
        .toBuffer();
      await mkdir(resolve(root, "frontend/public", "." + path, ".."), {
        recursive: true,
      });
      await writeFile(resolve(root, "frontend/public", "." + path), output);
      const info = await sharp(output).metadata();
      receipts.push({
        ...s,
        path,
        bytes: output.length,
        width: info.width,
        height: info.height,
        sha256: createHash("sha256").update(output).digest("hex"),
        original_sha256: createHash("sha256").update(input).digest("hex"),
        processing:
          "Auto-orient, fit within 1100 × 1100, JPEG quality 82, metadata stripped; no generated content.",
      });
    }),
  );
}
await writeFile(
  resolve(root, "frontend/public/demo-images/photos/tokyo-candidates.json"),
  JSON.stringify(
    receipts.sort((a, b) => a.item.localeCompare(b.item)),
    null,
    2,
  ) + "\n",
);
console.log(
  `Downloaded and optimised ${receipts.length} candidates; visual review still required.`,
);
