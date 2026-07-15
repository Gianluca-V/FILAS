// Shared "fetch-on-mount with status tracking" logic, extracted from the
// near-identical wiring duplicated across NewsListView, GalleryView,
// FamiliesView, and OrganizationsView (PR8 corrective — see the 3-lens gate
// duplication finding).
import { onMounted, ref } from 'vue';

import { isNotFound } from '../utils/httpErrors.js';

// status: 'loading' | 'ready' | 'empty' | 'error'.
//
// `fetchFn` is called once, onMounted. When `notFoundIsEmpty` is true, a
// 404 rejection (see isNotFound) is treated as "no items yet" (status
// 'empty') instead of a failure — this mirrors the backend quirk where
// news/gallery/family/organizations return 404 for an empty list rather
// than 200 []. Any other rejection (or a 404 when notFoundIsEmpty is
// false) maps to 'error'.
export function useAsyncResource(fetchFn, { notFoundIsEmpty = false } = {}) {
  const data = ref([]);
  const status = ref('loading');

  onMounted(async () => {
    try {
      data.value = await fetchFn();
      status.value = 'ready';
    } catch (error) {
      if (notFoundIsEmpty && isNotFound(error)) {
        status.value = 'empty';
      } else {
        status.value = 'error';
      }
    }
  });

  return { data, status };
}
