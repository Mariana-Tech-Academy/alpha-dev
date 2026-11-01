import type { OrderHistory } from "../api/types";
import { money } from "../utils/format";
import { Totals } from "./ui";

export function OrderDetails({
  order,
}: {
  order: OrderHistory["orders"][number];
}) {
  const delivery = [
    order.delivery_name,
    order.delivery_phone,
    order.delivery_address,
    order.delivery_city,
    order.delivery_state,
  ].filter(Boolean);
  return (
    <details className="history-details">
      <summary>View purchased items and delivery details</summary>
      <div className="history-detail-body">
        <h4>Purchased items</h4>
        {order.items?.length ? (
          <ul className="purchased-items">
            {order.items.map((item, index) => (
              <li key={index}>
                <span>
                  {item.quantity} × {item.name}
                  <small>{money(item.unit_price)} each</small>
                </span>
                <strong>{money(item.line_total)}</strong>
              </li>
            ))}
          </ul>
        ) : (
          <p className="muted">Item details were not saved for this order.</p>
        )}
        <Totals
          subtotal={order.subtotal}
          delivery_fee={order.delivery_fee}
          total={order.total}
        />
        <h4>Saved delivery details</h4>
        {delivery.length ? (
          <address>
            {delivery.map((line, index) => (
              <div key={index}>{line}</div>
            ))}
          </address>
        ) : (
          <p className="muted">
            Delivery details were not saved for this order.
          </p>
        )}
        {order.delivery_instructions && (
          <p>
            <strong>Delivery instructions:</strong>{" "}
            {order.delivery_instructions}
          </p>
        )}
      </div>
    </details>
  );
}
