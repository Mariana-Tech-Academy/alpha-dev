import { useState } from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/context";
import { ErrorMessage } from "./ui";
export function Layout() {
  const { user, loading, expired, error, retry, logout } = useAuth();
  const [logoutError, setLogoutError] = useState<unknown>(null);
  const [busy, setBusy] = useState(false);
  const navigate = useNavigate();
  async function signOut() {
    setBusy(true);
    setLogoutError(null);
    try {
      await logout();
      navigate("/");
    } catch (err) {
      setLogoutError(err);
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <header className="site-header">
        <div className="header-inner">
          <Link className="brand" to="/" aria-label="FoodHub home">
            <span className="brand-mark" aria-hidden="true">
              fh
            </span>
            Food<span>Hub</span>
            <i />
          </Link>
          <nav aria-label="Main navigation">
            <NavLink to="/" end>
              Explore
            </NavLink>
            <NavLink to="/orders">My orders</NavLink>
            <NavLink className="basket-link" to="/cart">
              <span aria-hidden="true">♧</span> Basket
            </NavLink>
          </nav>
          <div className="account-nav">
            {loading ? (
              <span className="muted">Checking session…</span>
            ) : user ? (
              <>
                <span className="greeting">Hi, {user.first_name}</span>
                <button
                  className="secondary small"
                  onClick={signOut}
                  disabled={busy}
                >
                  {busy ? "Logging out…" : "Log out"}
                </button>
              </>
            ) : (
              <>
                <Link to="/login" className="login-link">
                  Log in
                </Link>
                <Link className="button small" to="/register">
                  Sign up
                </Link>
              </>
            )}
          </div>
        </div>
      </header>
      <main id="main" className="container">
        {expired && !user && (
          <p className="notice" role="status">
            Your session has expired. <Link to="/login">Log in again</Link> to
            continue your order.
          </p>
        )}
        <ErrorMessage error={error} retry={retry} />
        <ErrorMessage error={logoutError} />
        <Outlet />
      </main>
      <footer className="site-footer">
        <div>
          <Link className="brand" to="/">
            Food<span>Hub</span>
            <i />
          </Link>
          <p>A good meal brings us together.</p>
        </div>
        <span>
          FoodHub UK demo · All prices in GBP
          <br />
          <a
            href={`${import.meta.env.BASE_URL}demo-images/photos/gallery.html`}
          >
            Photo credits
          </a>
        </span>
      </footer>
    </>
  );
}
