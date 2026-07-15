<script setup>
import { onMounted, ref } from 'vue';

import { getGallery } from '../../api/resources.js';
import { resolveAssetPath } from '../../utils/assets.js';
import { isNotFound } from '../../utils/httpErrors.js';

const status = ref('loading');
const images = ref([]);
const selectedImage = ref(null);

onMounted(async () => {
  try {
    images.value = await getGallery();
    status.value = 'ready';
  } catch (error) {
    if (isNotFound(error)) {
      status.value = 'empty';
    } else {
      status.value = 'error';
    }
  }
});

function openImage(image) {
  selectedImage.value = image;
}

function closeImage() {
  selectedImage.value = null;
}
</script>

<template>
  <section v-if="selectedImage" class="viewimg">
    <div class="viewimg__header">
      <button type="button" class="viewimg__close" @click="closeImage">CERRAR</button>
    </div>
    <img class="viewimg__photo" :src="resolveAssetPath(selectedImage.Image)" alt="">
  </section>

  <section class="gallery">
    <div class="gallery__header">
      <h2 class="gallery__title">GALERÍA</h2>
    </div>

    <p v-if="status === 'loading'" class="gallery__status">Cargando galería...</p>
    <p v-else-if="status === 'error'" class="gallery__status">
      No pudimos cargar la galería. Intentá nuevamente más tarde.
    </p>
    <p v-else-if="status === 'empty'" class="gallery__status">No se encontraron imágenes.</p>

    <div v-else class="gallery__photos">
      <img
        v-for="(image, index) in images"
        :key="image.ID"
        :src="resolveAssetPath(image.Image)"
        :alt="`Imagen ${index + 1}`"
        loading="lazy"
        draggable="false"
        class="gallery__photo"
        @click="openImage(image)"
      >
    </div>
  </section>
</template>

<style scoped lang="scss">
.viewimg {
  position: fixed;
  inset: 0;
  z-index: 1000;
  margin: auto;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.9);
  box-shadow: 0 0 6px black;
  @include flex-column(center, center);
}

.viewimg__header {
  height: 10%;
  width: 100%;
  background-color: $color-secondary;
  @include flex-row(center, center);
}

.viewimg__close {
  padding: 10px;
  background-color: $color-secondary;
  color: $color-background;
  font-size: 2rem;
  width: 100%;
  height: 100%;
}

.viewimg__photo {
  height: 90%;
  padding-block: 10px;
  max-width: 100%;
}

.gallery {
  width: 90%;
  margin: auto;
  padding: 10px;
}

.gallery__header {
  margin-bottom: 10px;
}

.gallery__title {
  font-weight: 800;
  font-size: 2.2rem;
  color: $color-primary;
}

.gallery__status {
  text-align: center;
  padding-block: 2rem;
  font-size: 1.125rem;
}

.gallery__photos {
  column-count: 3;
  column-gap: 10px;
}

.gallery__photo {
  width: 100%;
  height: auto;
  display: block;
  margin-bottom: 10px;
  border-radius: 5px;
  cursor: pointer;
}

@include below-tablet {
  .gallery {
    width: 100%;
  }

  .gallery__photos {
    column-count: 2;
  }
}

@media (max-width: 650px) {
  .gallery {
    width: 100%;
  }

  .gallery__photos {
    column-count: 1;
  }
}
</style>
