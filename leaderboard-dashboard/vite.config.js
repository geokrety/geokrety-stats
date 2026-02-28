import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
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
