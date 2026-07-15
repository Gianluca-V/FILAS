<script setup>
import { ref } from 'vue';

import { useCartStore } from '../stores/cart.js';

const cart = useCartStore();

const name = ref('');
const phone = ref('');
const status = ref('idle'); // idle | submitting | success | error
const errorMessage = ref('');

function onQuantityInput(productID, event) {
  cart.setQuantity(productID, Number(event.target.value));
}

async function onSubmit() {
  status.value = 'submitting';
  errorMessage.value = '';
  try {
    await cart.checkout({ name: name.value, phone: phone.value });
    status.value = 'success';
    name.value = '';
    phone.value = '';
  } catch (error) {
    status.value = 'error';
    errorMessage.value = error.message || 'Ocurrió un problema con su compra, vuelva a intentarlo';
  }
}
</script>

<template>
  <form class="cart" @submit.prevent="onSubmit">
    <div class="cart__title-container">
      <svg class="cart__icon" xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 576 512">
        <path
          d="M0 24C0 10.7 10.7 0 24 0H69.5c22 0 41.5 12.8 50.6 32h411c26.3 0 45.5 25 38.6 50.4l-41 152.3c-8.5 31.4-37 53.3-69.5 53.3H170.7l5.4 28.5c2.2 11.3 12.1 19.5 23.6 19.5H488c13.3 0 24 10.7 24 24s-10.7 24-24 24H199.7c-34.6 0-64.3-24.6-70.7-58.5L77.4 54.5c-.7-3.8-4-6.5-7.9-6.5H24C10.7 48 0 37.3 0 24zM128 464a48 48 0 1 1 96 0 48 48 0 1 1 -96 0zm336-48a48 48 0 1 1 0 96 48 48 0 1 1 0-96z" />
      </svg>
      <h2 class="cart__title">Tu compra:</h2>
    </div>

    <div class="cart__products-container">
      <p v-if="!cart.hasItems">Agregue productos al carrito</p>
      <div v-for="item in cart.items" :key="item.productID" class="cart__product">
        <div class="cart__product-name">{{ item.name }}</div>
        <div class="cart__product-quantity-container">
          X
          <input
            type="number"
            class="cart__product-quantity"
            min="1"
            :value="item.quantity"
            @input="onQuantityInput(item.productID, $event)"
          >
        </div>
        <button
          type="button"
          class="cart__product-delete"
          aria-label="Quitar producto"
          @click="cart.removeItem(item.productID)"
        >
          <svg xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 448 512">
            <path
              d="M135.2 17.7L128 32H32C14.3 32 0 46.3 0 64S14.3 96 32 96H416c17.7 0 32-14.3 32-32s-14.3-32-32-32H320l-7.2-14.3C307.4 6.8 296.3 0 284.2 0H163.8c-12.1 0-23.2 6.8-28.6 17.7zM416 128H32L53.2 467c1.6 25.3 22.6 45 47.9 45H346.9c25.3 0 46.3-19.7 47.9-45L416 128z" />
          </svg>
        </button>
      </div>
    </div>

    <div class="cart__total-container">
      Total: $ <span class="cart__total">{{ cart.totalCost }}</span>
    </div>

    <div class="form__group">
      <label for="name">Nombre y apellido:</label>
      <input
        id="name"
        v-model="name"
        type="text"
        name="name"
        required
        placeholder="Ingrese su nombre y apellido"
      >
    </div>
    <div class="form__group">
      <label for="phone">Teléfono:</label>
      <input
        id="phone"
        v-model="phone"
        type="tel"
        name="phone"
        required
        placeholder="Ingrese su numero de teléfono"
      >
    </div>

    <p v-if="status === 'error'" class="cart__message cart__message--error">{{ errorMessage }}</p>
    <p v-if="status === 'success'" class="cart__message cart__message--success">
      Muchas gracias, su orden ha sido enviada
    </p>

    <button
      type="submit"
      class="cart__buy"
      :disabled="!cart.hasItems || status === 'submitting'"
    >
      Finalizar compra
    </button>
  </form>
</template>

<style scoped lang="scss">
.cart {
  max-width: 650px;
  margin: 4rem;
  margin-top: 5rem;
  margin-bottom: 2rem;
  margin-left: 4rem;
  @include flex-column(flex-start, flex-start, 1.5rem);
}

.cart__title-container {
  @include flex-row(flex-start, center, 1rem);
  font-size: 2rem;
}

.cart__icon {
  fill: $color-text;
}

.cart__title {
  font-size: clamp(1.666rem, 8vw, 2.255rem);
}

.cart__products-container {
  @include flex-column(flex-start, flex-start, 0.5rem);
  margin-left: 5rem;
  font-size: 1.3rem;
}

.cart__product {
  @include flex-row(flex-start, center);
  width: 100%;
}

.cart__product-name {
  margin-right: 1rem;
}

.cart__product-quantity-container {
  @include flex-row(flex-start, center, 1rem);
  margin-left: auto;
}

.cart__product-quantity {
  width: 100%;
  max-width: 100px;
  text-align: center;
  border: 1px solid #eee;
  font-size: 1.5rem;
}

.cart__product-delete {
  order: -1;
  font-size: 1.5rem;
  margin-right: 1.5rem;

  svg {
    cursor: pointer;
    fill: crimson;
  }

  &:hover svg {
    fill: red;
  }
}

.cart__total-container {
  margin-left: 5rem;
  font-size: 1.5rem;
  font-weight: 600;
}

.form__group {
  @include flex-column(flex-start, flex-start, 0.35rem);

  input {
    border: 1px solid rgba(0, 0, 0, 0.2);
    border-radius: 2px;
    padding: 5px;
    width: min(100%, 320px);
  }
}

.cart__message {
  font-weight: 600;
}

.cart__message--error {
  color: $color-primary;
}

.cart__message--success {
  color: $color-accent;
}

.cart__buy {
  font-size: clamp(1rem, 8vw, 1.66rem);
  width: 50%;
  text-align: center;
  border-radius: 1rem;
  padding: 0.25rem;
  background-color: $color-accent;
  color: $color-background;

  &:hover:not(:disabled) {
    background-color: #0d6e0d;
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }
}

@include below-mobile {
  .cart {
    margin-top: 5rem;
    margin-bottom: 2rem;
    margin-inline: 2rem;
  }

  .cart__products-container,
  .cart__total-container {
    margin-left: 0;
  }
}
</style>
