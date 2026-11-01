import type { Cart } from "../api/types";
import { money } from "../utils/format";
import { Totals } from "./ui";

export function BasketTotals({ cart }: { cart: Cart }) {
  if (typeof cart.delivery_fee === "number" && typeof cart.total === "number") {
    return (
      <Totals
        subtotal={cart.subtotal}
        delivery_fee={cart.delivery_fee}
        total={cart.total}
      />
    );
  }
  // Never invent a client-side fee while the old backend is still running.
  return (
    <>
      <dl className="totals">
        <div>
          <dt>Subtotal</dt>
          <dd>{money(cart.subtotal)}</dd>
        </div>
      </dl>
      <p className="muted">Delivery and final totals are shown at checkout.</p>
    </>
  );
}
