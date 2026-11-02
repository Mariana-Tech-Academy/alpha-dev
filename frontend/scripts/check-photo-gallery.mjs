// Optional local Chrome QA. Opens only the file-based asset gallery; no API/DB.
import { spawn } from "node:child_process";
import { mkdtemp, readFile, writeFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import assert from "node:assert/strict";
const chrome = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const profile = await mkdtemp(join(tmpdir(), "foodhub-photo-qa-"));
const gallery = fileURLToPath(
  new URL("../public/demo-images/photos/gallery.html", import.meta.url),
);
const browser = spawn(
  chrome,
  [
    "--headless=new",
    "--remote-debugging-port=0",
    "--user-data-dir=" + profile,
    "--no-first-run",
    "--no-default-browser-check",
    "about:blank",
  ],
  { stdio: "ignore" },
);
let socket;
const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
try {
  let port;
  for (let i = 0; i < 100; i++) {
    try {
      port = (
        await readFile(join(profile, "DevToolsActivePort"), "utf8")
      ).split("\n")[0];
      break;
    } catch {
      await pause(100);
    }
  }
  if (!port) throw new Error("Chrome debugging endpoint unavailable");
  const target = await (
    await fetch(`http://127.0.0.1:${port}/json/new?about:blank`, {
      method: "PUT",
    })
  ).json();
  socket = new WebSocket(target.webSocketDebuggerUrl);
  await new Promise((resolve, reject) => {
    socket.onopen = resolve;
    socket.onerror = reject;
  });
  let serial = 0;
  const pending = new Map();
  socket.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (pending.has(message.id)) {
      const { resolve, reject } = pending.get(message.id);
      pending.delete(message.id);
      message.error
        ? reject(new Error(JSON.stringify(message.error)))
        : resolve(message.result);
    }
  };
  const send = (method, params = {}) =>
    new Promise((resolve, reject) => {
      const id = ++serial;
      pending.set(id, { resolve, reject });
      socket.send(JSON.stringify({ id, method, params }));
    });
  const evaluate = async (expression) => {
    const result = await send("Runtime.evaluate", {
      expression,
      awaitPromise: true,
      returnByValue: true,
    });
    if (result.exceptionDetails)
      throw new Error(JSON.stringify(result.exceptionDetails));
    return result.result.value;
  };
  await send("Page.enable");
  for (const [label, width, height, mobile] of [
    ["desktop", 1280, 900, false],
    ["mobile", 390, 844, true],
  ]) {
    await send("Emulation.setDeviceMetricsOverride", {
      width,
      height,
      deviceScaleFactor: 1,
      mobile,
    });
    await send("Page.navigate", { url: pathToFileURL(gallery).href });
    for (let i = 0; i < 100; i++) {
      if (
        await evaluate(
          'document.readyState === "complete" && !!document.querySelector("main")',
        )
      )
        break;
      await pause(50);
    }
    const result = await evaluate(`(async()=>{
   const images=[...document.querySelectorAll('img')];images.forEach(img=>img.loading='eager');
   await Promise.all(images.map(img=>img.decode().catch(()=>{})));
   return {width:innerWidth,scrollWidth:document.documentElement.scrollWidth,count:images.length,failed:images.filter(img=>!img.complete||img.naturalWidth===0).length,photos:images.filter(img=>img.getAttribute('src').includes('.jpg')).length};
  })()`);
    assert.equal(result.count, 48);
    assert.equal(result.failed, 0);
    assert.equal(result.photos, 48);
    assert.ok(result.scrollWidth <= result.width);
    const screenshot = await send("Page.captureScreenshot", {
      format: "png",
      captureBeyondViewport: false,
    });
    const screenshotPath = join(tmpdir(), `foodhub-photos-${label}.png`);
    await writeFile(screenshotPath, Buffer.from(screenshot.data, "base64"));
    const fallback = await evaluate(
      `(()=>{const button=document.querySelector('[data-test-fallback]');const frame=button.closest('article').querySelector('.frame');button.click();return {neutral:!frame.querySelector('img') && frame.textContent.includes('Photo unavailable')};})()`,
    );
    assert.equal(fallback.neutral, true);
    console.log(
      JSON.stringify({
        viewport: label,
        ...result,
        fallback: "PASS",
        screenshot: screenshotPath,
      }),
    );
  }
  await send("Browser.close").catch(() => {});
  console.log(
    "PASS desktop/mobile gallery images, no horizontal overflow and neutral placeholder. No backend/API access.",
  );
} finally {
  if (socket) socket.close();
  if (browser.exitCode === null) {
    browser.kill("SIGTERM");
    await Promise.race([
      new Promise((resolve) => browser.once("exit", resolve)),
      pause(3000),
    ]);
  }
  // This profile is uniquely created above; no existing browser profiles touched.
  if (browser.exitCode !== null)
    await rm(profile, { recursive: true, force: true });
}
