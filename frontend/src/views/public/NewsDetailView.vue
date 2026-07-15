<script setup>
import { onMounted, ref, watch } from 'vue';

import { getNewsItem } from '../../api/resources.js';
import { resolveAssetPath } from '../../utils/assets.js';
import { isNotFound } from '../../utils/httpErrors.js';

// Registered with `props: true` in router/index.js, so `id` arrives as a
// route param prop (replaces the legacy `?ID=` query-string reader in
// js/newsID.mjs).
const props = defineProps({
  id: {
    type: String,
    required: true,
  },
});

const status = ref('loading');
const article = ref(null);

async function loadArticle(id) {
  status.value = 'loading';
  try {
    // GET /api/news/:id wraps the single result in a one-element array
    // (see backend dto.NewsResponse doc comment) — unwrap it here.
    const [item] = await getNewsItem(id);
    article.value = item;
    status.value = 'ready';
  } catch (error) {
    if (isNotFound(error)) {
      status.value = 'not-found';
    } else {
      status.value = 'error';
    }
  }
}

onMounted(() => loadArticle(props.id));
watch(() => props.id, loadArticle);
</script>

<template>
  <p v-if="status === 'loading'" class="article-status">Cargando noticia...</p>
  <p v-else-if="status === 'not-found'" class="article-status">No se encontró la noticia solicitada.</p>
  <p v-else-if="status === 'error'" class="article-status">
    No pudimos cargar la noticia. Intentá nuevamente más tarde.
  </p>

  <div v-else class="article">
    <img :src="resolveAssetPath(article.Image)" :alt="article.Title" class="article__image">
    <h1 class="article__title">{{ article.Title }}</h1>
    <p class="article__content">{{ article.Body }}</p>
  </div>
</template>

<style scoped lang="scss">
.article-status {
  text-align: center;
  padding-block: 3rem;
  font-size: 1.125rem;
}

.article {
  margin-block: 5rem;
  width: 100%;
  @include flex-column(space-around, center, 1rem);
}

.article__image {
  max-width: 800px;
  width: 100%;
}

.article__title {
  font-size: clamp(2rem, 4vw, 4rem);
  font-weight: 500;
  text-align: center;
}

.article__content {
  padding-inline: 10%;
}
</style>
