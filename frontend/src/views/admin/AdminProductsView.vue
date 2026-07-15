<script setup>
// Ports js/admin/productosController.mjs + its .html partial: list, create,
// edit, delete products. GET /api/products is never 404-on-empty (see
// backend/internal/handler/rest/product.go), so no notFoundIsEmpty flag.
import { ref } from 'vue';

import { createProduct, deleteProduct, getProducts, updateProduct } from '../../api/resources.js';
import ConfirmDialog from '../../components/ConfirmDialog.vue';
import ResourceForm from '../../components/ResourceForm.vue';
import { useAsyncResource } from '../../composables/useAsyncResource.js';

const { data: products, status, refresh } = useAsyncResource(getProducts);

const fields = [
  { key: 'Name', label: 'Nombre', type: 'text', required: true },
  { key: 'Price', label: 'Precio', type: 'number', required: true },
  { key: 'Image', label: 'Imagen', type: 'text' },
  { key: 'Stock', label: 'Stock', type: 'number', required: true },
];

const isFormOpen = ref(false);
const formMode = ref('create');
const formValues = ref({});
const editingId = ref(null);
const formError = ref('');

function openCreateForm() {
  formMode.value = 'create';
  formValues.value = { Name: '', Price: '', Image: '', Stock: '' };
  editingId.value = null;
  formError.value = '';
  isFormOpen.value = true;
}

function openEditForm(product) {
  formMode.value = 'edit';
  formValues.value = { Name: product.Name, Price: product.Price, Image: product.Image ?? '', Stock: product.Stock };
  editingId.value = product.ID;
  formError.value = '';
  isFormOpen.value = true;
}

function closeForm() {
  isFormOpen.value = false;
}

async function handleSubmit(values) {
  const payload = {
    Name: values.Name,
    Price: Number(values.Price),
    Image: values.Image || null,
    Stock: Number(values.Stock),
  };

  try {
    if (formMode.value === 'create') {
      await createProduct(payload);
    } else {
      await updateProduct(editingId.value, payload);
    }
    isFormOpen.value = false;
    await refresh();
  } catch (error) {
    formError.value = error.body?.message || 'No se pudo guardar el producto.';
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
  await deleteProduct(id);
  await refresh();
}
</script>

<template>
  <section class="admin-resource">
    <div class="admin-resource__header">
      <h3 class="admin-resource__title">Productos</h3>
      <button type="button" class="admin-resource__add" @click="openCreateForm">+</button>
    </div>

    <p v-if="status === 'loading'" class="admin-resource__status">Cargando productos...</p>
    <p v-else-if="status === 'error'" class="admin-resource__status">
      No pudimos cargar los productos. Intentá nuevamente más tarde.
    </p>
    <p v-else-if="products.length === 0" class="admin-resource__status">No hay productos cargados.</p>

    <table v-else class="admin-table">
      <thead>
        <tr>
          <th>#</th>
          <th>Nombre</th>
          <th>Precio</th>
          <th>Imagen</th>
          <th>Stock</th>
          <th>Acciones</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="product in products" :key="product.ID">
          <td>{{ product.ID }}</td>
          <td>{{ product.Name }}</td>
          <td>{{ product.Price }}</td>
          <td>
            <img class="admin-table__image" :src="product.Image || '/assets/default-img.png'" :alt="product.Name">
          </td>
          <td>{{ product.Stock }}</td>
          <td class="admin-table__actions">
            <button type="button" :data-testid="`edit-${product.ID}`" @click="openEditForm(product)">Editar</button>
            <button type="button" :data-testid="`delete-${product.ID}`" @click="requestDelete(product.ID)">
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
      :title="formMode === 'create' ? 'Agregar producto' : 'Editar producto'"
      :error="formError"
      @update:model-value="formValues = $event"
      @submit="handleSubmit"
      @cancel="closeForm"
    />

    <ConfirmDialog
      :open="confirmingId !== null"
      message="¿Estás seguro? Esto eliminará el producto para siempre."
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
