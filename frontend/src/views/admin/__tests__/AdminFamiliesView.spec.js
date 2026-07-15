import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../../api/resources.js', () => ({
  getFamily: vi.fn(),
  createFamilyItem: vi.fn(),
  updateFamilyItem: vi.fn(),
  deleteFamilyItem: vi.fn(),
}));

import { createFamilyItem, deleteFamilyItem, getFamily } from '../../../api/resources.js';
import AdminFamiliesView from '../AdminFamiliesView.vue';

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('AdminFamiliesView', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders family/workshop rows from a mocked resources call', async () => {
    getFamily.mockResolvedValue([
      { ID: '1', Body: 'Taller de ceramica', Category: 'Taller protegido', Image: null },
      { ID: '2', Body: 'Centro de dia norte', Category: 'Centro de dia', Image: null },
    ]);

    const wrapper = mount(AdminFamiliesView);
    await flushPromises();

    expect(wrapper.text()).toContain('Taller de ceramica');
    expect(wrapper.text()).toContain('Centro de dia norte');
  });

  it('constrains Category to the two valid enum options in the create form', async () => {
    getFamily.mockResolvedValue([]);

    const wrapper = mount(AdminFamiliesView);
    await flushPromises();

    await wrapper.find('.admin-resource__add').trigger('click');
    const options = wrapper.find('#Category').findAll('option').map((option) => option.text());

    expect(options).toEqual(['Centro de dia', 'Taller protegido']);
  });

  it('creates a family item sending Body/Category/Image with the selected enum value', async () => {
    getFamily.mockResolvedValue([]);
    createFamilyItem.mockResolvedValue({ ID: '9' });

    const wrapper = mount(AdminFamiliesView);
    await flushPromises();

    await wrapper.find('.admin-resource__add').trigger('click');
    await wrapper.find('#Body').setValue('Taller de huerta');
    await wrapper.find('#Category').setValue('Taller protegido');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(createFamilyItem).toHaveBeenCalledWith({
      Body: 'Taller de huerta',
      Category: 'Taller protegido',
      Image: null,
    });
  });

  it('calls deleteFamilyItem with the right id once a delete is confirmed', async () => {
    getFamily.mockResolvedValue([{ ID: '7', Body: 'Taller de prueba', Category: 'Centro de dia', Image: null }]);
    deleteFamilyItem.mockResolvedValue(undefined);

    const wrapper = mount(AdminFamiliesView);
    await flushPromises();

    await wrapper.find('[data-testid="delete-7"]').trigger('click');
    await wrapper.find('[data-testid="confirm-delete"]').trigger('click');
    await flushPromises();

    expect(deleteFamilyItem).toHaveBeenCalledWith('7');
  });
});
