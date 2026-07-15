<script setup>
// Generic create/edit form driven by a field-schema prop, replacing the
// per-resource hand-rolled `<dialog>` forms in js/admin/*Controller.html
// (productosController/galeriaController/familiaController/
// organizacionesController/noticiasController all shared the same
// header+group+submit structure, only the field list differed).
//
// Each field: { key, label, type: 'text' | 'number' | 'textarea' | 'select',
// required, options (for 'select') }.
const props = defineProps({
  fields: { type: Array, required: true },
  modelValue: { type: Object, required: true },
  title: { type: String, default: '' },
  error: { type: String, default: '' },
  submitLabel: { type: String, default: 'Confirmar' },
});

const emit = defineEmits(['update:modelValue', 'submit', 'cancel']);

function updateField(key, value) {
  emit('update:modelValue', { ...props.modelValue, [key]: value });
}

function handleSubmit() {
  emit('submit', props.modelValue);
}
</script>

<template>
  <div class="resource-form-overlay">
    <form class="resource-form" @submit.prevent="handleSubmit">
      <div class="resource-form__header">
        <h3 class="resource-form__title">{{ title }}</h3>
        <button type="button" class="resource-form__close" @click="$emit('cancel')">×</button>
      </div>

      <p v-if="error" class="resource-form__error">{{ error }}</p>

      <div v-for="field in fields" :key="field.key" class="resource-form__group">
        <label :for="field.key" class="resource-form__label">{{ field.label }}:</label>

        <select
          v-if="field.type === 'select'"
          :id="field.key"
          class="resource-form__input"
          :required="field.required"
          :value="modelValue[field.key]"
          @change="updateField(field.key, $event.target.value)"
        >
          <option v-for="option in field.options" :key="option" :value="option">{{ option }}</option>
        </select>

        <textarea
          v-else-if="field.type === 'textarea'"
          :id="field.key"
          class="resource-form__input"
          rows="6"
          :required="field.required"
          :value="modelValue[field.key]"
          @input="updateField(field.key, $event.target.value)"
        />

        <input
          v-else
          :id="field.key"
          class="resource-form__input"
          :type="field.type"
          :required="field.required"
          :value="modelValue[field.key]"
          @input="updateField(field.key, $event.target.value)"
        >
      </div>

      <input type="submit" class="resource-form__input resource-form__input--submit" :value="submitLabel">
    </form>
  </div>
</template>

<style scoped lang="scss">
.resource-form-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1000;
}

.resource-form {
  border: black solid 1px;
  width: 90%;
  max-width: 500px;
  padding: 20px;
  position: fixed;
  background-color: $color-background;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  border-radius: 0.5rem;
  @include flex-column(flex-start, stretch);
  max-height: 90vh;
  overflow-y: auto;
}

.resource-form__header {
  @include flex-row(space-between, center);
}

.resource-form__close {
  font-size: 2rem;
  line-height: 1;
}

.resource-form__error {
  color: $color-primary;
  font-weight: 600;
  margin-block: 0.5rem;
}

.resource-form__group {
  display: flex;
  flex-direction: column;
  margin-block: 0.5rem;
}

.resource-form__label {
  font-weight: bold;
  margin-bottom: 0.25rem;
}

.resource-form__input:not(.resource-form__input--submit) {
  border-bottom: 1px $color-text solid;
  padding-block: 5px;

  &:focus {
    border-bottom: 1px $color-primary solid;
  }
}

.resource-form__input--submit {
  margin-top: 1rem;
  text-align: center;
  font-size: 1.2rem;
  font-weight: 600;
  padding: 10px;
  cursor: pointer;

  &:hover {
    background-color: black;
    color: white;
  }
}
</style>
