// Shared admin CRUD state machine, extracted from the near-identical
// isFormOpen/formMode/formValues/editingId/formError +
// openCreateForm/openEditForm/closeForm/handleSubmit +
// confirmingId/requestDelete/cancelDelete/confirmDelete block duplicated
// across AdminProductsView, AdminGalleryView, AdminFamiliesView,
// AdminOrganizationsView, and AdminNewsView. Only the field mapping
// (toFormValues/toPayload) and the create/update/delete API calls differ
// per resource — those are supplied by the caller.
import { ref } from 'vue';

// options:
//   createFn(payload) / updateFn(id, payload) / deleteFn(id) — required,
//     resource-specific API calls (see api/resources.js).
//   refresh() — required, re-fetches the list after a mutation (typically
//     the `refresh` returned by useAsyncResource).
//   emptyValues — object (or factory function) used to seed formValues on
//     openCreateForm.
//   toFormValues(item) — maps a list item to the edit-form's initial
//     values. Defaults to a shallow copy of the item.
//   toPayload(values) — maps form values to the create/update API payload.
//     Defaults to a shallow copy of the values.
//   getId(item) — extracts the identifier used for editingId/updateFn.
//     Defaults to `item.ID`, matching every admin resource shape.
//   errorMessage — fallback message shown when a failed submit's error has
//     no body.message (mirrors each view's own Spanish copy).
export function useResourceCrud({
  createFn,
  updateFn,
  deleteFn,
  refresh,
  emptyValues = {},
  toFormValues = (item) => ({ ...item }),
  toPayload = (values) => ({ ...values }),
  getId = (item) => item.ID,
  errorMessage = 'No se pudo guardar el registro.',
}) {
  const isFormOpen = ref(false);
  const formMode = ref('create');
  const formValues = ref({});
  const editingId = ref(null);
  const formError = ref('');

  function resolveEmptyValues() {
    return typeof emptyValues === 'function' ? emptyValues() : { ...emptyValues };
  }

  function openCreateForm() {
    formMode.value = 'create';
    formValues.value = resolveEmptyValues();
    editingId.value = null;
    formError.value = '';
    isFormOpen.value = true;
  }

  function openEditForm(item) {
    formMode.value = 'edit';
    formValues.value = toFormValues(item);
    editingId.value = getId(item);
    formError.value = '';
    isFormOpen.value = true;
  }

  function closeForm() {
    isFormOpen.value = false;
  }

  async function handleSubmit(values) {
    const payload = toPayload(values);

    try {
      if (formMode.value === 'create') {
        await createFn(payload);
      } else {
        await updateFn(editingId.value, payload);
      }
      isFormOpen.value = false;
      await refresh();
    } catch (error) {
      formError.value = error.body?.message || errorMessage;
    }
  }

  const confirmingId = ref(null);

  function requestDelete(id) {
    confirmingId.value = id;
  }

  function cancelDelete() {
    confirmingId.value = null;
  }

  async function confirmDelete() {
    const id = confirmingId.value;
    confirmingId.value = null;
    await deleteFn(id);
    await refresh();
  }

  return {
    isFormOpen,
    formMode,
    formValues,
    editingId,
    formError,
    openCreateForm,
    openEditForm,
    closeForm,
    handleSubmit,
    confirmingId,
    requestDelete,
    cancelDelete,
    confirmDelete,
  };
}
