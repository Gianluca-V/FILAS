<script setup>
// Ports js/admin/ordenesController.mjs + its .html partial: list orders,
// view line items/client data inline, change state (PATCH finished/
// canceled), and replace line items (PUT). GET /api/orders is never
// 404-on-empty (see backend/internal/handler/rest/order.go), so no
// notFoundIsEmpty flag — unlike news/gallery/family/organizations.
//
// Orders are NOT created here (public checkout only, see
// views/public/WildeArtesanalView.vue). The line-item editor below lets
// an admin pick REPLACEMENT products/quantities for an order (matching
// PUT /api/orders/:id's ReplaceProducts semantics — it swaps the whole
// line-item set, it does not patch individual lines) rather than editing
// existing lines in place: dto.OrderProductResponse deliberately omits
// each line's productID (see backend/internal/handler/rest/dto/response.go),
// so there is no stable id to edit against — only productName, which is
// used here to best-effort preselect a matching catalog product.
import { computed, ref } from 'vue';

import { getOrders, getProducts, patchOrderState, updateOrder } from '../../api/resources.js';
import { useAsyncResource } from '../../composables/useAsyncResource.js';

const { data: orders, status, refresh } = useAsyncResource(getOrders);

const STATE_LABELS = { pending: 'Pendiente', finished: 'Finalizado', canceled: 'Cancelada' };

const filterState = ref('all');
const filteredOrders = computed(() => {
  if (filterState.value === 'all') {
    return orders.value;
  }
  return orders.value.filter((order) => order.orderState === filterState.value);
});

const actionErrors = ref({});

// A non-pending order can't be re-patched (409 "order X is not pending" —
// see backend/internal/usecase/order.go); surfaced here as the same
// Spanish message legacy showed for that response (js/admin/
// ordenesController.mjs).
async function transitionState(order, target) {
  actionErrors.value = { ...actionErrors.value, [order.orderID]: '' };
  try {
    await patchOrderState(order.orderID, target);
    await refresh();
  } catch (error) {
    const message =
      error?.status === 409
        ? 'No se pudo modificar la orden, es posible que ya esté finalizada/cancelada.'
        : 'No se pudo actualizar el estado de la orden.';
    actionErrors.value = { ...actionErrors.value, [order.orderID]: message };
  }
}

const editingOrderId = ref(null);
const editableItems = ref([]);
const productCatalog = ref([]);
const itemsError = ref('');

function findProductIdByName(name) {
  const match = productCatalog.value.find((product) => product.Name === name);
  return match ? match.ID : '';
}

async function openItemsEditor(order) {
  editingOrderId.value = order.orderID;
  itemsError.value = '';
  productCatalog.value = await getProducts().catch(() => []);
  editableItems.value = order.products.map((item) => ({
    productID: findProductIdByName(item.productName),
    productName: item.productName,
    quantity: item.productQuantity,
  }));
}

function closeItemsEditor() {
  editingOrderId.value = null;
}

function addItemRow() {
  editableItems.value.push({ productID: '', productName: '', quantity: 1 });
}

function removeItemRow(index) {
  editableItems.value.splice(index, 1);
}

// Defense in depth: the editor is a real <form> now (see template) so the
// product <select>'s `required` participates in the browser's own
// validation on submit, but we don't rely on that alone — a synthetic
// form submission (e.g. in tests, or a browser that skips constraint
// validation) would otherwise let a falsy/0/NaN productID reach the API.
//
// Residual limitation: dto.OrderProductResponse exposes no per-line
// productID (see backend/internal/handler/rest/dto/response.go), so
// findProductIdByName's name→id preselect is best-effort and an admin
// must confirm each product before saving. This guard only prevents an
// empty/0 id from being submitted — it cannot detect a duplicate product
// NAME silently preselecting the wrong id, which would need a backend
// productID to fully fix.
function hasInvalidItemRow(item) {
  const productId = Number(item.productID);
  const quantity = Number(item.quantity);
  return !item.productID || !productId || !Number.isInteger(quantity) || quantity <= 0;
}

async function submitItems() {
  itemsError.value = '';

  if (editableItems.value.length === 0 || editableItems.value.some(hasInvalidItemRow)) {
    itemsError.value = 'Seleccioná un producto válido y una cantidad mayor a 0 en cada línea antes de guardar.';
    return;
  }

  try {
    await updateOrder(editingOrderId.value, {
      orderProducts: editableItems.value.map((item) => ({
        productID: Number(item.productID),
        quantity: Number(item.quantity),
      })),
    });
    editingOrderId.value = null;
    await refresh();
  } catch (error) {
    itemsError.value = error.body?.message || 'No se pudieron actualizar los productos de la orden.';
  }
}
</script>

<template>
  <section class="admin-resource">
    <div class="admin-resource__header">
      <h3 class="admin-resource__title">Ordenes</h3>
    </div>

    <div class="admin-resource__filter">
      <label for="order-filter">Mostrar: </label>
      <select id="order-filter" v-model="filterState">
        <option value="all">Todas</option>
        <option value="pending">Pendientes</option>
        <option value="finished">Finalizadas</option>
        <option value="canceled">Canceladas</option>
      </select>
    </div>

    <p v-if="status === 'loading'" class="admin-resource__status">Cargando ordenes...</p>
    <p v-else-if="status === 'error'" class="admin-resource__status">
      No pudimos cargar las ordenes. Intentá nuevamente más tarde.
    </p>
    <p v-else-if="filteredOrders.length === 0" class="admin-resource__status">No hay ordenes para mostrar.</p>

    <table v-else class="admin-table">
      <thead>
        <tr>
          <th>#</th>
          <th>Productos</th>
          <th>Total</th>
          <th>Estado</th>
          <th>Cliente</th>
          <th>Fecha de Inicio</th>
          <th>Fecha de Finalización</th>
          <th>Acciones</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="order in filteredOrders" :key="order.orderID">
          <td>{{ order.orderID }}</td>
          <td>
            <!-- productPrice is the product's CURRENT catalog price (live JOIN,
                 see the doc comment on OrderProductResponse in
                 backend/internal/handler/rest/dto/response.go), NOT the
                 price billed at order time. Do not multiply it by quantity
                 and present it as a "subtotal" — it can disagree with the
                 frozen order.orderTotal shown in the Total column. Label it
                 explicitly as the current price instead. -->
            <ul class="admin-order__products">
              <li v-for="(product, index) in order.products" :key="index">
                {{ product.productName }} x{{ product.productQuantity }}
                (precio actual: $ {{ product.productPrice }} c/u)
              </li>
            </ul>
          </td>
          <td>$ {{ order.orderTotal }}</td>
          <td>{{ STATE_LABELS[order.orderState] ?? order.orderState }}</td>
          <td>{{ order.orderName }} — {{ order.orderPhone }}</td>
          <td>{{ order.orderStartDate }}</td>
          <td>{{ order.orderFinishDate == null ? 'Pendiente' : order.orderFinishDate }}</td>
          <td class="admin-table__actions">
            <button type="button" :data-testid="`finish-${order.orderID}`" @click="transitionState(order, 'finished')">
              Finalizar
            </button>
            <button type="button" :data-testid="`cancel-${order.orderID}`" @click="transitionState(order, 'canceled')">
              Cancelar
            </button>
            <button
              type="button"
              :data-testid="`edit-items-${order.orderID}`"
              @click="openItemsEditor(order)"
            >
              Editar productos
            </button>
            <p v-if="actionErrors[order.orderID]" class="admin-order__error">{{ actionErrors[order.orderID] }}</p>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="editingOrderId !== null" class="order-items-editor-overlay">
      <div class="order-items-editor">
        <h4>Editar productos de la orden #{{ editingOrderId }}</h4>
        <p v-if="itemsError" class="admin-order__error">{{ itemsError }}</p>

        <!-- Real <form> (not a <div>) so the product <select>'s `required`
             participates in the browser's constraint validation on submit;
             see the hasInvalidItemRow guard in submitItems for the
             defense-in-depth check that doesn't rely on that alone. -->
        <form @submit.prevent="submitItems">
          <div v-for="(item, index) in editableItems" :key="index" class="order-items-editor__row">
            <span class="order-items-editor__original">{{ item.productName }}</span>
            <select v-model="item.productID" required>
              <option value="" disabled>Seleccionar producto</option>
              <option v-for="product in productCatalog" :key="product.ID" :value="product.ID">
                {{ product.Name }}
              </option>
            </select>
            <input v-model.number="item.quantity" type="number" min="1">
            <button type="button" @click="removeItemRow(index)">Quitar</button>
          </div>

          <button type="button" class="order-items-editor__add" @click="addItemRow">+ Agregar producto</button>

          <div class="order-items-editor__actions">
            <button type="submit" data-testid="save-items">Guardar</button>
            <button type="button" @click="closeItemsEditor">Cancelar</button>
          </div>
        </form>
      </div>
    </div>
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

.admin-resource__filter {
  @include flex-row(flex-end, center, 1rem);
  padding: 1rem;
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
    vertical-align: top;
  }
}

.admin-order__products {
  list-style: none;
  text-align: left;
}

.admin-order__error {
  color: $color-primary;
  font-weight: 600;
  margin-top: 0.5rem;
}

.admin-table__actions {
  @include flex-column(center, stretch, 0.5rem);
}

.order-items-editor-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  @include flex-row(center, center);
  z-index: 1000;
}

.order-items-editor {
  background-color: $color-background;
  border-radius: 0.5rem;
  padding: 1.5rem;
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.order-items-editor__row {
  @include flex-row(space-between, center, 0.5rem);
  margin-block: 0.5rem;
}

.order-items-editor__actions {
  @include flex-row(flex-end, center, 1rem);
  margin-top: 1rem;
}
</style>
