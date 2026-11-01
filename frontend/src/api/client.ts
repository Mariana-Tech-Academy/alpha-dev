const baseUrl = (
  import.meta.env?.VITE_API_BASE_URL || "http://localhost:8080/api"
).replace(/\/$/, "");
const tokenKey = "foodhub.token";
export const session = {
  get: () => sessionStorage.getItem(tokenKey),
  set: (token: string) => sessionStorage.setItem(tokenKey, token),
  clear: () => sessionStorage.removeItem(tokenKey),
};
export class ApiError extends Error {
  status: number;
  fields: Record<string, string>;
  constructor(
    message: string,
    status: number,
    fields: Record<string, string> = {},
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.fields = fields;
  }
}
export async function request<T>(
  path: string,
  options: RequestInit = {},
  authenticated = false,
): Promise<T> {
  const headers = new Headers(options.headers);
  if (options.body) headers.set("Content-Type", "application/json");
  const token = authenticated ? session.get() : null;
  if (token) headers.set("Authorization", `Bearer ${token}`);
  let response: Response;
  try {
    response = await fetch(`${baseUrl}${path}`, { ...options, headers });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError")
      throw error;
    throw new ApiError(
      "Cannot reach FoodHub. Check that the API is running and try again.",
      0,
    );
  }
  if (response.status === 401 && authenticated) {
    // A response from an older session must not clear a newer login.
    if (session.get() === token) {
      session.clear();
      window.dispatchEvent(new Event("foodhub:expired"));
    }
    throw new ApiError("Your session has expired. Please log in again.", 401);
  }
  if (response.status === 204 && response.ok) return undefined as T;
  const body = await response.json().catch(() => null);
  if (!response.ok || body?.success === false) {
    throw new ApiError(
      body?.message || `Request failed (${response.status}). Please try again.`,
      response.status,
      body?.errors,
    );
  }
  if (!body || !("success" in body))
    throw new ApiError(
      "The API returned an unexpected response.",
      response.status,
    );
  return body.data as T;
}
