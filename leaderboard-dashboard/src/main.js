import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router/index.js'

const app = createApp(App)
app.use(router)
app.mount('#app')

// Enable Bootstrap tooltips everywhere
// https://getbootstrap.com/docs/5.0/components/tooltips/#example-enable-tooltips-everywhere
router.afterEach(() => {
  // Re-initialize tooltips after each navigation
  setTimeout(() => {
    if (window.bootstrap?.Tooltip) {
      document.querySelectorAll('[data-bs-toggle="tooltip"], [title]:not([data-bs-toggle])').forEach((el) => {
        // Dispose existing tooltip if any
        const existing = window.bootstrap.Tooltip.getInstance(el)
        if (existing) existing.dispose()
        new window.bootstrap.Tooltip(el, { trigger: 'hover focus' })
      })
    }
  }, 150)
})
