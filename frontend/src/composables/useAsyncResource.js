// Shared "fetch-on-mount with status tracking" logic, extracted from the
// near-identical wiring duplicated across NewsListView, GalleryView,
// FamiliesView, and OrganizationsView (PR8 corrective — see the 3-lens gate
// duplication finding).
import { onMounted, ref } from 'vue';

import { isNotFound } from '../utils/httpErrors.js';

// status: 'loading' | 'ready' | 'empty' | 'error'.
//
// `fetchFn` is called once, onMounted, and again whenever the returned
// `refresh()` is invoked — admin views (task 8.1) call it after a
// create/update/delete mutation to reload the list without a full
// component remount. When `notFoundIsEmpty` is true, a 404 rejection (see
// isNotFound) is treated as "no items yet" (status 'empty') instead of a
// failure — this mirrors the backend quirk where news/gallery/family/
// organizations return 404 for an empty list rather than 200 []. Any
// other rejection (or a 404 when notFoundIsEmpty is false) maps to
// 'error'.
export function useAsyncResource(fetchFn, { notFoundIsEmpty = false } = {}) {
  const data = ref([]);
  const status = ref('loading');

  async function load() {
    status.value = 'loading';
    try {
      data.value = await fetchFn();
      status.value = 'ready';
    } catch (error) {
      if (notFoundIsEmpty && isNotFound(error)) {
        data.value = [];
        status.value = 'empty';
      } else {
        status.value = 'error';
      }
    }
  }

  onMounted(load);

  return { data, status, refresh: load };
}
