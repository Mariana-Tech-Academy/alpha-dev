export const money = (minorUnits: number) =>
  new Intl.NumberFormat("en-GB", { style: "currency", currency: "GBP" }).format(
    minorUnits / 100,
  );
export const date = (value: string) =>
  new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
export const statusLabel = (value: string) => value.replaceAll("_", " ");
