// vue-router (history mode), replacing the legacy hand-rolled hashchange
// router in js/adminRouter.mjs (design §3.2). Public views (task 7.2) and
// admin views (task 8.1) are both real here; only the auth guard's
// contract predates the views it protects.
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
  {
    path: '/admin',
    component: () => import('../views/admin/AdminLayout.vue'),
    children: [
      { path: '', redirect: { name: 'admin-products' } },
      {
        path: 'products',
        name: 'admin-products',
        component: () => import('../views/admin/AdminProductsView.vue'),
      },
      {
        path: 'gallery',
        name: 'admin-gallery',
        component: () => import('../views/admin/AdminGalleryView.vue'),
      },
      {
        path: 'families',
        name: 'admin-families',
        component: () => import('../views/admin/AdminFamiliesView.vue'),
      },
      {
        path: 'organizations',
        name: 'admin-organizations',
        component: () => import('../views/admin/AdminOrganizationsView.vue'),
      },
      {
        path: 'news',
        name: 'admin-news',
        component: () => import('../views/admin/AdminNewsView.vue'),
      },
      {
        path: 'orders',
        name: 'admin-orders',
        component: () => import('../views/admin/AdminOrdersView.vue'),
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// Route guard: every /admin/* route except /admin/login requires an
// authenticated session (design §3.3).
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
