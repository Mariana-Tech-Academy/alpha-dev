import { useCallback, useState } from "react";
import { Link } from "react-router-dom";
import { BasketTotals } from "../components/BasketTotals";
import { api } from "../api/foodhub";
import { useResource } from "../hooks/useResource";
import { Empty, ErrorMessage, Loading } from "../components/ui";
import { money } from "../utils/format";
import type { Cart as CartData } from "../api/types";
export function Cart() {
  const load = useCallback((signal: AbortSignal) => api.cart(signal), []);
  const { data, error, loading, retry } = useResource(load);
  const [updated, setUpdated] = useState<CartData | null>(null);
  const [busy, setBusy] = useState<number | null>(null);
  const [actionError, setActionError] = useState<unknown>(null);
  const cart = updated || data;
  async function change(id: number, quantity?: number) {
    setBusy(id);
    setActionError(null);
    try {
      setUpdated(
        await (quantity !== undefined
          ? api.updateQuantity(id, quantity)
          : id === 0
            ? api.clearCart()
            : api.removeItem(id)),
      );
    } catch (err) {
      setActionError(err);
    } finally {
      setBusy(null);
    }
  }
  if (loading) return <Loading label="Loading your basket…" />;
  if (error || !cart) return <ErrorMessage error={error} retry={retry} />;
  if (!cart.items.length)
    return (
      <Empty title="Your basket is waiting">
        <p>A delicious meal starts with a little browsing.</p>
        <Link className="button" to="/">
          Explore restaurants →
        </Link>
      </Empty>
    );
  const blocked = cart.items.some((item) => !item.is_available);
  return (
    <section className="section">
      <span className="eyebrow">SOMETHING GOOD IS COMING</span>
      <h1>Your basket</h1>
      <div className="two-column">
        <div>
          <ErrorMessage error={actionError} />
          <div className="panel cart-lines">
            {cart.items.map((item) => (
              <article className="cart-line" key={item.id}>
                <div className="line-main">
                  <h3>{item.name || "Unavailable menu item"}</h3>
                  <p>{money(item.locked_price)} each</p>
                  {item.price_changed && (
                    <p className="notice">
                      Price changed: now {money(item.price)}. Your saved price
                      is {money(item.locked_price)}.
                    </p>
                  )}
                  {!item.is_available && (
                    <p className="notice">
                      This item is currently unavailable and cannot be checked
                      out.
                    </p>
                  )}
                  {item.notice && !item.price_changed && item.is_available && (
                    <p className="notice">{item.notice}</p>
                  )}
                </div>
                <div className="quantity">
                  <button
                    className="secondary"
                    aria-label={`Decrease ${item.name} quantity`}
                    disabled={busy !== null || item.quantity <= 1}
                    onClick={() => change(item.id, item.quantity - 1)}
                  >
                    −
                  </button>
                  <span aria-live="polite">
                    {busy === item.id ? "…" : item.quantity}
                  </span>
                  <button
                    className="secondary"
                    aria-label={`Increase ${item.name} quantity`}
                    disabled={busy !== null || item.quantity >= 99}
                    onClick={() => change(item.id, item.quantity + 1)}
                  >
                    +
                  </button>
                </div>
                <button
                  className="secondary"
                  disabled={busy !== null}
                  onClick={() => change(item.id)}
                  aria-label={`Remove ${item.name} from basket`}
                >
                  Remove
                </button>
                <strong>{money(item.line_total)}</strong>
              </article>
            ))}
          </div>
          <p className="subtle-note">
            Quantities can be changed from 1 to 99. Removing basket items does
            not change placed orders.
          </p>
          <button
            className="secondary"
            disabled={busy !== null}
            onClick={() => change(0)}
          >
            Clear basket
          </button>
          <Link className="back-link" to={`/restaurants/${cart.restaurant_id}`}>
            ← Back to the menu
          </Link>
        </div>
        <aside className="panel order-summary">
          <h2>Basket summary</h2>
          <BasketTotals cart={cart} />
          {blocked ? (
            <p className="notice">
              Checkout is unavailable while your basket contains unavailable
              items. Remove unavailable items to continue.
            </p>
          ) : busy !== null ? (
            <button disabled>Updating basket…</button>
          ) : (
            <Link className="button" to="/checkout">
              Continue to checkout →
            </Link>
          )}
        </aside>
      </div>
    </section>
  );
}
