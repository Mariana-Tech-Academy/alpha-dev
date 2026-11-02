// Isolated manager-login UI smoke test. ALL backend HTTP requests mocked and blocked.
// Requires existing Vite; never starts/restarts services or uses real credentials.
import { spawn } from "node:child_process";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import assert from "node:assert/strict";
const chrome = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const profile = await mkdtemp(join(tmpdir(), "foodhub-manager-qa-"));
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
  let liveBackendRequests = 0;
  socket.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (
      message.method === "Network.requestWillBeSent" &&
      /:8080\//.test(message.params.request.url)
    )
      liveBackendRequests++;
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
  await send("Network.enable");
  await send("Network.setBlockedURLs", {
    urls: ["*://localhost:8080/*", "*://127.0.0.1:8080/*", "*://[::1]:8080/*"],
  });
  await send("Page.addScriptToEvaluateOnNewDocument", {
    source: `(${function () {
      const realFetch = window.fetch.bind(window);
      const customer = {
        id: 3,
        role: "customer",
        status: "active",
        first_name: "Demo",
        last_name: "Customer",
        created_at: "2026-11-01T12:00:00Z",
      };
      const manager = { id: 4, role: "manager", status: "active" };
      const token = "synthetic-staff-session";
      sessionStorage.setItem("foodhub.token", "synthetic-customer-session");
      const originalSet = Storage.prototype.setItem;
      const stats = (window.__managerUI = {
        mode: "success",
        login: 0,
        me: 0,
        advance: 0,
        history: 0,
        logout: 0,
        status: "pending",
        writes: [],
      });
      Storage.prototype.setItem = function (key, value) {
        stats.writes.push(key);
        return originalSet.call(this, key, value);
      };
      const answer = (data, status = 200) =>
        Promise.resolve(
          Response.json(
            {
              success: status === 200,
              data,
              message: status === 200 ? "ok" : "rejected",
            },
            { status },
          ),
        );
      window.fetch = (input, options = {}) => {
        const url = new URL(
          typeof input === "string" ? input : input.url,
          location.origin,
        );
        if (!url.pathname.startsWith("/api/")) return realFetch(input, options);
        const auth = new Headers(options.headers).get("Authorization");
        if (url.pathname === "/api/auth/login") {
          stats.login++;
          if (stats.mode === "invalid") return answer(null, 401);
          if (stats.mode === "late")
            return new Promise((resolve) => {
              stats.release = () =>
                resolve(
                  Response.json({
                    success: true,
                    data: { token, user: manager },
                  }),
                );
            });
          return answer({ token, user: manager });
        }
        if (url.pathname === "/api/auth/me") {
          if (auth === "Bearer synthetic-customer-session")
            return answer(customer);
          stats.me++;
          return answer(
            stats.mode === "wrong-role"
              ? { ...manager, role: "customer" }
              : manager,
          );
        }
        if (url.pathname === "/api/restaurants/2")
          return answer({
            id: 2,
            manager_id: stats.mode === "wrong-owner" ? 8 : 4,
            status: "active",
          });
        if (url.pathname === "/api/orders") {
          stats.history++;
          return answer({
            orders: [
              {
                id: 5,
                order_number: "TEST-ONLY",
                restaurant_id: 2,
                restaurant_name: "Test fixture",
                status: stats.status,
                created_at: "2026-11-01T12:00:00Z",
                items: [],
                subtotal: 1000,
                delivery_fee: 299,
                total: 1299,
              },
            ],
            page: 1,
            limit: 10,
            total: 1,
          });
        }
        if (url.pathname === "/api/demo/orders/5/advance") {
          stats.advance++;
          if (auth !== "Bearer " + token) return answer(null, 401);
          stats.status =
            stats.status === "pending" ? "preparing" : "out_for_delivery";
          return answer({ order_number: "TEST-ONLY", status: stats.status });
        }
        if (url.pathname === "/api/auth/logout") {
          stats.logout++;
          return answer(null);
        }
        throw Error("Unexpected API request blocked by test fixture");
      };
    }.toString()})()`,
  });
  const waitFor = async (expression) => {
    for (let i = 0; i < 100; i++) {
      if (await evaluate(expression)) return;
      await pause(50);
    }
    throw Error("UI fixture condition timed out: " + expression);
  };
  await send("Page.navigate", { url: "http://localhost:5173/orders" });
  await waitFor(`!!document.querySelector('[name="demo-manager-email"]')`);
  assert.equal(
    await evaluate(
      `document.querySelector('[name="demo-manager-email"]').type`,
    ),
    "email",
  );
  assert.equal(
    await evaluate(
      `document.querySelector('[name="demo-manager-password"]').type`,
    ),
    "password",
  );
  const submit = async (mode) => {
    await evaluate(`window.__managerUI.mode=${JSON.stringify(mode)};
      document.querySelector('[name="demo-manager-email"]').value='manager@example.test';
      document.querySelector('[name="demo-manager-password"]').value='synthetic-password';
      document.querySelector('[name="demo-manager-password"]').form.requestSubmit();`);
  };
  const hasManager = `[...document.querySelectorAll('[role="status"]')].some(node=>node.textContent==='Manager authenticated')`;
  const blank = `document.querySelector('[name="demo-manager-email"]')?.value==='' && document.querySelector('[name="demo-manager-password"]')?.value===''`;
  const click = async (label) =>
    evaluate(
      `[...document.querySelectorAll('button')].find(button=>button.textContent===${JSON.stringify(label)}).click()`,
    );
  await submit("success");
  await waitFor(hasManager);
  assert.equal(
    await evaluate(
      `sessionStorage.getItem('foodhub.token')==='synthetic-customer-session' && localStorage.length===0 && window.__managerUI.writes.length===0`,
    ),
    true,
  );
  assert.equal(
    await evaluate(
      `document.body.textContent.includes('synthetic-staff-session')`,
    ),
    false,
  );
  let history = await evaluate("window.__managerUI.history");
  await click("Advance Status → In Progress");
  await waitFor(
    `window.__managerUI.history>${history} && [...document.querySelectorAll('button')].some(b=>b.textContent==='Advance Status → Out for Delivery'&&!b.disabled)`,
  );
  history = await evaluate("window.__managerUI.history");
  await click("Advance Status → Out for Delivery");
  await waitFor(
    `window.__managerUI.history>${history} && [...document.querySelectorAll('button')].some(b=>b.textContent==='Advance Status → Complete'&&!b.disabled)`,
  );
  assert.equal(await evaluate("window.__managerUI.login"), 1);
  assert.equal(await evaluate("window.__managerUI.advance"), 2);
  await click("Manager logout");
  await waitFor(blank);
  assert.equal(await evaluate("window.__managerUI.logout"), 0);
  for (const mode of ["invalid", "wrong-role", "wrong-owner"]) {
    await submit(mode);
    await waitFor(
      `!!document.querySelector('[role="alert"]') && [...document.querySelectorAll('button')].some(b=>b.textContent==='Sign in as manager'&&!b.disabled)`,
    );
    assert.equal(await evaluate(blank), true);
    assert.equal(
      await evaluate(
        `sessionStorage.getItem('foodhub.token')==='synthetic-customer-session'`,
      ),
      true,
    );
    assert.equal(await evaluate("window.__managerUI.advance"), 2);
  }
  await submit("late");
  await waitFor(`typeof window.__managerUI.release==='function'`);
  await click("Cancel manager login");
  await evaluate("window.__managerUI.release()");
  await pause(150);
  assert.equal(await evaluate(hasManager), false);
  assert.equal(await evaluate(blank), true);
  assert.equal(await evaluate("window.__managerUI.writes.length"), 0);
  await submit("success");
  await waitFor(hasManager);
  await send("Page.reload", { ignoreCache: true });
  await waitFor(`!!document.querySelector('[name="demo-manager-email"]')`);
  assert.equal(await evaluate(hasManager), false);
  assert.equal(await evaluate(blank), true);
  assert.equal(await evaluate("window.__managerUI.login"), 0);
  await submit("success");
  await waitFor(hasManager);
  await evaluate(`window.dispatchEvent(new Event('foodhub:expired'))`);
  await waitFor(`location.pathname==='/login'`);
  assert.equal(await evaluate(hasManager), false);
  assert.equal(
    await evaluate(
      `!!document.querySelector('[name="demo-manager-password"]')`,
    ),
    false,
  );
  assert.equal(liveBackendRequests, 0);
  await send("Browser.close").catch(() => {});
  console.log(
    "PASS isolated browser: visible email, masked/cleared password, verified manager UI, token reuse, two mocked advances with history refresh, logout, role/ownership/401 rejection, late cancellation, refresh/customer-session cleanup, unchanged customer storage during staff operations. Zero real backend requests.",
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
  if (browser.exitCode !== null)
    await rm(profile, { recursive: true, force: true });
}
