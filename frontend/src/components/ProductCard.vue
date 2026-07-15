<script setup>
import { resolveAssetPath } from '../utils/assets.js';

// Layout/classes mirror the legacy product__card markup (js/wildeArtesanal.mjs
// generateProductCard) for visual parity — no redesign.
defineProps({
  product: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(['add-to-cart']);
</script>

<template>
  <article class="product__card">
    <img
      :src="resolveAssetPath(product.Image)"
      :alt="product.Name"
      loading="lazy"
      class="product__img"
    >
    <div class="product__data">
      <h2 class="product__name">{{ product.Name }}</h2>
      <p class="product__price">${{ product.Price }}</p>
      <button type="button" class="product__buy" @click="emit('add-to-cart', product)">
        Añadir al carrito
      </button>
    </div>
  </article>
</template>

<style scoped lang="scss">
.product__card {
  text-wrap: balance;
  border: 1px solid $color-text;
  border-radius: 1rem;
  background-color: $color-primary;
  height: 25rem;
}

.product__img {
  height: 300px;
  border-radius: 1rem 1rem 0 0;
  width: 100%;
  object-fit: cover;
}

.product__data {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  grid-template-rows: clamp(0.7rem, 8vw, 2.5rem) 100%;
  align-items: center;
}

.product__name {
  grid-area: 1 / 1 / 1 / 3;
  line-height: clamp(1rem, 4.8vw, 1.5rem);
  font-size: clamp(1rem, 4.8vw, 1.5rem);
  color: $color-background;
  padding-inline: 0.5rem;
}

.product__price {
  grid-area: 2 / 1;
  font-size: clamp(1rem, 5vw, 1.3rem);
  font-weight: 600;
  color: $color-background;
  padding-left: 0.5rem;
}

.product__buy {
  grid-area: 2 / 2;
  margin: 5px;
  font-size: clamp(1rem, 5vw, 1.2rem);
  border-radius: 2rem;
  cursor: pointer;
  background-color: $color-accent;
  color: $color-background;
  transition: background-color 0.2s ease;

  &:hover {
    background-color: #0d6e0d;
  }
}

@media (width <= 425px) {
  .product__data {
    grid-template-rows: clamp(1.5rem, 6.5vw, 2.5rem) 100%;
  }
}
</style>
