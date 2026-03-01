import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router/index.js'

const THEME_KEY = 'gk-dashboard-theme'

function detectInitialTheme() {
  const stored = localStorage.getItem(THEME_KEY)
  if (stored === 'light' || stored === 'dark') return stored
  return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

document.documentElement.setAttribute('data-bs-theme', detectInitialTheme())

const app = createApp(App)
app.use(router)
app.mount('#app')

// Enable Bootstrap tooltips everywhere
// https://getbootstrap.com/docs/5.0/components/tooltips/#example-enable-tooltips-everywhere
router.afterEach(() => {
  // Re-initialize tooltips after each navigation
  setTimeout(() => {
    document.querySelectorAll('button:not([type])').forEach((el) => {
      el.setAttribute('type', 'button')
    })

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
