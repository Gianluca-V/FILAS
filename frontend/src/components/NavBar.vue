<script setup>
import { ref } from 'vue';

// Nav structure/labels/brand mirror the legacy header markup (index.html)
// so the Vue rebuild keeps visual parity (design §3 — no redesign).
const isMenuOpen = ref(false);

function toggleMenu() {
  isMenuOpen.value = !isMenuOpen.value;
}
</script>

<template>
  <header class="header">
    <router-link to="/"><img class="logo" src="/assets/logo.png" alt="Fundación FILAS"></router-link>
    <nav class="nav">
      <ul class="nav__ul" :class="{ active: isMenuOpen }">
        <li class="nav__li">
          <router-link to="/" class="nav__a" @click="isMenuOpen = false">Inicio</router-link>
        </li>
        <li class="nav__li">
          <router-link to="/familias" class="nav__a" @click="isMenuOpen = false">Familia</router-link>
        </li>
        <li class="nav__li">
          <router-link to="/organizaciones" class="nav__a" @click="isMenuOpen = false">
            <span>Instituciones</span> <span>aliadas</span>
          </router-link>
        </li>
        <li class="nav__li">
          <router-link to="/wilde-artesanal" class="nav__a" @click="isMenuOpen = false">Productos</router-link>
        </li>
        <li class="nav__li">
          <router-link to="/noticias" class="nav__a" @click="isMenuOpen = false">Noticias</router-link>
        </li>
        <li class="nav__li">
          <router-link to="/galeria" class="nav__a" @click="isMenuOpen = false">Galería</router-link>
        </li>
      </ul>
      <button type="button" class="toggle-menu" aria-label="Abrir menú" @click="toggleMenu">
        <svg xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 448 512">
          <path
            d="M0 96C0 78.3 14.3 64 32 64H416c17.7 0 32 14.3 32 32s-14.3 32-32 32H32C14.3 128 0 113.7 0 96zM0 256c0-17.7 14.3-32 32-32H416c17.7 0 32 14.3 32 32s-14.3 32-32 32H32c-17.7 0-32-14.3-32-32zM448 416c0 17.7-14.3 32-32 32H32c-17.7 0-32-14.3-32-32s14.3-32 32-32H416c17.7 0 32 14.3 32 32z" />
        </svg>
      </button>
    </nav>
  </header>
</template>

<style scoped lang="scss">
.header {
  @include flex-row(space-between, center);
  width: 100%;
  background-color: $color-background;
}

.logo {
  max-width: 250px;
  padding: 10px;
}

.nav {
  position: relative;
}

.nav__ul {
  @include flex-row(space-around, center, 1.25rem);
  padding: 30px;
  list-style: none;
}

.nav__li {
  display: block;
  font-size: 1.125rem;
}

.nav__a {
  color: $color-text;

  &:hover {
    text-decoration: underline;
  }
}

.toggle-menu {
  display: none;

  svg {
    fill: $color-text;
  }
}

@include below-tablet {
  .logo {
    max-width: 150px;
  }

  .nav__ul {
    display: none;
    gap: 0;
    position: absolute;
    top: 150%;
    right: 0;
    width: 10%;
    transition: all 0.5s ease-in-out;
    background-color: $color-background;
    border-left: 1px solid rgba(0, 0, 0, 0.2666666667);
    border-bottom: 1px solid rgba(0, 0, 0, 0.2666666667);
    padding-inline: 100px;
    border-radius: 0 0 0 10px;
    z-index: 999999;

    &.active {
      @include flex-column(center, center);
    }
  }

  .nav__li {
    margin-block: 2rem;
    color: $color-text;
  }

  .nav__a {
    @include flex-column(center, center);
  }

  .toggle-menu {
    display: block;
    padding-right: 0.5rem;
    font-size: 2rem;
  }
}
</style>
