import { test } from "node:test";
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { readFileSync, existsSync } from "node:fs";
import { fileURLToPath } from "node:url";
import ts from "typescript";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

// Compile the real TSX component for Node's renderer; no browser or API writes.
registerHooks({
  resolve(specifier, context, nextResolve) {
    if (
      specifier.startsWith(".") &&
      context.parentURL?.includes("/frontend/src/")
    ) {
      for (const extension of [".ts", ".tsx"]) {
        const candidate = new URL(specifier + extension, context.parentURL);
        if (existsSync(fileURLToPath(candidate)))
          return nextResolve(candidate.href, context);
      }
    }
    return nextResolve(specifier, context);
  },
  load(url, context, nextLoad) {
    if (url.endsWith("/src/utils/demo.ts")) {
      return {
        format: "module",
        shortCircuit: true,
        source: ts.transpileModule(
          readFileSync(new URL(url), "utf8").replaceAll(
            "import.meta.env",
            "({DEV:true,VITE_FOODHUB_DEMO_STATUS:'1'})",
          ),
          {
            compilerOptions: {
              module: ts.ModuleKind.ESNext,
              target: ts.ScriptTarget.ESNext,
            },
          },
        ).outputText,
      };
    }
    if (url.includes("/frontend/src/") && url.endsWith(".tsx")) {
      return {
        format: "module",
        shortCircuit: true,
        source: ts.transpileModule(readFileSync(new URL(url), "utf8"), {
          compilerOptions: {
            jsx: ts.JsxEmit.ReactJSX,
            module: ts.ModuleKind.ESNext,
            target: ts.ScriptTarget.ESNext,
          },
        }).outputText,
      };
    }
    return nextLoad(url, context);
  },
});
const { OrderDetails } = await import("../src/components/OrderDetails.tsx");
const order = {
  id: 22,
  restaurant_id: 1,
  restaurant_name: "FoodHub Test Restaurant",
  order_number: "ORD-SAVED",
  status: "pending",
  created_at: "2026-11-01T12:00:00Z",
  items: [
    {
      name: "Purchased burger",
      quantity: 2,
      unit_price: 1250,
      line_total: 2500,
    },
  ],
  subtotal: 2500,
  delivery_fee: 100,
  total: 2600,
  delivery_name: "Saved Customer",
  delivery_phone: "07000000000",
  delivery_address: "12 Saved Street",
  delivery_city: "London",
  delivery_state: "Greater London",
  delivery_instructions: "Ring once",
};
test("expandable history renders actual item, money and delivery snapshots", () => {
  const html = renderToStaticMarkup(createElement(OrderDetails, { order }));
  for (const value of [
    "<details",
    "<summary",
    "2 × Purchased burger",
    "£12.50",
    "£25.00",
    "£26.00",
    "Saved Customer",
    "07000000000",
    "12 Saved Street",
    "London",
    "Greater London",
    "Ring once",
  ])
    assert.ok(html.includes(value), value);
  assert.ok(!html.includes("/restaurants/"));
  assert.ok(!html.includes("Cancel order"));
});
test("missing historical snapshots are explained and delivery strings are escaped", () => {
  const html = renderToStaticMarkup(
    createElement(OrderDetails, {
      order: {
        ...order,
        items: [],
        delivery_name: "<script>alert(1)</script>",
        delivery_phone: "",
        delivery_address: "",
        delivery_city: "",
        delivery_state: "",
        delivery_instructions: "",
      },
    }),
  );
  assert.ok(html.includes("Item details were not saved"));
  assert.ok(html.includes("&lt;script&gt;"));
  assert.ok(!html.includes("<script>"));
  const empty = renderToStaticMarkup(
    createElement(OrderDetails, {
      order: {
        ...order,
        items: [],
        delivery_name: "",
        delivery_phone: "",
        delivery_address: "",
        delivery_city: "",
        delivery_state: "",
        delivery_instructions: "",
      },
    }),
  );
  assert.ok(empty.includes("Delivery details were not saved"));
});

test("demo form shows email, masks only password, excludes JWT paste/reveal and terminal orders", async () => {
  const { LocalDemoStatus } =
    await import("../src/components/LocalDemoStatus.tsx");
  const { AuthContext } = await import("../src/auth/context.ts");
  globalThis.window = { location: { hostname: "localhost" } };
  const render = (orders) =>
    renderToStaticMarkup(
      createElement(
        AuthContext.Provider,
        { value: { user: { id: 3 } } },
        createElement(LocalDemoStatus, { orders, onUpdated: () => {} }),
      ),
    );
  try {
    const html = render([order]);
    assert.ok(html.includes("LOCAL-DEMO-ONLY"));
    assert.ok(html.includes("Advance Status → In Progress"));
    assert.match(
      html,
      /<input(?=[^>]*type="email")(?=[^>]*name="demo-manager-email")[^>]*>/,
    );
    assert.match(
      html,
      /<input(?=[^>]*type="password")(?=[^>]*name="demo-manager-password")[^>]*>/,
    );
    assert.equal(html.match(/type="password"/g)?.length, 1);
    assert.ok(html.includes("Sign in as manager"));
    assert.ok(!html.includes("Manager/admin JWT"));
    assert.ok(!html.includes("Reveal password"));
    assert.ok(html.includes('disabled=""'));
    assert.equal(order.status, "pending");
    for (const status of ["delivered", "cancelled"]) {
      const done = render([{ ...order, status }]);
      assert.ok(done.includes("No orders on this page can be advanced"));
      assert.ok(!done.includes('type="password"'));
    }
  } finally {
    delete globalThis.window;
  }
});

test("Basket displays the backend subtotal, one delivery fee and grand total in GBP", async () => {
  const { BasketTotals } = await import("../src/components/BasketTotals.tsx");
  const cart = {
    cart_id: 4,
    items: [],
    subtotal: 3850,
    delivery_fee: 299,
    total: 4149,
  };
  const html = renderToStaticMarkup(createElement(BasketTotals, { cart }));
  for (const value of [
    "Subtotal",
    "Delivery",
    "Total",
    "£38.50",
    "£2.99",
    "£41.49",
  ])
    assert.ok(html.includes(value), value);
  assert.equal(html.match(/£2\.99/g)?.length, 1);
  const empty = renderToStaticMarkup(
    createElement(BasketTotals, {
      cart: { ...cart, subtotal: 0, delivery_fee: 0, total: 0 },
    }),
  );
  assert.equal(empty.match(/£0\.00/g)?.length, 3);
});
test("Basket does not invent fees while the existing backend awaits approved activation", async () => {
  const { BasketTotals } = await import("../src/components/BasketTotals.tsx");
  const html = renderToStaticMarkup(
    createElement(BasketTotals, {
      cart: { cart_id: 4, items: [], subtotal: 2500 },
    }),
  );
  assert.ok(html.includes("£25.00"));
  assert.ok(html.includes("shown at checkout"));
  assert.ok(!html.includes("£2.99"));
  assert.ok(!html.includes("NaN"));
});
test("My Orders renders saved new and historical delivery fees without repricing", () => {
  const historical = renderToStaticMarkup(
    createElement(OrderDetails, {
      order: { ...order, subtotal: 2500, delivery_fee: 0, total: 2500 },
    }),
  );
  assert.ok(historical.includes("£0.00"));
  assert.ok(!historical.includes("£2.99"));
  assert.ok(!historical.includes("£27.99"));
  const current = renderToStaticMarkup(
    createElement(OrderDetails, {
      order: { ...order, subtotal: 2500, delivery_fee: 299, total: 2799 },
    }),
  );
  for (const value of ["£25.00", "£2.99", "£27.99"])
    assert.ok(current.includes(value));
});
