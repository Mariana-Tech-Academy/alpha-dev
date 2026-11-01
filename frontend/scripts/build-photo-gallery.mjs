// Generate an offline review gallery/credits from the actual promoted asset manifest.
import { readFile, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { resolve } from "node:path";
const root = fileURLToPath(new URL("../../", import.meta.url));
const dir = resolve(root, "frontend/public/demo-images/photos");
const rows = JSON.parse(await readFile(resolve(dir, "manifest.json"), "utf8"));
const escape = (value) =>
  String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll('"', "&quot;");
const link = (path) =>
  path.startsWith("/demo-images/photos/")
    ? path.slice("/demo-images/photos/".length)
    : "../" + path.slice("/demo-images/".length);
const approved = rows.filter((row) => row.status === "photograph-verified");
const unique = [
  ...new Map(approved.map((row) => [row.source_url, row])).values(),
];
let credits =
  "# FoodHub photograph credits\n\nVerified on 2026-11-01. All selected photographs use the Pexels License:\nhttps://www.pexels.com/license/\n\nPermits use in an application/presentation and editing. This is not CC0.\nNo endorsement by photographers, pictured establishments or brands is implied.\nDo not sell unaltered photos or redistribute them as a stock-photo service.\nPhotos are representative servings for fictional restaurants; garnishes and\nportion sizes may differ. No recipe, allergen or dietary guarantees are inferred.\n\n";
for (const row of unique)
  credits += `- ${row.photographer}: ${row.source_url}\n  Licence: ${row.licence_url}\n  Processing: ${row.processing}\n`;
credits += "\nNeutral placeholders for these dishes:\n";
for (const row of rows.filter((row) => row.status === "neutral-placeholder"))
  credits += `- ${row.label}: ${row.reason}\n`;
await writeFile(resolve(dir, "ATTRIBUTION.md"), credits);
const style =
  "*{box-sizing:border-box}body{font:15px/1.6 system-ui;background:#faf8f2;color:#233c2e;margin:24px auto;max-width:1260px;padding:0 20px}h1{line-height:1.15}main{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:20px}article{border:1px solid #ddd;border-radius:14px;overflow:hidden;background:white}.frame{height:220px;background:#eee8dd;display:flex;align-items:center;justify-content:center}.neutral-photo{font-size:12px;color:#797b76;letter-spacing:.03em}.frame img{width:100%;height:100%;object-fit:contain}.copy{padding:18px}.tag{display:inline-block;border-radius:20px;background:#e8f0df;padding:4px 9px;font-size:11px}small{display:block;color:#66745b}a{color:#275b40}p{margin:8px 0}.notes{font-size:12px}button{padding:10px;border:1px solid #a3b59c;border-radius:6px;background:#eef3e7;cursor:pointer}@media(max-width:800px){main{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:480px){main{grid-template-columns:1fr}.frame{height:230px}}";
let html = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>FoodHub photo review and credits</title><style>${style}</style><h1>FoodHub photography review</h1><p>${unique.length} verified Pexels photographs across ${approved.length}/${rows.length} image slots. ${rows.length - approved.length} neutral placeholders.</p><p>Artwork preview only. The restaurants are fictional; no API requests or database changes occur here. Photographs show representative servings, not the actual menu products.</p><p><a href="ATTRIBUTION.md">Photographer credits and licence notes</a> · <a href="manifest.json">Asset manifest</a> · <a href="../gallery.html">Original illustrations</a></p><main>`;
for (const row of rows) {
  const photo = row.status === "photograph-verified";
  html += `<article><div class="frame">${photo ? `<img loading="lazy" src="${escape(link(row.path))}" alt="${escape(row.label)}">` : '<span class="neutral-photo">Photo unavailable</span>'}</div><div class="copy"><span class="tag">${photo ? "Verified photograph" : "Photo unavailable"}</span><h2>${escape(row.label)}</h2>${photo ? `<small>Photo: ${escape(row.photographer)}</small><p class="notes"><a href="${escape(row.source_url)}" rel="noreferrer">Source photograph</a> · <a href="${escape(row.licence_url)}">Pexels License</a></p><p class="notes">Representative serving; styling and portion may differ.</p>` : `<p class="notes">${escape(row.reason)}</p>`}<button type="button" data-test-fallback>Preview neutral placeholder</button></div></article>`;
}
html +=
  '</main><script>const placeholder=()=>{const node=document.createElement("span");node.className="neutral-photo";node.textContent="Photo unavailable";return node;};document.querySelectorAll("img").forEach(img=>img.addEventListener("error",()=>img.replaceWith(placeholder())));document.querySelectorAll("[data-test-fallback]").forEach(button=>button.addEventListener("click",()=>{const frame=button.closest("article").querySelector(".frame");frame.replaceChildren(placeholder());button.textContent="Neutral placeholder displayed"}));</script></html>';

await writeFile(resolve(dir, "gallery.html"), html + "\n");
console.log("Wrote offline photo gallery and photographer credits.");
