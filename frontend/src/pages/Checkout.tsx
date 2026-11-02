import { useCallback, useRef, useState } from "react";
import type { FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api/foodhub";
import type { DeliveryRequest } from "../api/types";
import { useAuth } from "../auth/context";
import { useResource } from "../hooks/useResource";
import { ErrorMessage, Loading, Totals } from "../components/ui";
import { money } from "../utils/format";
export function Checkout() {
  const load = useCallback((signal: AbortSignal) => api.summary(signal), []);
  const { data, error, loading, retry } = useResource(load);
  const { user } = useAuth();
  const navigate = useNavigate();
  const [busy, setBusy] = useState(false);
  const submitting = useRef(false);
  const [submitError, setSubmitError] = useState<unknown>(null);
  const [needsReview, setNeedsReview] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (submitting.current || !data || needsReview) return;
    submitting.current = true;
    setBusy(true);
    setSubmitError(null);
    const delivery = Object.fromEntries(
      new FormData(event.currentTarget),
    ) as unknown as DeliveryRequest;
    try {
      const order = await api.placeOrder(delivery);
      navigate("/order-confirmation", { replace: true, state: { order } });
    } catch (err) {
      setSubmitError(err);
      // Re-check the cart after any failed submission before allowing another attempt.
      setNeedsReview(true);
    } finally {
      submitting.current = false;
      setBusy(false);
    }
  }
  if (loading) return <Loading label="Checking your order totals…" />;
  if (error || !data)
    return (
      <section className="section">
        <h1>Checkout</h1>
        <ErrorMessage error={error} retry={retry} />
        <Link className="button secondary" to="/cart">
          Return to basket
        </Link>
      </section>
    );
  return (
    <section className="section">
      <Link className="back-link" to="/cart">
        ← Back to basket
      </Link>
      <span className="eyebrow">THE LAST LITTLE DETAILS</span>
      <h1>Make yourself at home.</h1>
      <p className="page-intro">
        Review your order and tell us where it’s going.
      </p>
      <div className="two-column">
        <div className="panel">
          <h2>Delivery details</h2>
          <ErrorMessage error={submitError} />
          <form className="form" onSubmit={submit}>
            <label>
              Full name
              <input
                name="delivery_name"
                autoComplete="name"
                defaultValue={`${user?.first_name || ""} ${user?.last_name || ""}`.trim()}
                required
                maxLength={100}
              />
            </label>
            <label>
              Phone number
              <input
                name="delivery_phone"
                type="tel"
                autoComplete="tel"
                defaultValue={user?.phone}
                required
                maxLength={30}
              />
            </label>
            <label>
              Street address
              <input
                name="delivery_address"
                autoComplete="street-address"
                required
                maxLength={200}
              />
            </label>
            <div className="form-row">
              <label>
                Town / city
                <input
                  name="delivery_city"
                  autoComplete="address-level2"
                  required
                  maxLength={100}
                />
              </label>
              <label>
                County / region
                <input
                  name="delivery_state"
                  autoComplete="address-level1"
                  required
                  maxLength={100}
                />
              </label>
            </div>
            <label>
              Delivery instructions <span className="muted">(optional)</span>
              <textarea
                name="delivery_instructions"
                rows={3}
                maxLength={200}
                placeholder="Anything that will help with your delivery?"
              />
            </label>
            <p className="subtle-note">
              This demo places an order without collecting payment.
            </p>
            <button disabled={busy || needsReview}>
              {busy
                ? "Placing your order…"
                : `Place order · ${money(data.total)}`}
            </button>
            {needsReview && (
              <div className="notice">
                <p>
                  Review the latest basket and order history before trying
                  again, especially if the connection was interrupted.
                </p>
                <Link to="/cart">Review basket</Link> ·{" "}
                <Link to="/orders">Check orders</Link>
                <button
                  className="text-button"
                  type="button"
                  onClick={() => {
                    setNeedsReview(false);
                    retry();
                  }}
                >
                  Refresh checkout
                </button>
              </div>
            )}
          </form>
        </div>
        <aside className="panel order-summary">
          <span className="eyebrow">YOUR ORDER FROM</span>
          <h2>{data.restaurant.name}</h2>
          <div className="summary-lines">
            {data.items.map((item) => (
              <div key={item.menu_item_id}>
                <span>
                  <b>{item.quantity}×</b> {item.name}
                </span>
                <strong>{money(item.line_total)}</strong>
              </div>
            ))}
          </div>
          <Totals {...data} />
          <p className="subtle-note">
            Totals confirmed by FoodHub. Prices are in GBP.
          </p>
        </aside>
      </div>
    </section>
  );
}
