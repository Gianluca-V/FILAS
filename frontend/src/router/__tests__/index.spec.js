import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useAuthStore } from '../../stores/auth.js';
import router from '../index.js';

describe('router auth guard', () => {
  beforeEach(async () => {
    setActivePinia(createPinia());
    // /admin/products doesn't exist yet (admin views ship in task 8.1); the
    // guard runs before route matching, so vue-router logs a harmless "no
    // match" warning for it below. Silenced to keep test output focused on
    // the guard behavior under test.
    vi.spyOn(console, 'warn').mockImplementation(() => {});
    await router.push('/');
    await router.isReady();
  });

  it('allows public routes without authentication', async () => {
    await router.push('/noticias');

    expect(router.currentRoute.value.name).toBe('news-list');
  });

  it('allows /admin/login without authentication', async () => {
    await router.push('/admin/login');

    expect(router.currentRoute.value.name).toBe('admin-login');
  });

  it('redirects unauthenticated access to a protected /admin/* route to /admin/login', async () => {
    await router.push('/admin/products');

    expect(router.currentRoute.value.name).toBe('admin-login');
  });

  it('allows an authenticated session to reach a protected /admin/* route', async () => {
    const auth = useAuthStore();
    auth.token = 'valid-token';

    await router.push('/admin/products');

    expect(router.currentRoute.value.name).not.toBe('admin-login');
  });
});
