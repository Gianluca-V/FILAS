<script setup>
// Ports js/admin/organizacionesController.mjs + its .html partial: list,
// create, edit, delete allied organizations. GET /api/organizations is
// 404-on-empty (see backend/internal/handler/rest/organization.go), like
// the public OrganizationsView (notFoundIsEmpty: true).
import { ref } from 'vue';

import { createOrganization, deleteOrganization, getOrganizations, updateOrganization } from '../../api/resources.js';
import ConfirmDialog from '../../components/ConfirmDialog.vue';
import ResourceForm from '../../components/ResourceForm.vue';
import { useAsyncResource } from '../../composables/useAsyncResource.js';

const { data: organizations, status, refresh } = useAsyncResource(getOrganizations, { notFoundIsEmpty: true });

const fields = [
  { key: 'Title', label: 'Titulo', type: 'text', required: true },
  { key: 'Description', label: 'Cuerpo', type: 'textarea' },
  { key: 'Image', label: 'Imagen', type: 'text', required: true },
];

const isFormOpen = ref(false);
const formMode = ref('create');
const formValues = ref({});
const editingId = ref(null);
const formError = ref('');

function openCreateForm() {
  formMode.value = 'create';
  formValues.value = { Title: '', Description: '', Image: '' };
  editingId.value = null;
  formError.value = '';
  isFormOpen.value = true;
}

function openEditForm(organization) {
  formMode.value = 'edit';
  formValues.value = {
    Title: organization.Title,
    Description: organization.Description ?? '',
    Image: organization.Image,
  };
  editingId.value = organization.ID;
  formError.value = '';
  isFormOpen.value = true;
}

function closeForm() {
  isFormOpen.value = false;
}

async function handleSubmit(values) {
  const payload = { Title: values.Title, Description: values.Description || null, Image: values.Image };

  try {
    if (formMode.value === 'create') {
      await createOrganization(payload);
    } else {
      await updateOrganization(editingId.value, payload);
    }
    isFormOpen.value = false;
    await refresh();
  } catch (error) {
    formError.value = error.body?.message || 'No se pudo guardar la organización.';
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
  await deleteOrganization(id);
  await refresh();
}
</script>

<template>
  <section class="admin-resource">
    <div class="admin-resource__header">
      <h3 class="admin-resource__title">Organizaciones</h3>
      <button type="button" class="admin-resource__add" @click="openCreateForm">+</button>
    </div>

    <p v-if="status === 'loading'" class="admin-resource__status">Cargando organizaciones...</p>
    <p v-else-if="status === 'error'" class="admin-resource__status">
      No pudimos cargar las organizaciones. Intentá nuevamente más tarde.
    </p>
    <p v-else-if="status === 'empty'" class="admin-resource__status">Todavía no hay organizaciones cargadas.</p>

    <table v-else class="admin-table">
      <thead>
        <tr>
          <th>#</th>
          <th>Titulo</th>
          <th>Cuerpo</th>
          <th>Imagen</th>
          <th>Acciones</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="organization in organizations" :key="organization.ID">
          <td>{{ organization.ID }}</td>
          <td>{{ organization.Title }}</td>
          <td>{{ organization.Description }}</td>
          <td><img class="admin-table__image" :src="organization.Image" alt=""></td>
          <td class="admin-table__actions">
            <button type="button" :data-testid="`edit-${organization.ID}`" @click="openEditForm(organization)">
              Editar
            </button>
            <button type="button" :data-testid="`delete-${organization.ID}`" @click="requestDelete(organization.ID)">
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
      :title="formMode === 'create' ? 'Agregar organización' : 'Editar organización'"
      :error="formError"
      @update:model-value="formValues = $event"
      @submit="handleSubmit"
      @cancel="closeForm"
    />

    <ConfirmDialog
      :open="confirmingId !== null"
      message="¿Estás seguro? Esto eliminará la organización para siempre."
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
