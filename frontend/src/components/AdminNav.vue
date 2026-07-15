<script setup>
// Sidebar nav for the admin shell, replacing legacy admin.html's static
// `<nav class="nav--admin">` + the inline "if no token cookie, redirect"
// script — the router guard (router/index.js) now owns that check, and
// logout is a real reactive action against the Pinia auth store.
import { useRouter } from 'vue-router';

import { useAuthStore } from '../stores/auth.js';

const auth = useAuthStore();
const router = useRouter();

const links = [
  { to: '/admin/products', label: 'Productos' },
  { to: '/admin/gallery', label: 'Galería' },
  { to: '/admin/families', label: 'Familia' },
  { to: '/admin/organizations', label: 'Organizaciones' },
  { to: '/admin/news', label: 'Noticias' },
  { to: '/admin/orders', label: 'Ordenes' },
];

function handleLogout() {
  auth.logout();
  router.push({ name: 'admin-login' });
}
</script>

<template>
  <nav class="admin-nav">
    <router-link to="/" class="admin-nav__logo-link">
      <img class="admin-nav__logo" src="/assets/logo.png" alt="Fundación FILAS">
    </router-link>

    <ul class="admin-nav__list">
      <li v-for="link in links" :key="link.to">
        <router-link :to="link.to" class="admin-nav__link">{{ link.label }}</router-link>
      </li>
    </ul>

    <button type="button" class="admin-nav__logout" @click="handleLogout">Cerrar sesión</button>
  </nav>
</template>

<style scoped lang="scss">
.admin-nav {
  max-width: 200px;
  min-width: 200px;
  min-height: 100%;
  background-color: $color-primary;
  @include flex-column(flex-start, center);
  padding-block: 1rem;
}

.admin-nav__logo-link {
  display: block;
  margin-bottom: 2rem;
}

.admin-nav__logo {
  max-width: 100%;
}

.admin-nav__list {
  list-style: none;
  width: 100%;
  @include flex-column(space-around, center, 1rem);
}

.admin-nav__link {
  display: block;
  width: 100%;
  text-align: center;
  color: $color-background;
  padding: 10px;
  font-size: clamp(1rem, 2vw, 1.2rem);
  transition: background-color 0.25s ease;

  &:hover {
    background-color: #f24646;
    text-decoration: underline;
  }
}

.admin-nav__logout {
  margin-top: 2rem;
  color: $color-background;
  border: 1px solid $color-background;
  border-radius: 0.25rem;
  padding: 0.5rem 1rem;

  &:hover {
    background-color: #f24646;
  }
}
</style>
