<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { MapPin, Menu, X } from 'lucide-vue-next'
import LiveBadge from '@/components/LiveBadge.vue'
import ThemeToggle from '@/components/ThemeToggle.vue'

const mobileOpen = ref(false)

const navLinks = [
  { label: 'Home', to: '/' },
  { label: 'Countries', to: '/countries' },
]
</script>

<template>
  <nav class="sticky top-0 z-50 w-full border-b bg-background shadow-sm">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div class="flex h-16 items-center justify-between">
        <!-- Logo -->
        <RouterLink to="/" class="flex items-center gap-2 group flex-shrink-0">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground transition-transform group-hover:scale-110"
          >
            <MapPin class="h-5 w-5" />
          </div>
          <span class="text-lg font-bold text-foreground hidden sm:inline">
            GeoKrety Stats
          </span>
        </RouterLink>

        <div class="flex items-center gap-2">
          <LiveBadge />
          <ThemeToggle />
          <button
            class="rounded-lg p-2 text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
            @click="mobileOpen = !mobileOpen"
            :aria-label="mobileOpen ? 'Close menu' : 'Open menu'"
          >
            <X v-if="mobileOpen" class="h-6 w-6" />
            <Menu v-else class="h-6 w-6" />
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile menu -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div v-if="mobileOpen" class="border-t bg-background px-4 pb-4 pt-2">
        <RouterLink
          v-for="link in navLinks"
          :key="link.to"
          :to="link.to"
          class="block rounded-lg px-3 py-2.5 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
          active-class="bg-accent text-accent-foreground"
          @click="mobileOpen = false"
        >
          {{ link.label }}
        </RouterLink>
      </div>
    </Transition>
  </nav>
</template>
