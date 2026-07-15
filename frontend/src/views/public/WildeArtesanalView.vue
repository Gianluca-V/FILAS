<script setup>
import { onMounted, ref } from 'vue';

import { getProducts } from '../../api/resources.js';
import CartPanel from '../../components/CartPanel.vue';
import ProductCard from '../../components/ProductCard.vue';
import { useCartStore } from '../../stores/cart.js';

// GET /api/products is never 404-on-empty (unlike news/gallery/family/
// organizations) — an empty catalog is a genuine 200 [] (see backend
// handler/rest/product.go doc comment), so no isNotFound special-casing
// is needed here.
const status = ref('loading');
const products = ref([]);
const cart = useCartStore();

onMounted(async () => {
  try {
    products.value = await getProducts();
    status.value = 'ready';
  } catch {
    status.value = 'error';
  }
});

function addToCart(product) {
  cart.addItem({ productID: Number(product.ID), name: product.Name, price: product.Price });
}
</script>

<template>
  <img src="/assets/WildeArtesanalBanner.jpg" alt="Banner Wilde Artesanal" class="title" draggable="false">
  <p class="description">
    Wilde Artesanal es un emprendimiento gastronómico desarrollado los operarios del Taller Protegido de Producción
    de la Fundación F.I.L.A.S. Entre nuestros productos podrán encontrar alimentos de panadería, pastelería, dulces
    envasados y encurtidos. Todos ellos elaborados artesanalmente, respetando las normas de seguridad e higiene
    establecidas por el Ministerio de Desarrollo Agrario y Desarrollo Productivo de la Provincia de Buenos Aires,
    necesarias para comerciar en todo el territorio. Los invitamos a elegir nuestros productos y comunicarse por el
    formulario o Whatsapp para completar su pedido.
  </p>

  <p v-if="status === 'loading'" class="products-status">Cargando productos...</p>
  <p v-else-if="status === 'error'" class="products-status">
    No pudimos cargar los productos. Intentá nuevamente más tarde.
  </p>
  <p v-else-if="products.length === 0" class="products-status">No hay productos disponibles por el momento.</p>

  <section v-else class="products-section">
    <ProductCard
      v-for="product in products"
      :key="product.ID"
      :product="product"
      @add-to-cart="addToCart"
    />
  </section>

  <div class="separator" />

  <CartPanel />
</template>

<style scoped lang="scss">
.title {
  text-align: center;
  display: block;
  width: 100%;
  max-height: 300px;
  object-fit: cover;
}

.description {
  text-align: center;
  font-size: 1.2rem;
  padding: 24px;
  margin: auto;
  font-weight: 600;
}

.products-status {
  text-align: center;
  padding-block: 2rem;
  font-size: 1.125rem;
}

.products-section {
  padding: 3rem;
  text-align: center;
  display: grid;
  grid-template-columns: repeat(auto-fit, min(20rem, 80vw));
  gap: 2rem;
  justify-content: center;
  margin: auto;
}

.separator {
  width: 90%;
  height: 2px;
  background-color: $color-text;
  margin: auto;
}
</style>
