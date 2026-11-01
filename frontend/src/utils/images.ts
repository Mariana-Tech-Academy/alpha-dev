import photoMap from "./demo-photo-map.json" with { type: "json" };
/** Local public assets use Vite's base; API-provided HTTP(S) URLs stay intact. */
export function imageUrl(
  value?: string,
  base = import.meta.env?.BASE_URL || "/",
) {
  const path = value?.trim();
  if (!path) return undefined;
  if (path.startsWith("/demo-images/") && !path.includes("..")) {
    return `${base.replace(/\/$/, "")}${path}`;
  }
  if (/^https?:\/\//i.test(path)) {
    try {
      const url = new URL(path);
      return url.hostname ? path : undefined;
    } catch {
      return undefined;
    }
  }
  if (path.startsWith("/") && !path.startsWith("//")) return path;
  return undefined;
}

/** Approved photos only; legacy local art is suppressed on customer cards. */
export function imageCandidates(
  value?: string,
  base = import.meta.env?.BASE_URL || "/",
) {
  const path = value?.trim();
  if (!path) return [];
  const promoted = (photoMap as Record<string, string>)[path];
  if (promoted) return [imageUrl(promoted, base)!];
  if (
    path.startsWith("/demo-images/") &&
    !path.startsWith("/demo-images/photos/")
  )
    return [];
  const resolved = imageUrl(path, base);
  return resolved ? [resolved] : [];
}

/** Visibility depends on the supplied image path, never network load success. */
export function hasMenuPhoto(value?: string) {
  return imageCandidates(value).length > 0;
}

/** Filter from configured image metadata, never an image's load/error state. */
export function hasRestaurantPhoto(value?: string) {
  if (!hasMenuPhoto(value)) return false;
  const path = value!.trim().split(/[?#]/)[0];
  return !/(?:^|[/_.-])(?:placeholder|fallback|illustration)(?:[/_.-]|$)|\.svg$/i.test(
    path,
  );
}
