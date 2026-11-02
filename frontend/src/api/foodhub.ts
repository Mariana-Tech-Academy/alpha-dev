import { localDemoAvailable } from "../utils/demo.ts";
import { request } from "./client.ts";
import type {
  Cart,
  OrderStatus,
  CheckoutSummary,
  DeliveryRequest,
  LoginResult,
  OrderHistory,
  PlacedOrder,
  RegisterRequest,
  Restaurant,
  RestaurantDetails,
  User,
} from "./types";
const json = (value: unknown) => JSON.stringify(value);
export const api = {
  demoManagerLogin: (email: string, password: string, signal?: AbortSignal) => {
    if (!localDemoAvailable())
      throw new Error("Local demo controls are disabled.");
    return request<LoginResult>("/auth/login", {
      method: "POST",
      body: json({ email, password }),
      signal,
    });
  },
  demoManagerMe: (staffToken: string, signal?: AbortSignal) => {
    if (!localDemoAvailable())
      throw new Error("Local demo controls are disabled.");
    return request<User>("/auth/me", {
      headers: { Authorization: `Bearer ${staffToken}` },
      signal,
    });
  },
  advanceDemoOrder: (
    id: number,
    status: OrderStatus,
    staffToken: string,
    signal?: AbortSignal,
  ) => {
    if (!localDemoAvailable())
      throw new Error("Local demo controls are disabled.");
    return request<{
      id: number;
      order_number: string;
      status: OrderStatus;
      updated_at: string;
    }>(`/demo/orders/${id}/advance`, {
      method: "POST",
      headers: { Authorization: `Bearer ${staffToken}` },
      body: json({ expected_status: status }),
      signal,
    });
  },
  register: (data: RegisterRequest) =>
    request<User>("/auth/register", { method: "POST", body: json(data) }),
  login: (email: string, password: string) =>
    request<LoginResult>("/auth/login", {
      method: "POST",
      body: json({ email, password }),
    }),
  me: () => request<User>("/auth/me", {}, true),
  logout: () => request<void>("/auth/logout", { method: "POST" }, true),
  restaurants: (query: string, page: number, signal?: AbortSignal) =>
    request<Restaurant[]>(
      `${query ? "/restaurants/search" : "/restaurants"}?${new URLSearchParams({ ...(query ? { q: query } : {}), page: String(page), limit: "12" })}`,
      { signal },
    ),
  restaurant: (id: string, signal?: AbortSignal) =>
    request<RestaurantDetails>(`/restaurants/${encodeURIComponent(id)}`, {
      signal,
    }),
  cart: (signal?: AbortSignal) => request<Cart>("/cart", { signal }, true),
  addItem: (menu_item_id: number) =>
    request<{
      id: number;
      cart_id: number;
      menu_item_id: number;
      quantity: number;
      price: number;
    }>(
      "/cart/items",
      { method: "POST", body: json({ menu_item_id, quantity: 1 }) },
      true,
    ),
  updateQuantity: (id: number, quantity: number) =>
    request<Cart>(
      `/cart/items/${id}`,
      { method: "PUT", body: json({ quantity }) },
      true,
    ),
  removeItem: (id: number) =>
    request<Cart>(`/cart/items/${id}`, { method: "DELETE" }, true),
  clearCart: () => request<Cart>("/cart", { method: "DELETE" }, true),
  summary: (signal?: AbortSignal) =>
    request<CheckoutSummary>("/checkout/summary", { signal }, true),
  placeOrder: (data: DeliveryRequest) =>
    request<PlacedOrder>("/orders", { method: "POST", body: json(data) }, true),
  orders: (page: number, signal?: AbortSignal) =>
    request<OrderHistory>(`/orders?page=${page}&limit=10`, { signal }, true),
};
