import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',        // Listen on all interfaces
    port: 3000,             // Dev server on port 3000
    middlewareMode: false,
    hmr: {
      protocol: 'ws',
      host: '192.168.130.65', // Browser connects to this host for HMR
      port: 3000,            // HMR on same port as dev server
    },
    proxy: {
      '/api': {
        target: 'http://192.168.130.65:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://192.168.130.65:8080',
        ws: true,
        changeOrigin: true,
      },
    },
  },
})
