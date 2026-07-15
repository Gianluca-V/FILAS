<script setup>
// Admin shell: sidebar nav + a <router-view/> slot for the nested
// per-resource views (router/index.js). Replaces legacy admin.html's
// static shell + js/adminRouter.mjs's hand-rolled hashchange/fetch-partial
// router (design §3.2) — vue-router nested routes + component lifecycle
// take over content swapping.
import AdminNav from '../../components/AdminNav.vue';
</script>

<template>
  <div class="admin-layout">
    <AdminNav />
    <main class="admin-layout__content">
      <router-view />
    </main>
  </div>
</template>

<style scoped lang="scss">
.admin-layout {
  @include flex-row(flex-start, stretch);
  min-height: 100svh;
  width: 100%;
}

.admin-layout__content {
  width: 100%;
  padding: 1rem;
}
</style>
