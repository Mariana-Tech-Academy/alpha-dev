import { test, beforeEach, after } from "node:test";
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { readFileSync } from "node:fs";
import ts from "typescript";
registerHooks({
  load(url, context, next) {
    if (url.endsWith("/src/utils/demo.ts"))
      return {
        format: "module",
        shortCircuit: true,
        source: ts.transpileModule(
          readFileSync(new URL(url), "utf8").replaceAll(
            "import.meta.env",
            "globalThis.__demoTestEnv",
          ),
          {
            compilerOptions: {
              module: ts.ModuleKind.ESNext,
              target: ts.ScriptTarget.ESNext,
            },
          },
        ).outputText,
      };
    return next(url, context);
  },
});
const { api } = await import("../src/api/foodhub.ts");
const original = globalThis.fetch;
let events, storage, writes;
beforeEach(() => {
  events = [];
  writes = 0;
  storage = new Map([["foodhub.token", "synthetic-customer-session"]]);
  globalThis.__demoTestEnv = { DEV: true, VITE_FOODHUB_DEMO_STATUS: "1" };
  globalThis.window = {
    location: { hostname: "localhost" },
    dispatchEvent: (event) => events.push(event.type),
  };
  globalThis.sessionStorage = {
    getItem: (key) => storage.get(key) ?? null,
    setItem: () => writes++,
    removeItem: () => writes++,
  };
  globalThis.localStorage = {
    setItem: () => {
      throw Error("staff data must never reach localStorage");
    },
  };
});
after(() => {
  globalThis.fetch = original;
  for (const key of [
    "__demoTestEnv",
    "window",
    "sessionStorage",
    "localStorage",
  ])
    delete globalThis[key];
});
function intact() {
  assert.equal(storage.get("foodhub.token"), "synthetic-customer-session");
  assert.equal(writes, 0);
  assert.deepEqual(events, []);
}
test("actual manager API helpers use separate credentials/JWT and abort signal without customer auth writes", async () => {
  const signal = new AbortController().signal;
  const calls = [];
  globalThis.fetch = async (url, options) => {
    calls.push([url, options]);
    assert.equal(options.signal, signal);
    return Response.json({
      success: true,
      data: url.endsWith("/login")
        ? {
            token: "synthetic.staff.token",
            user: { id: 4, role: "manager", status: "active" },
          }
        : { id: 4, role: "manager", status: "active" },
    });
  };
  await api.demoManagerLogin(
    "manager@example.test",
    "synthetic-password",
    signal,
  );
  await api.demoManagerMe("synthetic.staff.token", signal);
  await api.advanceDemoOrder(5, "pending", "synthetic.staff.token", signal);
  assert.equal(calls[0][0], "http://localhost:8080/api/auth/login");
  assert.equal(calls[0][1].headers.get("Authorization"), null);
  assert.deepEqual(JSON.parse(calls[0][1].body), {
    email: "manager@example.test",
    password: "synthetic-password",
  });
  assert.equal(calls[1][0], "http://localhost:8080/api/auth/me");
  assert.equal(
    calls[1][1].headers.get("Authorization"),
    "Bearer synthetic.staff.token",
  );
  assert.equal(calls[2][0], "http://localhost:8080/api/demo/orders/5/advance");
  assert.equal(
    calls[2][1].headers.get("Authorization"),
    "Bearer synthetic.staff.token",
  );
  assert.deepEqual(JSON.parse(calls[2][1].body), {
    expected_status: "pending",
  });
  intact();
});
test("manager login/profile rejection cannot log out or overwrite the customer", async () => {
  globalThis.fetch = async () =>
    Response.json(
      { success: false, message: "invalid or expired token" },
      { status: 401 },
    );
  await assert.rejects(
    api.demoManagerLogin("manager@example.test", "synthetic-password"),
    (error) => error.status === 401,
  );
  await assert.rejects(
    api.demoManagerMe("synthetic.staff.token"),
    (error) => error.status === 401,
  );
  intact();
});
test("production, missing opt-in or remote host prevents manager login/profile/status requests", () => {
  let calls = 0;
  globalThis.fetch = async () => {
    calls++;
    throw Error("must not fetch");
  };
  for (const env of [
    { DEV: false, VITE_FOODHUB_DEMO_STATUS: "1" },
    { DEV: true },
    {
      DEV: true,
      VITE_FOODHUB_DEMO_STATUS: "1",
      VITE_API_BASE_URL: "https://remote.example/api",
    },
  ]) {
    globalThis.__demoTestEnv = env;
    assert.throws(
      () => api.demoManagerLogin("manager@example.test", "synthetic-password"),
      /disabled/,
    );
    assert.throws(() => api.demoManagerMe("synthetic.staff.token"), /disabled/);
    assert.throws(
      () => api.advanceDemoOrder(5, "pending", "synthetic.staff.token"),
      /disabled/,
    );
  }
  assert.equal(calls, 0);
  intact();
});
