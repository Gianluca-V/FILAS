// vue-router (history mode), replacing the legacy hand-rolled hashchange
// router in js/adminRouter.mjs (design §3.2). Public views are registered
// here as lazy-loaded placeholders; task 7.2 replaces the stub SFCs under
// views/public/ with the real implementations. Admin views land in task
// 8.1 — only the login placeholder and the auth guard are real here.
import { createRouter, createWebHistory } from 'vue-router';

import { useAuthStore } from '../stores/auth.js';

const routes = [
  {
    path: '/',
    name: 'home',
    component: () => import('../views/public/HomeView.vue'),
  },
  {
    path: '/noticias',
    name: 'news-list',
    component: () => import('../views/public/NewsListView.vue'),
  },
  {
    path: '/noticias/:id',
    name: 'news-detail',
    component: () => import('../views/public/NewsDetailView.vue'),
    props: true,
  },
  {
    path: '/galeria',
    name: 'gallery',
    component: () => import('../views/public/GalleryView.vue'),
  },
  {
    path: '/familias',
    name: 'families',
    component: () => import('../views/public/FamiliesView.vue'),
  },
  {
    path: '/organizaciones',
    name: 'organizations',
    component: () => import('../views/public/OrganizationsView.vue'),
  },
  {
    path: '/informacion',
    name: 'info',
    component: () => import('../views/public/InfoView.vue'),
  },
  {
    path: '/wilde-artesanal',
    name: 'wilde-artesanal',
    component: () => import('../views/public/WildeArtesanalView.vue'),
  },
  {
    path: '/admin/login',
    name: 'admin-login',
    component: () => import('../views/admin/AdminLoginView.vue'),
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// Route guard: every /admin/* route except /admin/login requires an
// authenticated session (design §3.3). Admin views themselves ship in
// task 8.1, but the guard logic is real and tested now so the auth
// contract is locked before those views exist.
router.beforeEach((to) => {
  const isAdminRoute = to.path.startsWith('/admin');
  const isLoginRoute = to.name === 'admin-login';

  if (!isAdminRoute || isLoginRoute) {
    return true;
  }

  const auth = useAuthStore();
  if (!auth.isAuthenticated) {
    return { name: 'admin-login' };
  }

  return true;
});

export default router;
