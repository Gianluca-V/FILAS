// Fetch wrapper shared by every resource call in api/resources.js. Replaces
// the legacy js/API.mjs class hierarchy (design §3.1) with a single
// interceptor: it injects the bearer token from the Pinia auth store, and
// centrally handles session expiry (any 401 logs the user out and sends
// them back to the admin login screen), so individual resource calls never
// have to think about auth plumbing.
import { useAuthStore } from '../stores/auth.js';

const BASE_URL = import.meta.env.VITE_API_URL || '/api';

export class ApiError extends Error {
  constructor(message, status, body) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

async function parseJsonSafely(response) {
  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    return null;
  }
  try {
    return await response.json();
  } catch {
    return null;
  }
}

async function request(path, { method = 'GET', body, headers = {} } = {}) {
  const auth = useAuthStore();
  const requestHeaders = { 'Content-Type': 'application/json', ...headers };

  if (auth.token) {
    requestHeaders.Authorization = `Bearer ${auth.token}`;
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    method,
    headers: requestHeaders,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  // Any 401 means the current session is no longer valid (missing, expired,
  // or malformed token) — log out and send the user back to the admin
  // login screen. Note: this also fires for a failed login attempt itself
  // (bad credentials -> 401), which is a harmless no-op there since no
  // session exists yet and the user is already on /admin/login.
  if (response.status === 401) {
    auth.logout();
    if (typeof window !== 'undefined' && window.location) {
      window.location.href = '/admin/login';
    }
    throw new ApiError('Unauthorized', 401, await parseJsonSafely(response));
  }

  const data = await parseJsonSafely(response);

  if (!response.ok) {
    throw new ApiError(data?.message || `Request failed with status ${response.status}`, response.status, data);
  }

  return data;
}

export const get = (path, options) => request(path, { ...options, method: 'GET' });
export const post = (path, body, options) => request(path, { ...options, method: 'POST', body });
export const put = (path, body, options) => request(path, { ...options, method: 'PUT', body });
export const patch = (path, body, options) => request(path, { ...options, method: 'PATCH', body });
export const del = (path, options) => request(path, { ...options, method: 'DELETE' });
