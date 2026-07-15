import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../../api/resources.js', () => ({
  getProducts: vi.fn(),
  createProduct: vi.fn(),
  updateProduct: vi.fn(),
  deleteProduct: vi.fn(),
}));

import { deleteProduct, getProducts } from '../../../api/resources.js';
import AdminProductsView from '../AdminProductsView.vue';

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('AdminProductsView', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders product rows from a mocked resources call', async () => {
    getProducts.mockResolvedValue([
      { ID: '1', Name: 'Pan casero', Price: '100', Stock: '10', Image: null },
      { ID: '2', Name: 'Dulce de leche', Price: '200', Stock: '5', Image: null },
    ]);

    const wrapper = mount(AdminProductsView);
    await flushPromises();

    expect(wrapper.text()).toContain('Pan casero');
    expect(wrapper.text()).toContain('Dulce de leche');
  });

  it('shows an error message when the products call fails', async () => {
    getProducts.mockRejectedValue(new Error('boom'));

    const wrapper = mount(AdminProductsView);
    await flushPromises();

    expect(wrapper.text()).toContain('No pudimos cargar los productos');
  });

  it('calls deleteProduct with the right id once a delete is confirmed', async () => {
    getProducts.mockResolvedValue([{ ID: '7', Name: 'Producto de prueba', Price: '50', Stock: '3', Image: null }]);
    deleteProduct.mockResolvedValue({ message: 'Product deleted successfully' });

    const wrapper = mount(AdminProductsView);
    await flushPromises();

    await wrapper.find('[data-testid="delete-7"]').trigger('click');
    await wrapper.find('[data-testid="confirm-delete"]').trigger('click');
    await flushPromises();

    expect(deleteProduct).toHaveBeenCalledWith('7');
  });

  it('does not call deleteProduct when the confirmation is canceled', async () => {
    getProducts.mockResolvedValue([{ ID: '9', Name: 'Otro producto', Price: '30', Stock: '1', Image: null }]);

    const wrapper = mount(AdminProductsView);
    await flushPromises();

    await wrapper.find('[data-testid="delete-9"]').trigger('click');
    await wrapper.find('[data-testid="confirm-cancel"]').trigger('click');
    await flushPromises();

    expect(deleteProduct).not.toHaveBeenCalled();
  });
});
