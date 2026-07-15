<script setup>
// Ports js/admin/familiaController.mjs + its .html partial: list, create,
// edit, delete family (Taller protegido / Centro de dia) items. GET
// /api/family is 404-on-empty (see backend/internal/handler/rest/family.go),
// like the public FamiliesView (notFoundIsEmpty: true). Category is a
// fixed 2-value enum (domain.ValidFamilyCategory) — rendered as a select,
// never free text.
import { createFamilyItem, deleteFamilyItem, getFamily, updateFamilyItem } from '../../api/resources.js';
import ConfirmDialog from '../../components/ConfirmDialog.vue';
import ResourceForm from '../../components/ResourceForm.vue';
import { useAsyncResource } from '../../composables/useAsyncResource.js';
import { useResourceCrud } from '../../composables/useResourceCrud.js';

const { data: items, status, refresh } = useAsyncResource(getFamily, { notFoundIsEmpty: true });

const CATEGORY_OPTIONS = ['Centro de dia', 'Taller protegido'];

const fields = [
  { key: 'Body', label: 'Cuerpo', type: 'textarea', required: true },
  { key: 'Category', label: 'Categoría', type: 'select', options: CATEGORY_OPTIONS, required: true },
  { key: 'Image', label: 'Imagen', type: 'text' },
];

const {
  isFormOpen,
  formMode,
  formValues,
  formError,
  openCreateForm,
  openEditForm,
  closeForm,
  handleSubmit,
  confirmingId,
  requestDelete,
  cancelDelete,
  confirmDelete,
} = useResourceCrud({
  createFn: createFamilyItem,
  updateFn: updateFamilyItem,
  deleteFn: deleteFamilyItem,
  refresh,
  emptyValues: { Body: '', Category: CATEGORY_OPTIONS[0], Image: '' },
  toFormValues: (item) => ({ Body: item.Body, Category: item.Category, Image: item.Image ?? '' }),
  toPayload: (values) => ({ Body: values.Body, Category: values.Category, Image: values.Image || null }),
  errorMessage: 'No se pudo guardar el taller.',
});
</script>

<template>
  <section class="admin-resource">
    <div class="admin-resource__header">
      <h3 class="admin-resource__title">Familias (CD/TP)</h3>
      <button type="button" class="admin-resource__add" @click="openCreateForm">+</button>
    </div>

    <p v-if="status === 'loading'" class="admin-resource__status">Cargando familias...</p>
    <p v-else-if="status === 'error'" class="admin-resource__status">
      No pudimos cargar las familias. Intentá nuevamente más tarde.
    </p>
    <p v-else-if="status === 'empty'" class="admin-resource__status">Todavía no hay talleres cargados.</p>

    <table v-else class="admin-table">
      <thead>
        <tr>
          <th>#</th>
          <th>Cuerpo</th>
          <th>Categoría</th>
          <th>Imagen</th>
          <th>Acciones</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" :key="item.ID">
          <td>{{ item.ID }}</td>
          <td>{{ item.Body }}</td>
          <td>{{ item.Category }}</td>
          <td><img class="admin-table__image" :src="item.Image || '/assets/default-img.png'" alt=""></td>
          <td class="admin-table__actions">
            <button type="button" :data-testid="`edit-${item.ID}`" @click="openEditForm(item)">Editar</button>
            <button type="button" :data-testid="`delete-${item.ID}`" @click="requestDelete(item.ID)">
              Eliminar
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <ResourceForm
      v-if="isFormOpen"
      :fields="fields"
      :model-value="formValues"
      :title="formMode === 'create' ? 'Agregar taller' : 'Editar taller'"
      :error="formError"
      @update:model-value="formValues = $event"
      @submit="handleSubmit"
      @cancel="closeForm"
    />

    <ConfirmDialog
      :open="confirmingId !== null"
      message="¿Estás seguro? Esto eliminará el taller para siempre."
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
  </section>
</template>

<style scoped lang="scss">
.admin-resource__header {
  @include flex-row(space-between, center);
  padding-inline: 1rem;
}

.admin-resource__title {
  font-size: clamp(1.75rem, 5vw, 2.5rem);
  font-weight: 700;
}

.admin-resource__add {
  background-color: $color-accent;
  color: $color-background;
  font-size: 1.5rem;
  border-radius: 0.5rem;
  width: 3rem;
  height: 3rem;
}

.admin-resource__status {
  text-align: center;
  padding-block: 2rem;
  font-size: 1.125rem;
}

.admin-table {
  width: calc(100% - 2rem);
  text-align: center;
  margin: 1rem auto;
  border-collapse: collapse;

  th,
  td {
    padding: 0.75rem;
  }
}

.admin-table__image {
  max-width: 100px;
}

.admin-table__actions {
  @include flex-row(space-evenly, center, 0.5rem);
}
</style>
