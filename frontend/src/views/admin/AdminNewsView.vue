<script setup>
// Ports js/admin/noticiasController.mjs + its .html partial: list, create,
// edit, delete news. GET /api/news is 404-on-empty (see
// backend/internal/handler/rest/news.go), like the public NewsListView
// (notFoundIsEmpty: true). Title/Body/Image are all required on write —
// the backend's Create/Update reject a missing one with the verbatim
// legacy message "Missing Title, Body, or Image parameter" (see
// backend/internal/handler/rest/news.go), surfaced here as-is via
// error.body.message.
import { createNewsItem, deleteNewsItem, getNews, updateNewsItem } from '../../api/resources.js';
import ConfirmDialog from '../../components/ConfirmDialog.vue';
import ResourceForm from '../../components/ResourceForm.vue';
import { useAsyncResource } from '../../composables/useAsyncResource.js';
import { useResourceCrud } from '../../composables/useResourceCrud.js';

const { data: news, status, refresh } = useAsyncResource(getNews, { notFoundIsEmpty: true });

const fields = [
  { key: 'Title', label: 'Titulo', type: 'text', required: true },
  { key: 'Body', label: 'Cuerpo', type: 'textarea', required: true },
  { key: 'Image', label: 'Imagen', type: 'text', required: true },
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
  createFn: createNewsItem,
  updateFn: updateNewsItem,
  deleteFn: deleteNewsItem,
  refresh,
  emptyValues: { Title: '', Body: '', Image: '' },
  toFormValues: (article) => ({ Title: article.Title, Body: article.Body ?? '', Image: article.Image ?? '' }),
  toPayload: (values) => ({ Title: values.Title, Body: values.Body, Image: values.Image }),
  errorMessage: 'No se pudo guardar la noticia.',
});
</script>

<template>
  <section class="admin-resource">
    <div class="admin-resource__header">
      <h3 class="admin-resource__title">Noticias</h3>
      <button type="button" class="admin-resource__add" @click="openCreateForm">+</button>
    </div>

    <p v-if="status === 'loading'" class="admin-resource__status">Cargando noticias...</p>
    <p v-else-if="status === 'error'" class="admin-resource__status">
      No pudimos cargar las noticias. Intentá nuevamente más tarde.
    </p>
    <p v-else-if="status === 'empty'" class="admin-resource__status">Todavía no hay noticias cargadas.</p>

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
        <tr v-for="article in news" :key="article.ID">
          <td>{{ article.ID }}</td>
          <td>{{ article.Title }}</td>
          <td>{{ article.Body }}</td>
          <td><img class="admin-table__image" :src="article.Image || '/assets/default-img.png'" alt=""></td>
          <td class="admin-table__actions">
            <button type="button" :data-testid="`edit-${article.ID}`" @click="openEditForm(article)">Editar</button>
            <button type="button" :data-testid="`delete-${article.ID}`" @click="requestDelete(article.ID)">
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
      :title="formMode === 'create' ? 'Agregar noticia' : 'Editar noticia'"
      :error="formError"
      @update:model-value="formValues = $event"
      @submit="handleSubmit"
      @cancel="closeForm"
    />

    <ConfirmDialog
      :open="confirmingId !== null"
      message="¿Estás seguro? Esto eliminará la noticia para siempre."
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
