<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppSidebar from '@/components/AppSidebar.vue'
import TheNavbar from '@/components/TheNavbar.vue'
import TheFooter from '@/components/TheFooter.vue'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { provideTheme, colorMode, cycleColorMode, isDark } from '@/lib/theme'
provideTheme({ colorMode, cycleColorMode, isDark })
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset class="min-h-svh">
      <!-- Mobile + tablet top navbar (hidden on desktop ≥1024px) -->
      <TheNavbar class="lg:hidden" />

      <RouterView />
      <TheFooter />
    </SidebarInset>
  </SidebarProvider>
</template>
