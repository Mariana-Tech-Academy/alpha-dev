import { useState } from "react";
import type { ReactNode } from "react";
import { ApiError } from "../api/client";
import { imageCandidates } from "../utils/images";
import { money, statusLabel } from "../utils/format";
export function Loading({
  label = "Loading your next good meal…",
}: {
  label?: string;
}) {
  return (
    <div className="state" role="status">
      <span className="spinner" />
      {label}
    </div>
  );
}
export function ErrorMessage({
  error,
  retry,
}: {
  error: unknown;
  retry?: () => void;
}) {
  if (!error) return null;
  return (
    <div className="error" role="alert">
      <strong>{error instanceof Error ? error.message : String(error)}</strong>
      {error instanceof ApiError && Object.keys(error.fields).length > 0 && (
        <ul>
          {Object.entries(error.fields).map(([key, message]) => (
            <li key={key}>{message}</li>
          ))}
        </ul>
      )}
      {retry && (
        <button className="text-button" onClick={retry}>
          Try again
        </button>
      )}
    </div>
  );
}
export function Empty({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="state empty">
      <span className="empty-icon" aria-hidden="true">
        ♧
      </span>
      <h2>{title}</h2>
      {children}
    </div>
  );
}
export function FoodImage({
  src,
  name,
  className = "",
}: {
  src?: string;
  name: string;
  className?: string;
}) {
  const [failedSources, setFailedSources] = useState<string[]>([]);
  const resolved = imageCandidates(src).find(
    (candidate) => !failedSources.includes(candidate),
  );
  const photo = resolved?.includes("/demo-images/photos/");
  return (
    <div className={`food-image ${className}`}>
      {resolved ? (
        <img
          src={resolved}
          alt={photo ? `${name} — representative stock photograph` : name}
          loading="lazy"
          decoding="async"
          width={480}
          height={360}
          onError={() =>
            setFailedSources((previous) => [...previous, resolved])
          }
        />
      ) : (
        <div
          className="image-placeholder"
          role="img"
          aria-label={`${name}: image unavailable`}
        >
          <small>Photo unavailable</small>
        </div>
      )}
      {resolved && photo && (
        <span className="illustration-label">Serving suggestion</span>
      )}
    </div>
  );
}
export function Status({ value }: { value: string }) {
  return <span className={`badge status-${value}`}>{statusLabel(value)}</span>;
}
export function Totals({
  subtotal,
  delivery_fee,
  total,
}: {
  subtotal: number;
  delivery_fee: number;
  total: number;
}) {
  return (
    <dl className="totals">
      <div>
        <dt>Subtotal</dt>
        <dd>{money(subtotal)}</dd>
      </div>
      <div>
        <dt>Delivery</dt>
        <dd>{money(delivery_fee)}</dd>
      </div>
      <div className="grand-total">
        <dt>Total</dt>
        <dd>{money(total)}</dd>
      </div>
    </dl>
  );
}
