<script setup>
// Real admin login form, replacing the task 7.1 placeholder. Ports
// login.html's markup/copy onto the Pinia auth store (stores/auth.js):
// on success the store holds the JWT and this view redirects into the
// admin shell; on failure (401 -> ApiError, see api/client.js) it shows
// the same Spanish message legacy showed for a bad-credentials response.
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { useAuthStore } from '../../stores/auth.js';

const auth = useAuthStore();
const router = useRouter();

const username = ref('');
const password = ref('');
const errorMessage = ref('');
const isSubmitting = ref(false);

async function handleSubmit() {
  errorMessage.value = '';
  isSubmitting.value = true;

  try {
    await auth.login(username.value, password.value);
    router.push('/admin/products');
  } catch {
    errorMessage.value = 'Autentificación fallida. Intente de nuevo.';
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="admin-login">
    <h1 class="admin-login__title">Ingresar</h1>

    <form class="admin-login__form" @submit.prevent="handleSubmit">
      <label class="admin-login__label" for="username">Usuario:</label>
      <input id="username" v-model="username" class="admin-login__input" type="text" name="username" required>

      <label class="admin-login__label" for="password">Contraseña:</label>
      <input id="password" v-model="password" class="admin-login__input" type="password" name="password" required>

      <input
        class="admin-login__submit"
        type="submit"
        :value="isSubmitting ? 'Ingresando...' : 'Ingresar'"
        :disabled="isSubmitting"
      >
    </form>

    <p class="admin-login__message">{{ errorMessage }}</p>
  </div>
</template>

<style scoped lang="scss">
.admin-login {
  background-color: #f4f4f4;
  @include flex-column(center, center);
  min-height: 100svh;
}

.admin-login__title {
  text-align: center;
  margin-bottom: 20px;
}

.admin-login__form {
  border: 1px solid #ccc;
  border-radius: 5px;
  padding: 20px;
  box-shadow: 0px 0px 5px rgba(0, 0, 0, 0.1);
  width: 90%;
  max-width: 500px;
  @include flex-column(flex-start, stretch);
  background-color: $color-primary;
}

.admin-login__label {
  display: block;
  margin-bottom: 10px;
  font-weight: bold;
  color: $color-background;
}

.admin-login__input {
  margin-bottom: 10px;
  height: 35px;
  border-radius: 5px;
  background-color: $color-background;
  padding: 5px 10px;
}

.admin-login__submit {
  background-color: $color-text;
  color: $color-background;
  border: none;
  padding: 10px 20px;
  font-size: 16px;
  border-radius: 3px;
  cursor: pointer;

  &:hover {
    background-color: $color-accent;
  }
}

.admin-login__message {
  text-align: center;
  margin-top: 15px;
  color: $color-primary;
  font-weight: bold;
  min-height: 1.5em;
}
</style>
