import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../../api/resources.js', () => ({
  getOrganizations: vi.fn(),
  createOrganization: vi.fn(),
  updateOrganization: vi.fn(),
  deleteOrganization: vi.fn(),
}));

import { createOrganization, deleteOrganization, getOrganizations } from '../../../api/resources.js';
import AdminOrganizationsView from '../AdminOrganizationsView.vue';

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('AdminOrganizationsView', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders organization rows from a mocked resources call', async () => {
    getOrganizations.mockResolvedValue([
      { ID: '1', Title: 'Fundacion Norte', Description: 'Ayuda comunitaria', Image: 'https://example.com/n.jpg' },
      { ID: '2', Title: 'Fundacion Sur', Description: null, Image: 'https://example.com/s.jpg' },
    ]);

    const wrapper = mount(AdminOrganizationsView);
    await flushPromises();

    expect(wrapper.text()).toContain('Fundacion Norte');
    expect(wrapper.text()).toContain('Fundacion Sur');
  });

  it('sends Description as null when the field is left empty', async () => {
    getOrganizations.mockResolvedValue([]);
    createOrganization.mockResolvedValue({ ID: '9' });

    const wrapper = mount(AdminOrganizationsView);
    await flushPromises();

    await wrapper.find('.admin-resource__add').trigger('click');
    await wrapper.find('#Title').setValue('Fundacion Este');
    await wrapper.find('#Image').setValue('https://example.com/e.jpg');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(createOrganization).toHaveBeenCalledWith({
      Title: 'Fundacion Este',
      Description: null,
      Image: 'https://example.com/e.jpg',
    });
  });

  it('sends the Description text as-is when the field is filled', async () => {
    getOrganizations.mockResolvedValue([]);
    createOrganization.mockResolvedValue({ ID: '10' });

    const wrapper = mount(AdminOrganizationsView);
    await flushPromises();

    await wrapper.find('.admin-resource__add').trigger('click');
    await wrapper.find('#Title').setValue('Fundacion Oeste');
    await wrapper.find('#Description').setValue('Trabajo social');
    await wrapper.find('#Image').setValue('https://example.com/o.jpg');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(createOrganization).toHaveBeenCalledWith({
      Title: 'Fundacion Oeste',
      Description: 'Trabajo social',
      Image: 'https://example.com/o.jpg',
    });
  });

  it('calls deleteOrganization with the right id once a delete is confirmed', async () => {
    getOrganizations.mockResolvedValue([
      { ID: '3', Title: 'Fundacion Test', Description: null, Image: 'https://example.com/t.jpg' },
    ]);
    deleteOrganization.mockResolvedValue(undefined);

    const wrapper = mount(AdminOrganizationsView);
    await flushPromises();

    await wrapper.find('[data-testid="delete-3"]').trigger('click');
    await wrapper.find('[data-testid="confirm-delete"]').trigger('click');
    await flushPromises();

    expect(deleteOrganization).toHaveBeenCalledWith('3');
  });
});
