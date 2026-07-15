import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../../api/resources.js', () => ({
  getProducts: vi.fn(),
  createOrder: vi.fn(),
}));

import { createOrder, getProducts } from '../../../api/resources.js';
import WildeArtesanalView from '../WildeArtesanalView.vue';

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

// Regression lock for the ID coercion in WildeArtesanalView's addToCart():
// GET /api/products returns Product.ID as a JSON STRING (see backend
// dto.ProductResponse), but POST /api/orders binds productID to an int and
// silently defaults a type-mismatched field to 0 (the handler ignores the
// ShouldBindJSON error). `Number(product.ID)` at the addToCart call site is
// the only thing keeping checkout correct end-to-end — this test mounts the
// real view + cart store + CartPanel and asserts the payload actually
// reaching createOrder carries a numeric productID, so dropping the
// coercion fails the suite instead of only failing silently in production.
describe('WildeArtesanalView checkout', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it('coerces the string product ID to a number before calling createOrder', async () => {
    getProducts.mockResolvedValue([
      { ID: '5', Name: 'Alfajores', Price: '600', Stock: '3', Image: 'assets/x.jpg' },
    ]);
    createOrder.mockResolvedValue({ message: 'Order created successfully' });

    const wrapper = mount(WildeArtesanalView);
    await flushPromises();

    await wrapper.find('.product__buy').trigger('click');

    await wrapper.find('#name').setValue('Ana Pérez');
    await wrapper.find('#phone').setValue('1122334455');
    await wrapper.find('form.cart').trigger('submit');
    await flushPromises();

    expect(createOrder).toHaveBeenCalledTimes(1);
    const payload = createOrder.mock.calls[0][0];
    expect(payload.orderProducts).toEqual([{ productID: 5, quantity: 1 }]);
    expect(typeof payload.orderProducts[0].productID).toBe('number');
  });
});
