import { test, beforeEach, after } from "node:test";
import assert from "node:assert/strict";
import { request, session, ApiError } from "../src/api/client.ts";
import { money } from "../src/utils/format.ts";
const originalFetch = globalThis.fetch;
let storage;
let events;
beforeEach(() => {
  storage = new Map();
  events = [];
  globalThis.sessionStorage = {
    getItem: (key) => storage.get(key) ?? null,
    setItem: (key, value) => storage.set(key, value),
    removeItem: (key) => storage.delete(key),
  };
  globalThis.window = { dispatchEvent: (event) => events.push(event.type) };
});
after(() => {
  globalThis.fetch = originalFetch;
  delete globalThis.sessionStorage;
  delete globalThis.window;
});
test("sends JSON and bearer token, unwraps the API envelope", async () => {
  session.set("customer-token");
  globalThis.fetch = async (url, options) => {
    assert.equal(url, "http://localhost:8080/api/cart/items");
    assert.equal(options.headers.get("Authorization"), "Bearer customer-token");
    assert.equal(options.headers.get("Content-Type"), "application/json");
    assert.deepEqual(JSON.parse(options.body), {
      menu_item_id: 7,
      quantity: 1,
    });
    return Response.json({ success: true, message: "added", data: { id: 3 } });
  };
  assert.deepEqual(
    await request(
      "/cart/items",
      {
        method: "POST",
        body: JSON.stringify({ menu_item_id: 7, quantity: 1 }),
      },
      true,
    ),
    { id: 3 },
  );
});
test("logout 204 succeeds without parsing JSON", async () => {
  globalThis.fetch = async () => new Response(null, { status: 204 });
  assert.equal(
    await request("/auth/logout", { method: "POST" }, true),
    undefined,
  );
});
test("protected 401 clears session and announces expiry", async () => {
  session.set("expired");
  globalThis.fetch = async () =>
    Response.json({ success: false, message: "expired" }, { status: 401 });
  await assert.rejects(
    request("/cart", {}, true),
    (error) => error.status === 401,
  );
  assert.equal(session.get(), null);
  assert.deepEqual(events, ["foodhub:expired"]);
});
test("a late 401 does not clear a newer session", async () => {
  session.set("old");
  globalThis.fetch = async () => {
    session.set("new");
    return Response.json({ success: false }, { status: 401 });
  };
  await assert.rejects(request("/cart", {}, true));
  assert.equal(session.get(), "new");
  assert.deepEqual(events, []);
});
test("invalid login does not trigger session-expired UI", async () => {
  globalThis.fetch = async () =>
    Response.json(
      { success: false, message: "invalid email or password" },
      { status: 401 },
    );
  await assert.rejects(
    request("/auth/login", { method: "POST" }),
    /invalid email or password/,
  );
  assert.deepEqual(events, []);
});
test("preserves backend field validation errors", async () => {
  globalThis.fetch = async () =>
    Response.json(
      {
        success: false,
        message: "check delivery details",
        errors: { delivery_city: "city is required" },
      },
      { status: 422 },
    );
  await assert.rejects(
    request("/orders", {}, true),
    (error) =>
      error instanceof ApiError &&
      error.fields.delivery_city === "city is required",
  );
});
test("network errors are actionable and aborts remain aborts", async () => {
  globalThis.fetch = async () => {
    throw new TypeError("Failed to fetch");
  };
  await assert.rejects(request("/restaurants"), /Cannot reach FoodHub/);
  globalThis.fetch = async () => {
    throw new DOMException("Aborted", "AbortError");
  };
  await assert.rejects(
    request("/restaurants"),
    (error) => error.name === "AbortError",
  );
});
test("GBP uses minor units, including zero delivery", () => {
  assert.equal(money(1250), "£12.50");
  assert.equal(money(0), "£0.00");
});
