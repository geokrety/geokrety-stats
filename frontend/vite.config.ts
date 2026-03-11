import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'

const usePolling =
  process.env.VITE_USE_POLLING === 'true' || process.env.CHOKIDAR_USEPOLLING === 'true'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vueDevTools(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    watch: {
      // Polling is expensive; keep it opt-in for environments where inotify is unavailable.
      usePolling,
      interval: usePolling ? 300 : undefined,
      ignored: ['**/.git/**', '**/.pnpm-store/**', '**/dist/**', '**/node_modules/**'],
    },
    hmr: {
      // ← explicitly declare the HMR host so browsers pick it up correctly
      host: 'localhost',
    },
    fs: {
      // allow serving files from outside the workspace root (e.g. node_modules)
      strict: false,
    },
  },
  optimizeDeps: {
    // Pre-bundle CJS/large ESM packages so Vite serves them with the correct
    // MIME type instead of returning an empty response for the first request.
    include: ['leaflet', 'topojson-client'],
  },
})
