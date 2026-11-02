export const money = (minorUnits: number) =>
  new Intl.NumberFormat("en-GB", { style: "currency", currency: "GBP" }).format(
    minorUnits / 100,
  );
export const date = (value: string) =>
  new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
export function statusLabel(value: string) {
  const labels: Record<string, string> = {
    pending: "Pending",
    confirmed: "Pending",
    preparing: "In Progress",
    ready_for_delivery: "In Progress",
    out_for_delivery: "Out for Delivery",
    delivered: "Complete",
    cancelled: "Cancelled",
  };
  return labels[value] || value.replaceAll("_", " ");
}
