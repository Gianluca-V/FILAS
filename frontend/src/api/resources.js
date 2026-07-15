// Per-resource CRUD calls built on api/client.js, matching the Go API
// contract in backend/internal/handler/rest/router.go and
// backend/internal/handler/rest/dto/response.go. Endpoints are
// /api/<resource> — no /FilasServer prefix (that was legacy-only).
// Task 7.2 (public views) and task 8.1 (admin views) consume these.
import { del, get, post, put } from './client.js';

// --- Products (public read) ---
export const getProducts = () => get('/products');
export const getProduct = (id) => get(`/products/${id}`);

// --- News (public read) ---
export const getNews = () => get('/news');
export const getNewsItem = (id) => get(`/news/${id}`);

// --- Gallery (public read) ---
export const getGallery = () => get('/gallery');

// --- Family (public read) ---
export const getFamily = () => get('/family');

// --- Organizations (public read) ---
export const getOrganizations = () => get('/organizations');

// --- Orders (public write: checkout) ---
// payload: { orderProducts: [{ productID, quantity }], name, phone }
export const createOrder = (payload) => post('/orders', payload);

// --- Auth (admin login) ---
// POST /api/admins overloads on the request body: {login:true,...}
// performs an unauthenticated login and returns {token}; anything else
// requires an existing session and creates an admin (not exposed here —
// admin CRUD lands in task 8.1).
export const login = ({ username, password }) => post('/admins', { login: true, username, password });

// --- Admin CRUD writes (auth-gated, wired to real UI in task 8.1) ---
export const createProduct = (payload) => post('/products', payload);
export const updateProduct = (id, payload) => put(`/products/${id}`, payload);
export const deleteProduct = (id) => del(`/products/${id}`);

export const createNewsItem = (payload) => post('/news', payload);
export const updateNewsItem = (id, payload) => put(`/news/${id}`, payload);
export const deleteNewsItem = (id) => del(`/news/${id}`);

export const createGalleryItem = (payload) => post('/gallery', payload);
export const updateGalleryItem = (id, payload) => put(`/gallery/${id}`, payload);
export const deleteGalleryItem = (id) => del(`/gallery/${id}`);

export const createFamilyItem = (payload) => post('/family', payload);
export const updateFamilyItem = (id, payload) => put(`/family/${id}`, payload);
export const deleteFamilyItem = (id) => del(`/family/${id}`);

export const createOrganization = (payload) => post('/organizations', payload);
export const updateOrganization = (id, payload) => put(`/organizations/${id}`, payload);
export const deleteOrganization = (id) => del(`/organizations/${id}`);
