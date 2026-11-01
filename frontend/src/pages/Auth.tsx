import { useState } from "react";
import type { FormEvent } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { api } from "../api/foodhub";
import type { RegisterRequest } from "../api/types";
import { useAuth } from "../auth/context";
import { ErrorMessage, Loading } from "../components/ui";
export function Auth({ register = false }: { register?: boolean }) {
  const { user, loading, acceptLogin } = useAuth();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const [registered, setRegistered] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const candidate = location.state?.from;
  const from =
    typeof candidate === "string" &&
    candidate.startsWith("/") &&
    !candidate.startsWith("//")
      ? candidate
      : "/";
  if (loading) return <Loading label="Checking your session…" />;
  if (user) return <Navigate to={from} replace />;
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError(null);
    const values = Object.fromEntries(
      new FormData(event.currentTarget),
    ) as unknown as RegisterRequest;
    try {
      if (register && !registered) {
        await api.register(values);
        setRegistered(true);
      } else {
        acceptLogin(await api.login(values.email, values.password));
        navigate(from, { replace: true });
      }
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  }
  const showRegister = register && !registered;
  return (
    <section className="auth-shell">
      <aside className="auth-intro">
        <span className="eyebrow">WELCOME TO FOODHUB</span>
        <h1>
          Great food.
          <br />
          Good company.
          <br />
          <em>Your place.</em>
        </h1>
        <p>Your local favourites, all in one place.</p>
        <span className="auth-art" aria-hidden="true">
          ♨
        </span>
      </aside>
      <div className="panel auth-panel">
        <span className="eyebrow">LET’S GET YOU FED</span>
        <h2>{showRegister ? "Join the table" : "Welcome back"}</h2>
        <p className="muted">
          {showRegister
            ? "Create your FoodHub customer account."
            : "Log in to order your next favourite meal."}
        </p>
        {registered && (
          <p className="success" role="status">
            Account created. Log in to start your order.
          </p>
        )}
        <ErrorMessage error={error} />
        <form
          onSubmit={submit}
          className="form"
          key={showRegister ? "register" : "login"}
        >
          {showRegister && (
            <>
              <div className="form-row">
                <label>
                  First name
                  <input
                    name="first_name"
                    autoComplete="given-name"
                    required
                    maxLength={100}
                  />
                </label>
                <label>
                  Last name
                  <input
                    name="last_name"
                    autoComplete="family-name"
                    required
                    maxLength={100}
                  />
                </label>
              </div>
              <label>
                Phone number
                <input
                  name="phone"
                  type="tel"
                  autoComplete="tel"
                  required
                  maxLength={30}
                />
              </label>
            </>
          )}
          <label>
            Email address
            <input
              name="email"
              type="email"
              autoComplete="email"
              required
              maxLength={254}
            />
          </label>
          <label>
            Password
            <input
              name="password"
              type="password"
              autoComplete={showRegister ? "new-password" : "current-password"}
              minLength={showRegister ? 8 : undefined}
              required
            />
          </label>
          {showRegister && (
            <>
              <small className="muted">
                Use at least 8 characters (maximum 72 bytes).
              </small>
              <label>
                Confirm password
                <input
                  name="confirm_password"
                  type="password"
                  autoComplete="new-password"
                  minLength={8}
                  required
                />
              </label>
            </>
          )}
          <button disabled={busy}>
            {busy
              ? "Please wait…"
              : showRegister
                ? "Create account →"
                : "Log in →"}
          </button>
        </form>
        <p className="auth-switch">
          {showRegister ? "Already have an account?" : "New to FoodHub?"}{" "}
          <Link to={showRegister ? "/login" : "/register"} state={{ from }}>
            {showRegister ? "Log in" : "Create an account"}
          </Link>
        </p>
      </div>
    </section>
  );
}
