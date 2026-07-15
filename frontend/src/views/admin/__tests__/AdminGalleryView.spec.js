import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../../api/resources.js', () => ({
  getGallery: vi.fn(),
  createGalleryItem: vi.fn(),
  updateGalleryItem: vi.fn(),
  deleteGalleryItem: vi.fn(),
}));

import { createGalleryItem, deleteGalleryItem, getGallery } from '../../../api/resources.js';
import AdminGalleryView from '../AdminGalleryView.vue';

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('AdminGalleryView', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders gallery rows from a mocked resources call', async () => {
    getGallery.mockResolvedValue([
      { ID: '1', Image: 'https://example.com/one.jpg' },
      { ID: '2', Image: 'https://example.com/two.jpg' },
    ]);

    const wrapper = mount(AdminGalleryView);
    await flushPromises();

    expect(wrapper.findAll('tbody tr')).toHaveLength(2);
    expect(wrapper.find('img[src="https://example.com/one.jpg"]').exists()).toBe(true);
    expect(wrapper.find('img[src="https://example.com/two.jpg"]').exists()).toBe(true);
  });

  it('creates a gallery item with only the Image field, matching the resource schema', async () => {
    getGallery.mockResolvedValue([]);
    createGalleryItem.mockResolvedValue({ ID: '9' });

    const wrapper = mount(AdminGalleryView);
    await flushPromises();

    await wrapper.find('.admin-resource__add').trigger('click');
    await wrapper.find('#Image').setValue('https://example.com/new.jpg');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(createGalleryItem).toHaveBeenCalledWith({ Image: 'https://example.com/new.jpg' });
    expect(createGalleryItem).toHaveBeenCalledTimes(1);
  });

  it('calls deleteGalleryItem with the right id once a delete is confirmed', async () => {
    getGallery.mockResolvedValue([{ ID: '5', Image: 'https://example.com/five.jpg' }]);
    deleteGalleryItem.mockResolvedValue(undefined);

    const wrapper = mount(AdminGalleryView);
    await flushPromises();

    await wrapper.find('[data-testid="delete-5"]').trigger('click');
    await wrapper.find('[data-testid="confirm-delete"]').trigger('click');
    await flushPromises();

    expect(deleteGalleryItem).toHaveBeenCalledWith('5');
  });
});
