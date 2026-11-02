import { test } from "node:test";
import assert from "node:assert/strict";
import { createDemoManagerSession } from "../src/utils/demoManagerSession.ts";
const profile = { id: 4, role: "manager", status: "active" };
const signedIn = { user: profile, token: "synthetic.staff.token" };
const restaurant = { id: 2, manager_id: 4, status: "active" };
const order = { id: 5, restaurant_id: 2, status: "pending" };
const deferred = () => {
  let resolve;
  const promise = new Promise((r) => (resolve = r));
  return { promise, resolve };
};
function fixture(overrides = {}) {
  const calls = [];
  const context = { key: "customer-session", enabled: true };
  const api = {
    login: async (...args) => {
      calls.push(["login", ...args]);
      return signedIn;
    },
    me: async (...args) => {
      calls.push(["me", ...args]);
      return profile;
    },
    restaurant: async (...args) => {
      calls.push(["restaurant", ...args]);
      return restaurant;
    },
    advance: async (...args) => {
      calls.push(["advance", ...args]);
      return { order_number: "ORD-SAVED", status: "preparing" };
    },
    ...overrides,
  };
  const manager = createDemoManagerSession(api, {
    enabled: () => context.enabled,
    customerSession: () => context.key,
  });
  return { manager, calls, context };
}
test("verifies profile/ownership and reuses private staff token for advances and refresh", async () => {
  const { manager, calls, context } = fixture();
  const identity = await manager.login(
    " manager@example.test ",
    "synthetic-password",
    2,
  );
  assert.deepEqual(identity, { id: 4, restaurantId: 2 });
  assert.deepEqual(manager.getIdentity(), identity);
  assert.ok(!JSON.stringify(manager).includes(signedIn.token));
  assert.equal(calls[0][1], "manager@example.test");
  assert.equal(calls[1][1], signedIn.token);
  let refreshed = 0;
  await manager.advance(order, () => refreshed++);
  await manager.advance({ ...order, status: "preparing" }, () => refreshed++);
  assert.equal(calls.filter((c) => c[0] === "login").length, 1);
  assert.equal(calls.filter((c) => c[0] === "me").length, 1);
  assert.equal(calls.filter((c) => c[0] === "advance").length, 2);
  assert.ok(
    calls
      .filter((c) => c[0] === "advance")
      .every((c) => c[3] === signedIn.token),
  );
  assert.equal(refreshed, 2);
  assert.equal(context.key, "customer-session");
  manager.clear();
  assert.equal(manager.getIdentity(), undefined);
  assert.equal(fixture().manager.getIdentity(), undefined); // a new page/instance has no session
});
test("rejects wrong role, inactive user, mismatched identity or restaurant ownership", async () => {
  for (const overrides of [
    { me: async () => ({ ...profile, role: "customer" }) },
    { me: async () => ({ ...profile, role: "admin" }) },
    { me: async () => ({ ...profile, status: "suspended" }) },
    { me: async () => ({ ...profile, id: 9 }) },
    { restaurant: async () => ({ ...restaurant, manager_id: 9 }) },
    { restaurant: async () => ({ ...restaurant, id: 9 }) },
    { restaurant: async () => ({ ...restaurant, status: "inactive" }) },
  ]) {
    const { manager, calls } = fixture(overrides);
    await assert.rejects(
      manager.login("manager@example.test", "synthetic-password", 2),
      /Manager authentication failed/,
    );
    assert.equal(manager.getIdentity(), undefined);
    assert.equal(calls.filter((c) => c[0] === "advance").length, 0);
  }
});
test("login errors expose no secret/server payload and leave no staff session", async () => {
  const { manager } = fixture({
    login: async () => {
      throw Object.assign(new Error(signedIn.token), { status: 401 });
    },
  });
  await assert.rejects(
    manager.login("manager@example.test", "synthetic-password", 2),
    (err) => !err.message.includes(signedIn.token),
  );
  assert.equal(manager.getIdentity(), undefined);
});
test("logout/unmount ignores late login even when the API ignores abort", async () => {
  const task = deferred();
  const { manager, calls } = fixture({ login: () => task.promise });
  const login = manager.login("manager@example.test", "synthetic-password", 2);
  // Rendering during a pending login must not accidentally abort the attempt.
  assert.equal(manager.getIdentity(), undefined);
  manager.clear();
  task.resolve(signedIn);
  assert.equal(await login, undefined);
  assert.equal(manager.getIdentity(), undefined);
  assert.equal(calls.length, 0);
});
test("late older login cannot replace a newer manager session", async () => {
  const task = deferred();
  let attempts = 0;
  const { manager } = fixture({
    login: () => (++attempts === 1 ? task.promise : Promise.resolve(signedIn)),
  });
  const first = manager.login("manager@example.test", "synthetic-password", 2);
  await manager.login("manager@example.test", "synthetic-password", 2);
  task.resolve({ user: { ...profile, id: 8 }, token: "other.synthetic.token" });
  assert.equal(await first, undefined);
  assert.deepEqual(manager.getIdentity(), { id: 4, restaurantId: 2 });
});
test("customer session changes and demo disablement clear manager authentication", async () => {
  for (const change of [
    (c) => (c.key = "different-customer"),
    (c) => (c.key = null),
    (c) => (c.enabled = false),
  ]) {
    const { manager, context } = fixture();
    await manager.login("manager@example.test", "synthetic-password", 2);
    change(context);
    assert.equal(manager.getIdentity(), undefined);
    await assert.rejects(
      manager.advance(order, () => {}),
      /Sign in/,
    );
  }
  const { manager, context, calls } = fixture();
  context.enabled = false;
  await assert.rejects(
    manager.login("manager@example.test", "synthetic-password", 2),
  );
  assert.equal(calls.length, 0);
});
test("customer changes during login discard late credentials", async () => {
  const task = deferred();
  const { manager, context } = fixture({ me: () => task.promise });
  const login = manager.login("manager@example.test", "synthetic-password", 2);
  await Promise.resolve();
  context.key = "new-customer";
  task.resolve(profile);
  assert.equal(await login, undefined);
  assert.equal(manager.getIdentity(), undefined);
});
test("401/403 clears only staff token; failed transitions do not refresh orders", async () => {
  for (const status of [401, 403, 409]) {
    const { manager, context } = fixture({
      advance: async () => {
        throw Object.assign(new Error("private server message"), { status });
      },
    });
    await manager.login("manager@example.test", "synthetic-password", 2);
    let refresh = 0;
    await assert.rejects(manager.advance(order, () => refresh++));
    assert.equal(refresh, 0);
    assert.equal(context.key, "customer-session");
    assert.equal(!!manager.getIdentity(), status === 409);
  }
});
test("wrong restaurant and ownership drift cannot send status writes", async () => {
  const { manager, calls } = fixture();
  await manager.login("manager@example.test", "synthetic-password", 2);
  await assert.rejects(
    manager.advance({ ...order, restaurant_id: 3 }, () => {}),
    /Sign in/,
  );
  assert.equal(calls.filter((c) => c[0] === "advance").length, 0);
  let lookups = 0;
  const drift = fixture({
    restaurant: async () =>
      ++lookups === 1 ? restaurant : { ...restaurant, manager_id: 8 },
  });
  await drift.manager.login("manager@example.test", "synthetic-password", 2);
  await assert.rejects(drift.manager.advance(order, () => {}));
  assert.equal(drift.manager.getIdentity(), undefined);
  assert.equal(drift.calls.filter((c) => c[0] === "advance").length, 0);
});
test("logout while an advance is pending prevents late UI updates and refresh", async () => {
  const task = deferred();
  const { manager } = fixture({ advance: () => task.promise });
  await manager.login("manager@example.test", "synthetic-password", 2);
  let refresh = 0;
  const attempt = manager.advance(order, () => refresh++);
  await Promise.resolve();
  manager.clear();
  task.resolve({ status: "preparing", order_number: "ORD-SAVED" });
  assert.equal(await attempt, undefined);
  assert.equal(refresh, 0);
  assert.equal(manager.getIdentity(), undefined);
});

test("login diagnostics identify failure stages without exposing API secrets", async () => {
  const reject = (status) => async () => {
    throw Object.assign(new Error("synthetic-private-payload"), { status });
  };
  for (const [stage, overrides, http] of [
    ["login request", { login: reject(401) }, " (HTTP 401)"],
    ["profile request", { me: reject(403) }, " (HTTP 403)"],
    ["identity check", { me: async () => ({ ...profile, id: 9 }) }, ""],
    ["restaurant request", { restaurant: reject(undefined) }, ""],
    [
      "ownership check",
      { restaurant: async () => ({ ...restaurant, manager_id: 9 }) },
      "",
    ],
  ]) {
    const { manager, context, calls } = fixture(overrides);
    await assert.rejects(
      manager.login("manager@example.test", "synthetic-password", 2),
      (error) => {
        assert.equal(
          error.message,
          `Manager authentication failed at ${stage}${http}.`,
        );
        assert.equal(error.cause, undefined);
        return true;
      },
    );
    assert.equal(manager.getIdentity(), undefined);
    assert.equal(context.key, "customer-session");
    assert.equal(calls.filter((c) => c[0] === "advance").length, 0);
  }
});
