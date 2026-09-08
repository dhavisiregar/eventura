export const UPLOADS_BASE_URL = (process.env.NEXT_PUBLIC_UPLOADS_URL || "http://localhost:8080").replace(/\/$/, "");

/** Resolves a banner/avatar URL returned by the API (which may be a relative /uploads/... path) into an absolute URL. */
export function resolveAssetUrl(url: string | null | undefined): string | null {
  if (!url) return null;
  if (/^https?:\/\//i.test(url)) return url;
  return `${UPLOADS_BASE_URL}${url.startsWith("/") ? "" : "/"}${url}`;
}
