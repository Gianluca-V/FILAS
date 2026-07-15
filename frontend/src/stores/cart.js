// Pinia cart store backing the WildeArtesanal checkout flow. Replaces the
// legacy `let cart = []` module-level array (js/wildeArtesanal.mjs) with
// reactive, centrally testable state, and moves the client-side "quantity
// must be a positive integer" gate here so it's unit-testable without
// mounting the DOM (Strict TDD — see stores/__tests__/cart.spec.js).
import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

import { createOrder } from '../api/resources.js';

export const useCartStore = defineStore('cart', () => {
  // { productID, name, price (string, as returned by GET /api/products —
  // Price is a JSON string, see backend dto.ProductResponse), quantity }
  const items = ref([]);

  const totalCost = computed(() =>
    items.value.reduce((sum, item) => sum + Number(item.price) * Number(item.quantity), 0),
  );

  const hasItems = computed(() => items.value.length > 0);

  function addItem(product) {
    const existing = items.value.find((item) => item.productID === product.productID);
    if (existing) {
      existing.quantity += 1;
      return;
    }
    items.value.push({ productID: product.productID, name: product.name, price: product.price, quantity: 1 });
  }

  function setQuantity(productID, quantity) {
    const item = items.value.find((item) => item.productID === productID);
    if (item) {
      item.quantity = quantity;
    }
  }

  function removeItem(productID) {
    items.value = items.value.filter((item) => item.productID !== productID);
  }

  function clear() {
    items.value = [];
  }

  // A line item is checkout-eligible only when its quantity is a positive
  // integer — mirrors the backend's tightened validation (POST /api/orders
  // rejects quantity <= 0; spec: "Public Checkout with Tightened
  // Validation"). Checked here so an invalid cart never reaches
  // createOrder().
  function isValidQuantity(quantity) {
    return Number.isInteger(quantity) && quantity > 0;
  }

  const hasInvalidQuantity = computed(() => items.value.some((item) => !isValidQuantity(item.quantity)));

  // Validates the cart + contact details, then posts the order and clears
  // the cart on success. Throws (never calls createOrder) when validation
  // fails, so callers can surface the message without an extra round trip.
  async function checkout({ name, phone }) {
    if (!hasItems.value) {
      throw new Error('El carrito está vacío');
    }
    if (!name || !name.trim() || !phone || !phone.trim()) {
      throw new Error('Ingrese su nombre y teléfono');
    }
    if (hasInvalidQuantity.value) {
      throw new Error('Las cantidades deben ser números enteros mayores a cero');
    }

    const payload = {
      name: name.trim(),
      phone: phone.trim(),
      orderProducts: items.value.map((item) => ({ productID: item.productID, quantity: item.quantity })),
    };

    await createOrder(payload);
    clear();
  }

  return {
    items,
    totalCost,
    hasItems,
    hasInvalidQuantity,
    addItem,
    setQuantity,
    removeItem,
    clear,
    checkout,
  };
});
