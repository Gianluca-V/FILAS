import { fileURLToPath, URL } from 'node:url';

import vue from '@vitejs/plugin-vue';
import { defineConfig } from 'vite';

// Sass @import parses backslashes as escape sequences (e.g. "\De" inside a
// Windows path could be read as a unicode escape), so the injected path
// MUST use forward slashes even on Windows — fileURLToPath alone returns a
// native (backslash) path there.
function toPosixPath(fileUrl) {
  return fileURLToPath(fileUrl).replace(/\\/g, '/');
}

const variablesPath = toPosixPath(new URL('./src/styles/_variables.scss', import.meta.url));
const mixinsPath = toPosixPath(new URL('./src/styles/_mixins.scss', import.meta.url));

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        // Auto-inject shared design tokens/mixins into every SFC's
        // <style lang="scss"> block, so component styles never need to
        // hand-write these @import lines themselves.
        additionalData: `@import "${variablesPath}"; @import "${mixinsPath}";`,
      },
    },
  },
  server: {
    proxy: {
      // The Go API dev target (see backend/cmd/api). Demo/prod uses
      // nginx's /api reverse proxy instead (see design §4).
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: [],
  },
});
