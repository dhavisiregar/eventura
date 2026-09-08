// Lightweight browser cookie helpers. The auth token is kept in a regular
// (non-httpOnly) cookie so both the Next.js middleware (route protection)
// and client-side API calls (Authorization header) can read it.
const TOKEN_COOKIE = "token";

export function setTokenCookie(token: string, maxAgeSeconds = 60 * 60 * 24 * 3) {
  if (typeof document === "undefined") return;
  document.cookie = `${TOKEN_COOKIE}=${token}; path=/; max-age=${maxAgeSeconds}; SameSite=Lax`;
}

export function getTokenCookie(): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(new RegExp(`(?:^|; )${TOKEN_COOKIE}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : null;
}

export function clearTokenCookie() {
  if (typeof document === "undefined") return;
  document.cookie = `${TOKEN_COOKIE}=; path=/; max-age=0`;
}
