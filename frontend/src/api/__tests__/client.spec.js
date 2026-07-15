import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useAuthStore } from '../../stores/auth.js';
import { ApiError, get, post } from '../client.js';

function mockFetchOnce(status, body, headers = { 'content-type': 'application/json' }) {
  global.fetch = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    headers: { get: (key) => headers[key.toLowerCase()] ?? null },
    json: async () => body,
  });
}

describe('api client', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorage.clear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('prefixes requests with the configured base URL', async () => {
    mockFetchOnce(200, []);

    await get('/products');

    expect(global.fetch).toHaveBeenCalledWith('/api/products', expect.objectContaining({ method: 'GET' }));
  });

  it('omits the Authorization header when no token is present', async () => {
    mockFetchOnce(200, []);

    await get('/products');

    const [, options] = global.fetch.mock.calls[0];
    expect(options.headers.Authorization).toBeUndefined();
  });

  it('attaches the Authorization header when a token exists', async () => {
    const auth = useAuthStore();
    auth.token = 'my-token';
    mockFetchOnce(200, []);

    await get('/orders');

    const [, options] = global.fetch.mock.calls[0];
    expect(options.headers.Authorization).toBe('Bearer my-token');
  });

  it('sends a JSON body on POST requests', async () => {
    mockFetchOnce(201, { message: 'Order created successfully' });

    await post('/orders', { orderProducts: [{ productID: 1, quantity: 2 }], name: 'Ana', phone: '123' });

    const [, options] = global.fetch.mock.calls[0];
    expect(options.method).toBe('POST');
    expect(JSON.parse(options.body)).toEqual({
      orderProducts: [{ productID: 1, quantity: 2 }],
      name: 'Ana',
      phone: '123',
    });
  });

  it('surfaces a non-2xx response as an ApiError', async () => {
    mockFetchOnce(400, { message: 'Invalid category' });

    await expect(post('/family', { Category: 'Invalid' })).rejects.toThrow(ApiError);
  });

  it('resolves successfully on a 2xx response', async () => {
    mockFetchOnce(200, [{ ID: '1', Name: 'Producto' }]);

    const data = await get('/products');

    expect(data).toEqual([{ ID: '1', Name: 'Producto' }]);
  });

  it('calls auth.logout() on a 401 response', async () => {
    // jsdom doesn't implement real navigation; stub `location` so the
    // client's redirect assignment is observable without its "Not
    // implemented: navigation" console noise.
    vi.stubGlobal('location', { href: '' });
    const auth = useAuthStore();
    auth.token = 'expired-token';
    const logoutSpy = vi.spyOn(auth, 'logout');
    mockFetchOnce(401, { message: 'Unauthorized' });

    await expect(get('/orders')).rejects.toThrow(ApiError);

    expect(logoutSpy).toHaveBeenCalledOnce();
    expect(location.href).toBe('/admin/login');
    vi.unstubAllGlobals();
  });
});
