// Shared helper for the public list views: several read endpoints
// (news/gallery/family/organizations) return 404 for an EMPTY list rather
// than 200 [] (see backend/internal/handler/rest/*.go doc comments). That
// 404 is not a failure — it's "no items yet" — so views must distinguish
// it from a real error instead of surfacing an error banner.
import { ApiError } from '../api/client.js';

export function isNotFound(error) {
  return error instanceof ApiError && error.status === 404;
}
