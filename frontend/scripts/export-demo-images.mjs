// Export original SVG sources to compact PNGs. No network or database access.
import sharp from "sharp";
import { readFile, writeFile, stat } from "node:fs/promises";
import { createHash } from "node:crypto";
import { fileURLToPath } from "node:url";
import { resolve, relative } from "node:path";
const root = fileURLToPath(new URL("../../", import.meta.url));
const publicRoot = resolve(root, "frontend/public/demo-images");
const sourceRoot = resolve(root, "tools/seed-foodhub/image-sources");
const input = [];
for await (const chunk of process.stdin) input.push(chunk);
const manifest = JSON.parse(Buffer.concat(input).toString());
for (const item of manifest) {
  const source = resolve(root, item.source);
  const target = resolve(publicRoot, item.path.slice("/demo-images/".length));
  if (
    relative(sourceRoot, source).startsWith("..") ||
    relative(publicRoot, target).startsWith("..")
  )
    throw new Error("Asset path escapes artwork directories");
  await sharp(source)
    .png({ palette: true, colours: 128, compressionLevel: 9, effort: 7 })
    .toFile(target);
  const bytes = await readFile(target);
  item.bytes = (await stat(target)).size;
  item.sha256 = createHash("sha256").update(bytes).digest("hex");
}
await writeFile(
  resolve(publicRoot, "manifest.json"),
  JSON.stringify(manifest, null, 2) + "\n",
);
process.stdout.write(
  `Exported ${manifest.length} original PNG images; ${manifest.reduce((sum, item) => sum + item.bytes, 0).toLocaleString()} bytes total.\n`,
);
