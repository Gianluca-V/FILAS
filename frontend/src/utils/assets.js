// Resolves a DB-stored image path (e.g. "assets/xyz.jpg", sometimes with
// backslashes — a Windows-authored-admin-panel quirk) into a path the SPA
// can serve from frontend/public/assets/. Nullable/empty values fall back
// to a default placeholder, matching the legacy behavior in
// js/wildeArtesanal.mjs (`product.Image != '' ? product.Image :
// "assets/default-img.png"`).
const DEFAULT_IMAGE = '/assets/default-img.png';

export function resolveAssetPath(path, fallback = DEFAULT_IMAGE) {
  if (!path) {
    return fallback;
  }
  const normalized = path.replace(/\\/g, '/');
  return normalized.startsWith('/') ? normalized : `/${normalized}`;
}
