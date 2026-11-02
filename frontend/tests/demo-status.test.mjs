import { test } from "node:test";
import assert from "node:assert/strict";
import { isLocalDemoEnabled, demoNextStatus } from "../src/utils/demo.ts";
import { statusLabel } from "../src/utils/format.ts";

test("four customer-visible stages map to native statuses and stop at completion", () => {
  let status = "pending";
  const stages = [statusLabel(status)];
  while (demoNextStatus(status)) {
    status = demoNextStatus(status);
    stages.push(statusLabel(status));
  }
  assert.deepEqual(stages, [
    "Pending",
    "In Progress",
    "Out for Delivery",
    "Complete",
  ]);
  assert.equal(status, "delivered");
  assert.equal(demoNextStatus("cancelled"), undefined);
  assert.equal(demoNextStatus("confirmed"), "preparing");
  assert.equal(demoNextStatus("ready_for_delivery"), "out_for_delivery");
});

test("demo UI requires opt-in development mode and loopback frontend and API", () => {
  assert.equal(
    isLocalDemoEnabled(true, "1", "localhost", "http://localhost:8080/api"),
    true,
  );
  assert.equal(
    isLocalDemoEnabled(true, "1", "127.0.0.1", "http://[::1]:8080/api"),
    true,
  );
  for (const args of [
    [false, "1", "localhost", "http://localhost:8080/api"],
    [true, undefined, "localhost", "http://localhost:8080/api"],
    [true, "0", "localhost", "http://localhost:8080/api"],
    [true, "1", "foodhub.example", "http://localhost:8080/api"],
    [true, "1", "localhost", "https://api.foodhub.example/api"],
    [true, "1", "localhost", "invalid"],
  ])
    assert.equal(isLocalDemoEnabled(...args), false);
});
