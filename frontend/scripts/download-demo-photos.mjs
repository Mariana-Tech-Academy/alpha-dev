// Downloads only individually reviewed Pexels source candidates. No DB access.
// Outputs remain review candidates; this script never modifies the seed catalogue.
import { readFile, mkdir, writeFile } from "node:fs/promises";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { fileURLToPath } from "node:url";
import { resolve } from "node:path";
import { createHash } from "node:crypto";
import sharp from "sharp";
const run = promisify(execFile);
const root = fileURLToPath(new URL("../../", import.meta.url));
const candidates = JSON.parse(
  await readFile(
    resolve(root, "tools/seed-foodhub/photo-sources.json"),
    "utf8",
  ),
);
const dir = resolve(root, "frontend/public/demo-images/photos");
await mkdir(dir, { recursive: true });
const results = [];
for (let start = 0; start < candidates.length; start += 3) {
  await Promise.all(
    candidates.slice(start, start + 3).map(async (candidate) => {
      if (
        new URL(candidate.source_url).hostname !== "www.pexels.com" ||
        new URL(candidate.download_url).hostname !== "images.pexels.com" ||
        candidate.licence_url !== "https://www.pexels.com/license/" ||
        !candidate.source_verified_on
      )
        throw new Error("Unverified source");
      const raw = resolve("/private/tmp", `foodhub-pexels-${candidate.id}.jpg`);
      const path = `/demo-images/photos/pexels-${candidate.id}.jpg`;
      try {
        await run("curl", [
          "--silent",
          "--show-error",
          "--fail",
          "--location",
          "--connect-timeout",
          "10",
          "--max-time",
          "40",
          candidate.download_url,
          "--output",
          raw,
        ]);
        const original = await readFile(raw);
        await sharp(original)
          .rotate()
          .resize({
            width: 1100,
            height: 1100,
            fit: "inside",
            withoutEnlargement: true,
          })
          .jpeg({ quality: 84, mozjpeg: true })
          .toFile(resolve(dir, `pexels-${candidate.id}.jpg`));
        const processed = await readFile(
          resolve(dir, `pexels-${candidate.id}.jpg`),
        );
        const info = await sharp(processed).metadata();
        results.push({
          ...candidate,
          path,
          original_sha256: createHash("sha256").update(original).digest("hex"),
          sha256: createHash("sha256").update(processed).digest("hex"),
          bytes: processed.length,
          width: info.width,
          height: info.height,
          processing:
            "Auto-orientation; proportional resize to max 1100px; JPEG quality 84; metadata removed; no generated or substituted content.",
          download_status: "verified-image",
        });
        process.stdout.write(`Downloaded photo ${candidate.id}\n`);
      } catch (error) {
        results.push({
          ...candidate,
          download_status: "failed",
          error: String(error.message),
        });
        process.stdout.write(
          `Download failed for ${candidate.id}; retain illustration\n`,
        );
      }
    }),
  );
}
results.sort(
  (a, b) =>
    candidates.findIndex((x) => x.id === a.id) -
    candidates.findIndex((x) => x.id === b.id),
);
await writeFile(
  resolve(dir, "candidates.json"),
  JSON.stringify(results, null, 2) + "\n",
);
process.stdout.write(
  `Candidate downloads: ${results.filter((x) => x.download_status === "verified-image").length}/${candidates.length}. Await visual review before promotion.\n`,
);
