import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../api/resources.js', () => ({
  login: vi.fn(),
}));

import { login as loginRequest } from '../../api/resources.js';
import { useAuthStore } from '../auth.js';

const TOKEN_STORAGE_KEY = 'filas_auth_token';

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('starts unauthenticated when localStorage has no token', () => {
    const auth = useAuthStore();

    expect(auth.token).toBeNull();
    expect(auth.isAuthenticated).toBe(false);
  });

  it('restores the token from localStorage on init', () => {
    localStorage.setItem(TOKEN_STORAGE_KEY, 'persisted-token');

    const auth = useAuthStore();

    expect(auth.token).toBe('persisted-token');
    expect(auth.isAuthenticated).toBe(true);
  });

  it('login stores the token and flips isAuthenticated', async () => {
    loginRequest.mockResolvedValue({ token: 'fake-jwt-token' });
    const auth = useAuthStore();

    expect(auth.isAuthenticated).toBe(false);

    await auth.login('admin', 'secret');

    expect(loginRequest).toHaveBeenCalledWith({ username: 'admin', password: 'secret' });
    expect(auth.token).toBe('fake-jwt-token');
    expect(auth.isAuthenticated).toBe(true);
    expect(localStorage.getItem(TOKEN_STORAGE_KEY)).toBe('fake-jwt-token');
  });

  it('logout clears the token and localStorage', async () => {
    loginRequest.mockResolvedValue({ token: 'fake-jwt-token' });
    const auth = useAuthStore();
    await auth.login('admin', 'secret');

    auth.logout();

    expect(auth.token).toBeNull();
    expect(auth.isAuthenticated).toBe(false);
    expect(localStorage.getItem(TOKEN_STORAGE_KEY)).toBeNull();
  });

  it('propagates a login failure without storing a token', async () => {
    loginRequest.mockRejectedValue(new Error('Login failed'));
    const auth = useAuthStore();

    await expect(auth.login('admin', 'wrong-password')).rejects.toThrow('Login failed');

    expect(auth.token).toBeNull();
    expect(auth.isAuthenticated).toBe(false);
  });
});
