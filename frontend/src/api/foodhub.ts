import { request } from "./client";
import type {
  Cart,
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
  summary: (signal?: AbortSignal) =>
    request<CheckoutSummary>("/checkout/summary", { signal }, true),
  placeOrder: (data: DeliveryRequest) =>
    request<PlacedOrder>("/orders", { method: "POST", body: json(data) }, true),
  orders: (page: number, signal?: AbortSignal) =>
    request<OrderHistory>(`/orders?page=${page}&limit=10`, { signal }, true),
};
