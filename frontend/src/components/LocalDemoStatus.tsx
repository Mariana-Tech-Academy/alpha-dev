import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import { api } from "../api/foodhub";
import { session } from "../api/client";
import type { OrderHistory } from "../api/types";
import { useAuth } from "../auth/context";
import { demoNextStatus, localDemoAvailable } from "../utils/demo";
import { createDemoManagerSession } from "../utils/demoManagerSession";
import { statusLabel } from "../utils/format";
import { ErrorMessage } from "./ui";

type Props = { orders: OrderHistory["orders"]; onUpdated: () => void };

export function LocalDemoStatus(props: Props) {
  const { user } = useAuth();
  // A new customer identity remounts the panel, discarding credentials and staff JWT.
  const [binding, setBinding] = useState({ user, version: 0 });
  if (binding.user !== user) setBinding({ user, version: binding.version + 1 });
  return localDemoAvailable() ? (
    <ManagerPanel key={binding.version} {...props} />
  ) : null;
}

function ManagerPanel({ orders, onUpdated }: Props) {
  const [manager] = useState(() =>
    createDemoManagerSession(
      {
        login: api.demoManagerLogin,
        me: api.demoManagerMe,
        restaurant: (id, signal) => api.restaurant(String(id), signal),
        advance: api.advanceDemoOrder,
      },
      { enabled: localDemoAvailable, customerSession: session.get },
    ),
  );
  const [selected, setSelected] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const [message, setMessage] = useState("");
  const form = useRef<HTMLFormElement>(null);
  const mounted = useRef(false);
  const operation = useRef(0);
  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      manager.clear();
    };
  }, [manager]);
  const eligible = orders.filter((item) => demoNextStatus(item.status));
  const order =
    eligible.find((item) => String(item.id) === selected) || eligible[0];
  const identity = manager.getIdentity();
  const authorised = !!order && identity?.restaurantId === order.restaurant_id;
  const next = order && demoNextStatus(order.status);
  function logout() {
    operation.current++;
    manager.clear();
    form.current?.reset();
    setBusy(false);
    setError(null);
    setMessage("");
  }
  async function login(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!order || busy) return;
    const values = new FormData(event.currentTarget);
    const email = String(values.get("demo-manager-email") || "");
    const password = String(values.get("demo-manager-password") || "");
    // Uncontrolled inputs keep credentials out of React state; clear immediately.
    event.currentTarget.reset();
    const current = ++operation.current;
    setBusy(true);
    setError(null);
    setMessage("");
    try {
      const authenticated = await manager.login(
        email,
        password,
        order.restaurant_id,
      );
      if (mounted.current && current === operation.current && authenticated)
        setMessage("Manager authenticated");
    } catch (err) {
      if (mounted.current && current === operation.current) setError(err);
    } finally {
      if (mounted.current && current === operation.current) {
        form.current?.reset();
        setBusy(false);
      }
    }
  }
  async function advance() {
    if (!order || !authorised || busy) return;
    const current = ++operation.current;
    setBusy(true);
    setError(null);
    setMessage("");
    try {
      const result = await manager.advance(order, () => {
        if (mounted.current && current === operation.current) onUpdated();
      });
      if (mounted.current && current === operation.current && result) {
        setMessage(
          `${result.order_number}: ${statusLabel(result.status)} saved.`,
        );
      }
    } catch (err) {
      if (mounted.current && current === operation.current) setError(err);
    } finally {
      if (mounted.current && current === operation.current) setBusy(false);
    }
  }
  return (
    <aside className="panel local-demo-status">
      <h2>LOCAL-DEMO-ONLY — Advance Status</h2>
      <p>
        Separate manager login. Your customer session stays signed in. Managers
        can advance only their restaurant’s orders.
      </p>
      {identity && (
        <div>
          <p role="status">Manager authenticated</p>
          <button className="secondary" onClick={logout}>
            Manager logout
          </button>
        </div>
      )}
      {order ? (
        <>
          <label>
            Order
            <select
              value={String(order.id)}
              disabled={busy}
              onChange={(event) => {
                const chosen = eligible.find(
                  (item) => String(item.id) === event.target.value,
                );
                if (chosen?.restaurant_id !== order.restaurant_id) logout();
                setSelected(event.target.value);
              }}
            >
              {eligible.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.order_number} · {statusLabel(item.status)}
                </option>
              ))}
            </select>
          </label>
          {!authorised && (
            <form
              ref={form}
              className="form"
              autoComplete="off"
              onSubmit={login}
            >
              <label>
                Manager email
                <input
                  type="email"
                  name="demo-manager-email"
                  autoComplete="off"
                  spellCheck={false}
                  autoCapitalize="none"
                  required
                  disabled={busy}
                />
              </label>
              <label>
                Manager password (hidden)
                <input
                  type="password"
                  name="demo-manager-password"
                  autoComplete="off"
                  required
                  disabled={busy}
                />
              </label>
              <button type="submit" disabled={busy}>
                {busy ? "Checking manager…" : "Sign in as manager"}
              </button>
              {busy && (
                <button type="button" className="secondary" onClick={logout}>
                  Cancel manager login
                </button>
              )}
            </form>
          )}
          <button disabled={busy || !authorised} onClick={advance}>
            {busy && authorised
              ? "Saving…"
              : `Advance Status → ${statusLabel(next!)}`}
          </button>
        </>
      ) : (
        <p>No orders on this page can be advanced.</p>
      )}
      <p className="muted">
        Manager credentials and session are never saved in browser storage.
        Manager logout clears this local session; it does not log out the
        customer.
      </p>
      <ErrorMessage error={error} />
      {message && message !== "Manager authenticated" && (
        <p role="status">{message}</p>
      )}
    </aside>
  );
}
