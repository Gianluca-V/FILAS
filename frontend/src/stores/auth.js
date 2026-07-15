// Pinia auth store, replacing the legacy per-call getCookie('token') read
// (js/API.mjs) with reactive, centralized session state (design §3.3).
import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

import { login as loginRequest } from '../api/resources.js';

const TOKEN_STORAGE_KEY = 'filas_auth_token';

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem(TOKEN_STORAGE_KEY));

  const isAuthenticated = computed(() => Boolean(token.value));

  function setToken(newToken) {
    token.value = newToken;
    if (newToken) {
      localStorage.setItem(TOKEN_STORAGE_KEY, newToken);
    } else {
      localStorage.removeItem(TOKEN_STORAGE_KEY);
    }
  }

  async function login(username, password) {
    const { token: newToken } = await loginRequest({ username, password });
    setToken(newToken);
    return newToken;
  }

  function logout() {
    setToken(null);
  }

  return { token, isAuthenticated, login, logout };
});
