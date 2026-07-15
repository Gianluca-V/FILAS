import { mount } from '@vue/test-utils';
import { defineComponent, h } from 'vue';
import { describe, expect, it, vi } from 'vitest';

import { ApiError } from '../../api/client.js';
import { useAsyncResource } from '../useAsyncResource.js';

// useAsyncResource relies on onMounted, so it must run inside a real
// component setup context — this small host component gives the composable
// that context and hands the returned refs back out to the test.
function mountResource(fetchFn, options) {
  let result;
  const TestComponent = defineComponent({
    setup() {
      result = useAsyncResource(fetchFn, options);
      return () => h('div');
    },
  });
  mount(TestComponent);
  return result;
}

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('useAsyncResource', () => {
  it('resolves to ready with the fetched data', async () => {
    const fetchFn = vi.fn().mockResolvedValue([{ ID: '1' }]);

    const { data, status } = mountResource(fetchFn);
    await flushPromises();

    expect(status.value).toBe('ready');
    expect(data.value).toEqual([{ ID: '1' }]);
  });

  it('maps a 404 rejection to empty when notFoundIsEmpty is true', async () => {
    const fetchFn = vi.fn().mockRejectedValue(new ApiError('Not Found', 404, null));

    const { status } = mountResource(fetchFn, { notFoundIsEmpty: true });
    await flushPromises();

    expect(status.value).toBe('empty');
  });

  it('maps a 404 rejection to error when notFoundIsEmpty is not set', async () => {
    const fetchFn = vi.fn().mockRejectedValue(new ApiError('Not Found', 404, null));

    const { status } = mountResource(fetchFn);
    await flushPromises();

    expect(status.value).toBe('error');
  });

  it('maps a 500 rejection to error regardless of notFoundIsEmpty', async () => {
    const fetchFn = vi.fn().mockRejectedValue(new ApiError('Server error', 500, null));

    const { status } = mountResource(fetchFn, { notFoundIsEmpty: true });
    await flushPromises();

    expect(status.value).toBe('error');
  });

  it('refresh() re-invokes fetchFn and replaces data after the initial load', async () => {
    const fetchFn = vi
      .fn()
      .mockResolvedValueOnce([{ ID: '1' }])
      .mockResolvedValueOnce([{ ID: '1' }, { ID: '2' }]);

    const { data, status, refresh } = mountResource(fetchFn);
    await flushPromises();

    expect(fetchFn).toHaveBeenCalledTimes(1);
    expect(data.value).toEqual([{ ID: '1' }]);

    const refreshPromise = refresh();
    expect(status.value).toBe('loading');
    await refreshPromise;
    await flushPromises();

    expect(fetchFn).toHaveBeenCalledTimes(2);
    expect(status.value).toBe('ready');
    expect(data.value).toEqual([{ ID: '1' }, { ID: '2' }]);
  });
});
