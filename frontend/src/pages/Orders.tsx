import { useCallback } from "react";
import { Link, useLocation, useSearchParams } from "react-router-dom";
import { api } from "../api/foodhub";
import type { PlacedOrder } from "../api/types";
import { useResource } from "../hooks/useResource";
import { Empty, ErrorMessage, Loading, Status, Totals } from "../components/ui";
import { date, money } from "../utils/format";
export function Orders() {
  const [params, setParams] = useSearchParams();
  const page = Math.max(1, Number(params.get("page")) || 1);
  const load = useCallback(
    (signal: AbortSignal) => api.orders(page, signal),
    [page],
  );
  const { data, error, loading, retry } = useResource(load);
  return (
    <section className="section">
      <div className="section-heading">
        <div>
          <span className="eyebrow">GOOD MEALS, GOOD MEMORIES</span>
          <h1>Your orders</h1>
        </div>
        <button className="secondary" onClick={retry} disabled={loading}>
          Refresh
        </button>
      </div>
      {loading ? (
        <Loading label="Finding your orders…" />
      ) : error ? (
        <ErrorMessage error={error} retry={retry} />
      ) : !data?.orders.length ? (
        <Empty title="No orders here yet">
          <p>Your next favourite meal is waiting to be discovered.</p>
          <Link className="button" to="/">
            Explore restaurants →
          </Link>
        </Empty>
      ) : (
        <div className="order-list">
          {data.orders.map((order) => (
            <article className="panel history-card" key={order.id}>
              <div>
                <span className="eyebrow">{order.order_number}</span>
                <h3>{order.restaurant_name || "Restaurant"}</h3>
                <p className="muted">{date(order.created_at)}</p>
              </div>
              <Status value={order.status} />
              <strong>{money(order.total)}</strong>
              <Link
                className="text-button"
                to={`/restaurants/${order.restaurant_id}`}
              >
                View restaurant →
              </Link>
            </article>
          ))}
        </div>
      )}
      {data && (data.total > data.limit || page > 1) && (
        <nav className="pagination" aria-label="Order pages">
          <button
            className="secondary"
            disabled={loading || page <= 1}
            onClick={() => setParams({ page: String(page - 1) })}
          >
            ← Previous
          </button>
          <span>
            Page {data.page} · {data.total} orders
          </span>
          <button
            className="secondary"
            disabled={loading || data.page * data.limit >= data.total}
            onClick={() => setParams({ page: String(page + 1) })}
          >
            Next →
          </button>
        </nav>
      )}
    </section>
  );
}
export function Confirmation() {
  const location = useLocation();
  const order = location.state?.order as PlacedOrder | undefined;
  if (!order)
    return (
      <Empty title="Looking for your order?">
        <p>You can find your placed orders in your order history.</p>
        <Link className="button" to="/orders">
          View orders →
        </Link>
      </Empty>
    );
  return (
    <section className="confirmation">
      <div className="confirmation-check" aria-hidden="true">
        ✓
      </div>
      <span className="eyebrow">THANK YOU FOR ORDERING</span>
      <h1>Good food is on the way.</h1>
      <p>Your order has been placed with the restaurant.</p>
      <div className="panel confirmation-receipt">
        <div className="section-heading">
          <strong>{order.order_number}</strong>
          <Status value={order.status} />
        </div>
        <div className="summary-lines">
          {order.items.map((item, index) => (
            <div key={index}>
              <span>
                {item.quantity}× {item.name}
              </span>
              <strong>{money(item.line_total)}</strong>
            </div>
          ))}
        </div>
        <Totals {...order} />
      </div>
      <div className="actions">
        <Link className="button" to="/orders">
          View your orders →
        </Link>
        <Link className="button secondary" to="/">
          Keep exploring
        </Link>
      </div>
    </section>
  );
}
