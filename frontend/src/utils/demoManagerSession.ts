import type {
  LoginResult,
  User,
  RestaurantDetails,
  OrderStatus,
} from "../api/types";

type LoginStage =
  | "login request"
  | "profile request"
  | "identity check"
  | "restaurant request"
  | "ownership check";

type Identity = { id: number; restaurantId: number };
type Order = { id: number; restaurant_id: number; status: OrderStatus };
type Updated = { order_number: string; status: OrderStatus };
export interface DemoManagerAPI {
  login(
    email: string,
    password: string,
    signal: AbortSignal,
  ): Promise<LoginResult>;
  me(token: string, signal: AbortSignal): Promise<User>;
  restaurant(id: number, signal: AbortSignal): Promise<RestaurantDetails>;
  advance(
    id: number,
    status: OrderStatus,
    token: string,
    signal: AbortSignal,
  ): Promise<Updated>;
}

/** Staff JWT lives only in this component-owned closure; never returned to the UI. */
export function createDemoManagerSession(
  api: DemoManagerAPI,
  context: {
    enabled(): boolean;
    customerSession(): string | null;
  },
) {
  let token = "";
  let identity: Identity | undefined;
  let customer: string | null = null;
  let generation = 0;
  let pending: AbortController | undefined;

  function clear() {
    generation++;
    pending?.abort();
    pending = undefined;
    token = "";
    identity = undefined;
    customer = null;
  }
  function active() {
    if (!token || !identity) return false;
    if (!context.enabled() || customer !== context.customerSession()) {
      clear();
      return false;
    }
    return !!token && !!identity;
  }
  function begin() {
    const key = context.customerSession();
    if (!context.enabled() || !key) {
      clear();
      throw new Error(
        "Local demo manager login requires an active customer session.",
      );
    }
    pending?.abort();
    pending = new AbortController();
    const operation = ++generation;
    return {
      signal: pending.signal,
      current: () => {
        if (operation !== generation || pending?.signal.aborted) return false;
        if (key !== context.customerSession() || !context.enabled()) {
          clear();
          return false;
        }
        return true;
      },
      key,
    };
  }
  function failure(error: unknown, loginStage?: LoginStage) {
    const status = (error as { status?: number })?.status;
    if (status === 401 || status === 403) clear();
    return new Error(
      loginStage
        ? `Manager authentication failed at ${loginStage}${typeof status === "number" && Number.isInteger(status) && status >= 0 && status <= 599 ? ` (HTTP ${status})` : ""}.`
        : status === 401
          ? "Manager session rejected or expired. Sign in again."
          : status === 403
            ? "Manager is not authorised for this restaurant. Sign in again."
            : status === 409
              ? "Order changed. Refresh orders before trying again."
              : "Unable to contact or update this order. Refresh and try again.",
    );
  }
  return {
    clear,
    getIdentity(): Identity | undefined {
      return active() && identity ? { ...identity } : undefined;
    },
    async login(email: string, password: string, restaurantId: number) {
      clear();
      const attempt = begin();
      let stage: LoginStage = "login request";
      try {
        const result = await api.login(email.trim(), password, attempt.signal);
        if (!attempt.current()) return undefined;
        stage = "profile request";
        const profile = await api.me(result.token, attempt.signal);
        if (!attempt.current()) return undefined;
        stage = "identity check";
        if (
          profile.role !== "manager" ||
          profile.status !== "active" ||
          profile.id !== result.user.id ||
          result.user.role !== "manager" ||
          result.user.status !== "active" ||
          !result.token ||
          /\s/.test(result.token)
        ) {
          throw new Error("Invalid manager identity");
        }
        stage = "restaurant request";
        const restaurant = await api.restaurant(restaurantId, attempt.signal);
        if (!attempt.current()) return undefined;
        stage = "ownership check";
        if (
          restaurant.id !== restaurantId ||
          restaurant.manager_id !== profile.id ||
          restaurant.status !== "active"
        ) {
          throw new Error("Wrong restaurant owner");
        }
        token = result.token;
        identity = { id: profile.id, restaurantId };
        customer = attempt.key;
        return { ...identity };
      } catch (error) {
        if (!attempt.current()) return undefined;
        clear();
        throw failure(error, stage);
      }
    },
    async advance(order: Order, onUpdated: () => void) {
      if (!active() || identity?.restaurantId !== order.restaurant_id)
        throw new Error("Sign in as this restaurant’s manager first.");
      const attempt = begin();
      try {
        // Recheck visible ownership; the backend independently enforces it on update.
        const restaurant = await api.restaurant(
          order.restaurant_id,
          attempt.signal,
        );
        if (!attempt.current()) return undefined;
        if (
          restaurant.manager_id !== identity?.id ||
          restaurant.id !== order.restaurant_id ||
          restaurant.status !== "active"
        ) {
          throw Object.assign(new Error("Manager ownership changed"), {
            status: 403,
          });
        }
        const updated = await api.advance(
          order.id,
          order.status,
          token,
          attempt.signal,
        );
        if (!attempt.current()) return undefined;
        onUpdated();
        return updated;
      } catch (error) {
        if (!attempt.current()) return undefined;
        throw failure(error);
      }
    },
  };
}
