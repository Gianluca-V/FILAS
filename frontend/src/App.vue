<script setup>
import FooterBar from './components/FooterBar.vue';
import NavBar from './components/NavBar.vue';
</script>

<template>
  <NavBar />
  <main class="app__content">
    <router-view />
  </main>
  <FooterBar />
</template>

<style scoped lang="scss">
.app__content {
  flex: 1;
}
</style>
