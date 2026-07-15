import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../../api/resources.js', () => ({
  getNews: vi.fn(),
  createNewsItem: vi.fn(),
  updateNewsItem: vi.fn(),
  deleteNewsItem: vi.fn(),
}));

import { createNewsItem, deleteNewsItem, getNews } from '../../../api/resources.js';
import AdminNewsView from '../AdminNewsView.vue';

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('AdminNewsView', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders news rows from a mocked resources call', async () => {
    getNews.mockResolvedValue([
      { ID: '1', Title: 'Nueva sede', Body: 'Abrimos una nueva sede', Image: 'https://example.com/a.jpg' },
      { ID: '2', Title: 'Evento anual', Body: 'Este año celebramos', Image: 'https://example.com/b.jpg' },
    ]);

    const wrapper = mount(AdminNewsView);
    await flushPromises();

    expect(wrapper.text()).toContain('Nueva sede');
    expect(wrapper.text()).toContain('Evento anual');
  });

  it('marks Title, Body, and Image as required and sends all three on create', async () => {
    getNews.mockResolvedValue([]);
    createNewsItem.mockResolvedValue({ ID: '9' });

    const wrapper = mount(AdminNewsView);
    await flushPromises();

    await wrapper.find('.admin-resource__add').trigger('click');

    expect(wrapper.find('#Title').attributes('required')).toBeDefined();
    expect(wrapper.find('#Body').attributes('required')).toBeDefined();
    expect(wrapper.find('#Image').attributes('required')).toBeDefined();

    await wrapper.find('#Title').setValue('Nueva noticia');
    await wrapper.find('#Body').setValue('Contenido de la noticia');
    await wrapper.find('#Image').setValue('https://example.com/new.jpg');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(createNewsItem).toHaveBeenCalledWith({
      Title: 'Nueva noticia',
      Body: 'Contenido de la noticia',
      Image: 'https://example.com/new.jpg',
    });
  });

  it('calls deleteNewsItem with the right id once a delete is confirmed', async () => {
    getNews.mockResolvedValue([
      { ID: '4', Title: 'Noticia de prueba', Body: 'Cuerpo', Image: 'https://example.com/t.jpg' },
    ]);
    deleteNewsItem.mockResolvedValue(undefined);

    const wrapper = mount(AdminNewsView);
    await flushPromises();

    await wrapper.find('[data-testid="delete-4"]').trigger('click');
    await wrapper.find('[data-testid="confirm-delete"]').trigger('click');
    await flushPromises();

    expect(deleteNewsItem).toHaveBeenCalledWith('4');
  });
});
