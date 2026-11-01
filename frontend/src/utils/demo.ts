import type { OrderStatus } from "../api/types";

const loopback = (host: string) =>
  ["localhost", "127.0.0.1", "::1", "[::1]"].includes(host);
export function isLocalDemoEnabled(
  dev: boolean,
  flag: string | undefined,
  hostname: string,
  apiBase: string,
) {
  try {
    return (
      dev &&
      flag === "1" &&
      loopback(hostname) &&
      loopback(new URL(apiBase).hostname)
    );
  } catch {
    return false;
  }
}
export function localDemoAvailable() {
  return isLocalDemoEnabled(
    import.meta.env?.DEV === true,
    import.meta.env?.VITE_FOODHUB_DEMO_STATUS,
    window.location.hostname,
    import.meta.env?.VITE_API_BASE_URL || "http://localhost:8080/api",
  );
}
export function demoNextStatus(status: OrderStatus): OrderStatus | undefined {
  return (
    {
      pending: "preparing",
      confirmed: "preparing",
      preparing: "out_for_delivery",
      ready_for_delivery: "out_for_delivery",
      out_for_delivery: "delivered",
    } as Partial<Record<OrderStatus, OrderStatus>>
  )[status];
}
