import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "./context";
import { Empty, ErrorMessage, Loading } from "../components/ui";
export function CustomerRoute() {
  const { user, loading, error, retry } = useAuth();
  const location = useLocation();
  if (loading) return <Loading label="Checking your session…" />;
  if (error) return <ErrorMessage error={error} retry={retry} />;
  if (!user)
    return (
      <Navigate
        to="/login"
        state={{ from: location.pathname + location.search }}
        replace
      />
    );
  if (user.role !== "customer")
    return (
      <Empty title="A seat for customers">
        <p>
          Please log out and use a customer account to access your basket and
          orders.
        </p>
      </Empty>
    );
  return <Outlet />;
}
