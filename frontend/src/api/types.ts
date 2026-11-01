export interface User {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  role: "customer" | "manager" | "admin";
  status: string;
  created_at: string;
}
export interface RegisterRequest {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  password: string;
  confirm_password: string;
}
export interface LoginResult {
  token: string;
  user: User;
}
export interface MenuItem {
  id: number;
  name: string;
  description: string;
  price: number;
  image_url?: string;
  category: string;
}
export interface Restaurant {
  id: number;
  manager_id: number;
  name: string;
  description: string;
  address: string;
  city?: string;
  state?: string;
  phone: string;
  email: string;
  opening_time: string;
  closing_time: string;
  image_url?: string;
  status: string;
}
export interface RestaurantDetails extends Restaurant {
  menu_items: MenuItem[];
}
export interface CartItem {
  id: number;
  menu_item_id: number;
  name: string;
  price: number;
  locked_price: number;
  quantity: number;
  line_total: number;
  is_available: boolean;
  price_changed: boolean;
  notice?: string;
}
export interface Cart {
  cart_id: number;
  restaurant_id?: number;
  items: CartItem[];
  subtotal: number;
}
export interface OrderLine {
  name: string;
  unit_price: number;
  quantity: number;
  line_total: number;
}
export interface CheckoutSummary {
  restaurant: { id: number; name: string };
  items: (OrderLine & { menu_item_id: number })[];
  subtotal: number;
  delivery_fee: number;
  total: number;
}
export interface DeliveryRequest {
  delivery_name: string;
  delivery_phone: string;
  delivery_address: string;
  delivery_city: string;
  delivery_state: string;
  delivery_instructions: string;
}
export type OrderStatus =
  | "pending"
  | "confirmed"
  | "preparing"
  | "ready_for_delivery"
  | "out_for_delivery"
  | "delivered"
  | "cancelled";
export interface PlacedOrder {
  order_number: string;
  status: OrderStatus;
  subtotal: number;
  delivery_fee: number;
  total: number;
  items: OrderLine[];
}
export interface OrderHistory {
  orders: {
    id: number;
    order_number: string;
    restaurant_id: number;
    restaurant_name: string;
    total: number;
    status: OrderStatus;
    created_at: string;
  }[];
  page: number;
  limit: number;
  total: number;
}
