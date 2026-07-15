import { mount } from '@vue/test-utils';
import { describe, expect, it, vi } from 'vitest';

vi.mock('../../../api/resources.js', () => ({
  getNews: vi.fn(),
}));

import { getNews } from '../../../api/resources.js';
import { ApiError } from '../../../api/client.js';
import NewsListView from '../NewsListView.vue';

// router-link isn't registered in this unit test (no router installed), so
// it's stubbed to a plain anchor — only NewsListView's own rendering logic
// is under test here.
const globalStubs = { RouterLink: { template: '<a><slot /></a>' } };

describe('NewsListView', () => {
  it('renders the news list from a mocked resources call', async () => {
    getNews.mockResolvedValue([
      { ID: '1', Title: 'Primera noticia', Image: 'assets/n1.jpg' },
      { ID: '2', Title: 'Segunda noticia', Image: 'assets/n2.jpg' },
    ]);

    const wrapper = mount(NewsListView, { global: { stubs: globalStubs } });
    await flushPromises();

    expect(wrapper.text()).toContain('Primera noticia');
    expect(wrapper.text()).toContain('Segunda noticia');
  });

  it('shows the empty state instead of an error when the API 404s (no news yet)', async () => {
    getNews.mockRejectedValue(new ApiError('Not Found', 404, { message: 'No news found' }));

    const wrapper = mount(NewsListView, { global: { stubs: globalStubs } });
    await flushPromises();

    expect(wrapper.text()).toContain('Todavía no hay noticias publicadas.');
    expect(wrapper.text()).not.toContain('No pudimos cargar');
  });

  it('shows an error message on a non-404 failure', async () => {
    getNews.mockRejectedValue(new ApiError('Server error', 500, { message: 'boom' }));

    const wrapper = mount(NewsListView, { global: { stubs: globalStubs } });
    await flushPromises();

    expect(wrapper.text()).toContain('No pudimos cargar las noticias');
  });
});

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}
