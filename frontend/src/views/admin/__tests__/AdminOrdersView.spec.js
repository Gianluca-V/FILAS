import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../../api/resources.js', () => ({
  getOrders: vi.fn(),
  getProducts: vi.fn(),
  patchOrderState: vi.fn(),
  updateOrder: vi.fn(),
}));

import { ApiError } from '../../../api/client.js';
import { getOrders, patchOrderState } from '../../../api/resources.js';
import AdminOrdersView from '../AdminOrdersView.vue';

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

function pendingOrder(overrides = {}) {
  return {
    orderID: '1',
    orderTotal: '100',
    orderState: 'pending',
    orderStartDate: '2024-01-01 10:00:00',
    orderFinishDate: null,
    orderName: 'Ana',
    orderPhone: '11-2345-6789',
    products: [{ productName: 'Pan casero', productPrice: 100, productQuantity: 1 }],
    ...overrides,
  };
}

describe('AdminOrdersView', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders orders and their line items from a mocked resources call', async () => {
    getOrders.mockResolvedValue([pendingOrder()]);

    const wrapper = mount(AdminOrdersView);
    await flushPromises();

    expect(wrapper.text()).toContain('Ana');
    expect(wrapper.text()).toContain('Pan casero');
  });

  it('calls patchOrderState with "finished" when finishing a pending order', async () => {
    getOrders.mockResolvedValue([pendingOrder()]);
    patchOrderState.mockResolvedValue({ message: 'Order patched successfully' });

    const wrapper = mount(AdminOrdersView);
    await flushPromises();

    await wrapper.find('[data-testid="finish-1"]').trigger('click');
    await flushPromises();

    expect(patchOrderState).toHaveBeenCalledWith('1', 'finished');
  });

  it('shows a Spanish message when patching a non-pending order returns 409', async () => {
    getOrders.mockResolvedValue([pendingOrder()]);
    patchOrderState.mockRejectedValue(new ApiError('order 1 is not pending', 409, { message: 'order 1 is not pending' }));

    const wrapper = mount(AdminOrdersView);
    await flushPromises();

    await wrapper.find('[data-testid="cancel-1"]').trigger('click');
    await flushPromises();

    expect(wrapper.text()).toContain('ya esté finalizada/cancelada');
  });
});
