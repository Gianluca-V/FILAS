<script setup>
// Reusable confirmation prompt, replacing the legacy admin panel's
// `window.confirm(...)` calls (e.g. js/admin/productosController.mjs)
// with a testable, styled component shared by every admin delete action.
defineProps({
  open: { type: Boolean, default: false },
  message: { type: String, required: true },
});

defineEmits(['confirm', 'cancel']);
</script>

<template>
  <div v-if="open" class="confirm-dialog-overlay">
    <div class="confirm-dialog" role="alertdialog">
      <p class="confirm-dialog__message">{{ message }}</p>
      <div class="confirm-dialog__actions">
        <button
          type="button"
          data-testid="confirm-cancel"
          class="confirm-dialog__button confirm-dialog__button--cancel"
          @click="$emit('cancel')"
        >
          Cancelar
        </button>
        <button
          type="button"
          data-testid="confirm-delete"
          class="confirm-dialog__button confirm-dialog__button--confirm"
          @click="$emit('confirm')"
        >
          Confirmar
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.confirm-dialog-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  @include flex-row(center, center);
  z-index: 1000;
}

.confirm-dialog {
  background-color: $color-background;
  border-radius: 0.5rem;
  padding: 1.5rem;
  max-width: 400px;
  width: 90%;
  box-shadow: 0 0 25px #222;
}

.confirm-dialog__message {
  margin-bottom: 1.5rem;
  text-align: center;
}

.confirm-dialog__actions {
  @include flex-row(center, center, 1rem);
}

.confirm-dialog__button {
  padding: 0.5rem 1.25rem;
  border-radius: 0.25rem;
  font-weight: 600;
  color: $color-background;
}

.confirm-dialog__button--cancel {
  background-color: #666;
}

.confirm-dialog__button--confirm {
  background-color: $color-primary;
}
</style>
