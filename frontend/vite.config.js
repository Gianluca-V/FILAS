import { fileURLToPath, URL } from 'node:url';

import vue from '@vitejs/plugin-vue';
import { defineConfig } from 'vite';

const stylesDir = fileURLToPath(new URL('./src/styles', import.meta.url));

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
        // hand-write these @use lines themselves. `as *` keeps variables
        // ($color-primary, etc.) and mixins (flex-row, etc.) unprefixed,
        // matching how every component already references them.
        //
        // Sass's @use module resolution (unlike the legacy @import) does
        // not accept an absolute Windows drive-letter path directly — it
        // needs a relative module specifier resolved via `includePaths`.
        additionalData: '@use "variables" as *; @use "mixins" as *;',
        includePaths: [stylesDir],
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
