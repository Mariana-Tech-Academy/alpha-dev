import { createContext, useContext } from "react";
import type { LoginResult, User } from "../api/types";
export interface AuthState {
  user: User | null;
  loading: boolean;
  error: string;
  expired: boolean;
  acceptLogin: (result: LoginResult) => void;
  logout: () => Promise<void>;
  retry: () => void;
}
export const AuthContext = createContext<AuthState | null>(null);
export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("AuthProvider is missing");
  return value;
}
