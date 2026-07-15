import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../api/resources.js', () => ({
  createOrder: vi.fn(),
}));

import { createOrder } from '../../api/resources.js';
import { useCartStore } from '../cart.js';

describe('cart store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it('starts empty with a zero total', () => {
    const cart = useCartStore();

    expect(cart.items).toEqual([]);
    expect(cart.totalCost).toBe(0);
  });

  it('addItem adds a new line and increments quantity on repeat adds', () => {
    const cart = useCartStore();
    const product = { productID: 1, name: 'Alfajores', price: '150' };

    cart.addItem(product);
    cart.addItem(product);

    expect(cart.items).toEqual([{ productID: 1, name: 'Alfajores', price: '150', quantity: 2 }]);
  });

  it('totalCost coerces the string Price field before summing', () => {
    const cart = useCartStore();
    cart.addItem({ productID: 1, name: 'Alfajores', price: '150' });
    cart.setQuantity(1, 3);

    expect(cart.totalCost).toBe(450);
  });

  it('removeItem drops the matching line', () => {
    const cart = useCartStore();
    cart.addItem({ productID: 1, name: 'Alfajores', price: '150' });
    cart.addItem({ productID: 2, name: 'Budín', price: '200' });

    cart.removeItem(1);

    expect(cart.items).toEqual([{ productID: 2, name: 'Budín', price: '200', quantity: 1 }]);
  });

  // RED-first: this is the real logic under Strict TDD — checkout MUST
  // reject a non-positive or non-integer quantity client-side, matching the
  // backend's tightened POST /api/orders validation (quantity<=0 -> 400),
  // and MUST NOT call createOrder() when rejecting.
  it('rejects checkout when a line item has quantity 0', async () => {
    const cart = useCartStore();
    cart.addItem({ productID: 1, name: 'Alfajores', price: '150' });
    cart.setQuantity(1, 0);

    await expect(cart.checkout({ name: 'Ana', phone: '1122334455' })).rejects.toThrow();

    expect(createOrder).not.toHaveBeenCalled();
  });

  it('rejects checkout when a line item has a negative quantity', async () => {
    const cart = useCartStore();
    cart.addItem({ productID: 1, name: 'Alfajores', price: '150' });
    cart.setQuantity(1, -2);

    await expect(cart.checkout({ name: 'Ana', phone: '1122334455' })).rejects.toThrow();

    expect(createOrder).not.toHaveBeenCalled();
  });

  it('rejects checkout when a line item has a non-integer quantity', async () => {
    const cart = useCartStore();
    cart.addItem({ productID: 1, name: 'Alfajores', price: '150' });
    cart.setQuantity(1, 1.5);

    await expect(cart.checkout({ name: 'Ana', phone: '1122334455' })).rejects.toThrow();

    expect(createOrder).not.toHaveBeenCalled();
  });

  it('rejects checkout with an empty cart', async () => {
    const cart = useCartStore();

    await expect(cart.checkout({ name: 'Ana', phone: '1122334455' })).rejects.toThrow();

    expect(createOrder).not.toHaveBeenCalled();
  });

  it('rejects checkout when name or phone is blank', async () => {
    const cart = useCartStore();
    cart.addItem({ productID: 1, name: 'Alfajores', price: '150' });

    await expect(cart.checkout({ name: '  ', phone: '1122334455' })).rejects.toThrow();
    await expect(cart.checkout({ name: 'Ana', phone: '' })).rejects.toThrow();

    expect(createOrder).not.toHaveBeenCalled();
  });

  it('checks out a valid cart: calls createOrder with the mapped payload and clears the cart', async () => {
    createOrder.mockResolvedValue({ message: 'Order created successfully' });
    const cart = useCartStore();
    cart.addItem({ productID: 1, name: 'Alfajores', price: '150' });
    cart.setQuantity(1, 3);
    cart.addItem({ productID: 2, name: 'Budín', price: '200' });

    await cart.checkout({ name: 'Ana Pérez', phone: '1122334455' });

    expect(createOrder).toHaveBeenCalledWith({
      name: 'Ana Pérez',
      phone: '1122334455',
      orderProducts: [
        { productID: 1, quantity: 3 },
        { productID: 2, quantity: 1 },
      ],
    });
    expect(cart.items).toEqual([]);
  });
});
