<script setup>
import { getNews } from '../../api/resources.js';
import { useAsyncResource } from '../../composables/useAsyncResource.js';
import { resolveAssetPath } from '../../utils/assets.js';

// status: 'loading' | 'ready' | 'empty' | 'error'. Empty is a distinct
// state (not an error): GET /api/news returns 404 when there is no news
// yet (see backend/internal/handler/rest/news.go), so a 404 here means
// "no news", not a failure.
const { data: news, status } = useAsyncResource(getNews, { notFoundIsEmpty: true });
</script>

<template>
  <h1 class="news-title">Noticias</h1>

  <p v-if="status === 'loading'" class="news-status">Cargando noticias...</p>
  <p v-else-if="status === 'error'" class="news-status">
    No pudimos cargar las noticias. Intentá nuevamente más tarde.
  </p>
  <p v-else-if="status === 'empty'" class="news-status">Todavía no hay noticias publicadas.</p>

  <section v-else class="news">
    <div v-if="news[0]" class="news__new news__new--big">
      <img
        :src="resolveAssetPath(news[0].Image)"
        :alt="news[0].Title"
        loading="lazy"
        class="news__img news__img--big"
      >
      <div class="news__data news__data--big">
        <h2 class="news__title news__title--big">{{ news[0].Title }}</h2>
        <router-link :to="`/noticias/${news[0].ID}`" class="news__read news__read--big">Leer mas</router-link>
      </div>
    </div>

    <div class="news__container">
      <div v-for="article in news.slice(1)" :key="article.ID" class="news__new">
        <img :src="resolveAssetPath(article.Image)" :alt="article.Title" loading="lazy" class="news__img">
        <div class="news__data">
          <h2 class="news__title">{{ article.Title }}</h2>
          <router-link :to="`/noticias/${article.ID}`" class="news__read">Leer mas</router-link>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.news-title {
  font-size: 3rem;
  text-align: center;
  width: 100%;
  color: $color-background;
  background-color: $color-primary;
  margin-bottom: 16px;
}

.news-status {
  text-align: center;
  padding-block: 2rem;
  font-size: 1.125rem;
}

.news {
  max-width: 100%;
  @include flex-column(center, center);
}

.news__new {
  text-wrap: balance;
  text-align: center;
  @include flex-row(space-evenly, center);
  max-width: 100%;
  border: 1px solid rgba(68, 68, 68, 0.267);
  box-shadow: 5px 5px 5px rgba(68, 68, 68, 0.267);
  transition: box-shadow 0.25s ease;
  margin: 1.5rem;

  &:hover {
    box-shadow: 5px 5px 6px rgba(68, 68, 68, 0.667);
  }
}

.news__new--big {
  margin: auto;
  margin-bottom: 5rem;
  text-align: center;
  display: block;
  width: 90%;
  font-size: 2rem;
}

@include below-mobile {
  .news__new {
    flex-direction: column;
  }
}

.news__container {
  margin: 3rem;
  display: grid;
  grid-template-columns: repeat(2, 0.8fr);
  justify-content: center;
  gap: 3rem;
}

@media (width < 640px) {
  .news__container {
    grid-template-columns: repeat(1, 0.8fr);
  }
}

.news__img {
  max-width: 500px;
  width: 50%;
}

.news__img--big {
  width: 100%;
}

.news__title {
  line-height: 3rem;
}

.news__title--big {
  line-height: 5rem;
}

.news__data {
  display: flex;
  flex-direction: column;
}

.news__read {
  color: #44f;

  &:hover {
    text-decoration: underline;
  }
}
</style>
