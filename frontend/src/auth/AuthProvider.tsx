import { useCallback, useEffect, useState } from "react";
import type { ReactNode } from "react";
import { api } from "../api/foodhub";
import { ApiError, session } from "../api/client";
import type { LoginResult, User } from "../api/types";
import { AuthContext } from "./context";
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(() => !!session.get());
  const [error, setError] = useState("");
  const [expired, setExpired] = useState(false);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    let active = true;
    const token = session.get();
    if (!token) return;
    api
      .me()
      .then((profile) => {
        if (active && session.get() === token) setUser(profile);
      })
      .catch((err) => {
        if (active && !(err instanceof ApiError && err.status === 401))
          setError(err.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [attempt]);
  useEffect(() => {
    const onExpiry = () => {
      setUser(null);
      setExpired(true);
      setError("");
    };
    window.addEventListener("foodhub:expired", onExpiry);
    return () => window.removeEventListener("foodhub:expired", onExpiry);
  }, []);
  const acceptLogin = useCallback((result: LoginResult) => {
    session.set(result.token);
    setUser(result.user);
    setExpired(false);
    setError("");
    setLoading(false);
  }, []);
  const logout = async () => {
    await api.logout();
    session.clear();
    setUser(null);
    setExpired(false);
    setError("");
  };
  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        error,
        expired,
        acceptLogin,
        logout,
        retry: () => {
          setLoading(true);
          setError("");
          setAttempt((n) => n + 1);
        },
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
