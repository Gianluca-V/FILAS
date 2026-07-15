<script setup>
// Ports js/admin/galeriaController.mjs + its .html partial: list, create,
// delete gallery images (no edit in legacy beyond the same create form
// reused with prefilled values, kept here too). GET /api/gallery is
// 404-on-empty (see backend/internal/handler/rest/gallery.go), like the
// public GalleryView (notFoundIsEmpty: true).
import { ref } from 'vue';

import { createGalleryItem, deleteGalleryItem, getGallery, updateGalleryItem } from '../../api/resources.js';
import ConfirmDialog from '../../components/ConfirmDialog.vue';
import ResourceForm from '../../components/ResourceForm.vue';
import { useAsyncResource } from '../../composables/useAsyncResource.js';

const { data: images, status, refresh } = useAsyncResource(getGallery, { notFoundIsEmpty: true });

const fields = [{ key: 'Image', label: 'Imagen', type: 'text', required: true }];

const isFormOpen = ref(false);
const formMode = ref('create');
const formValues = ref({});
const editingId = ref(null);
const formError = ref('');

function openCreateForm() {
  formMode.value = 'create';
  formValues.value = { Image: '' };
  editingId.value = null;
  formError.value = '';
  isFormOpen.value = true;
}

function openEditForm(image) {
  formMode.value = 'edit';
  formValues.value = { Image: image.Image };
  editingId.value = image.ID;
  formError.value = '';
  isFormOpen.value = true;
}

function closeForm() {
  isFormOpen.value = false;
}

async function handleSubmit(values) {
  try {
    if (formMode.value === 'create') {
      await createGalleryItem({ Image: values.Image });
    } else {
      await updateGalleryItem(editingId.value, { Image: values.Image });
    }
    isFormOpen.value = false;
    await refresh();
  } catch (error) {
    formError.value = error.body?.message || 'No se pudo guardar la imagen.';
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
  await deleteGalleryItem(id);
  await refresh();
}
</script>

<template>
  <section class="admin-resource">
    <div class="admin-resource__header">
      <h3 class="admin-resource__title">Galería</h3>
      <button type="button" class="admin-resource__add" @click="openCreateForm">+</button>
    </div>

    <p v-if="status === 'loading'" class="admin-resource__status">Cargando galería...</p>
    <p v-else-if="status === 'error'" class="admin-resource__status">
      No pudimos cargar la galería. Intentá nuevamente más tarde.
    </p>
    <p v-else-if="status === 'empty'" class="admin-resource__status">Todavía no hay imágenes cargadas.</p>

    <table v-else class="admin-table">
      <thead>
        <tr>
          <th>#</th>
          <th>Imagen</th>
          <th>Acciones</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="image in images" :key="image.ID">
          <td>{{ image.ID }}</td>
          <td><img class="admin-table__image" :src="image.Image" alt=""></td>
          <td class="admin-table__actions">
            <button type="button" :data-testid="`edit-${image.ID}`" @click="openEditForm(image)">Editar</button>
            <button type="button" :data-testid="`delete-${image.ID}`" @click="requestDelete(image.ID)">
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
      :title="formMode === 'create' ? 'Agregar imagen' : 'Editar imagen'"
      :error="formError"
      @update:model-value="formValues = $event"
      @submit="handleSubmit"
      @cancel="closeForm"
    />

    <ConfirmDialog
      :open="confirmingId !== null"
      message="¿Estás seguro? Esto eliminará la imagen para siempre."
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
  max-width: 200px;
}

.admin-table__actions {
  @include flex-row(space-evenly, center, 0.5rem);
}
</style>
