import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const pushMock = vi.fn();
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router');
  return { ...actual, useRouter: () => ({ push: pushMock }) };
});

import { useAuthStore } from '../../../stores/auth.js';
import AdminLoginView from '../AdminLoginView.vue';

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('AdminLoginView', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    pushMock.mockClear();
  });

  it('calls auth.login with the entered credentials and redirects to /admin/products on success', async () => {
    const auth = useAuthStore();
    vi.spyOn(auth, 'login').mockResolvedValue('fake-token');

    const wrapper = mount(AdminLoginView);
    await wrapper.find('#username').setValue('admin');
    await wrapper.find('#password').setValue('secret');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(auth.login).toHaveBeenCalledWith('admin', 'secret');
    expect(pushMock).toHaveBeenCalledWith('/admin/products');
  });

  it('shows an error message and does not redirect when login fails', async () => {
    const auth = useAuthStore();
    vi.spyOn(auth, 'login').mockRejectedValue(new Error('Login failed'));

    const wrapper = mount(AdminLoginView);
    await wrapper.find('#username').setValue('admin');
    await wrapper.find('#password').setValue('wrong-password');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(wrapper.text()).toContain('Autentificación fallida');
    expect(pushMock).not.toHaveBeenCalled();
  });
});
