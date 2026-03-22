<script setup lang="ts">
import { RouterLink, useRoute } from 'vue-router'
import { Wifi } from 'lucide-vue-next'
import {
  BarChart3,
  ChevronsLeft,
  ChevronsRight,
  Globe,
  Home,
  MapPin,
  Package,
  Activity,
  Users,
} from 'lucide-vue-next'
import LiveBadge from '@/components/LiveBadge.vue'
import ThemeToggle from '@/components/ThemeToggle.vue'
import { cycleColorMode } from '@/lib/theme'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  useSidebar,
} from '@/components/ui/sidebar'

const route = useRoute()
const { state, toggleSidebar } = useSidebar()

const items = [
  { title: 'Home', to: '/', icon: Home },
  { title: 'Countries', to: '/countries', icon: Globe },
  { title: 'GeoKrety', to: '/geokrety', icon: Package },
  { title: 'Leaderboard', to: '/leaderboard', icon: BarChart3 },
  { title: 'Recent Moves', to: '/recent-moves', icon: Activity },
  { title: 'Users', to: '/users', icon: Users },
]

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>

<template>
  <Sidebar collapsible="icon" variant="inset">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton as-child size="lg" :is-active="route.path === '/'">
            <RouterLink
              to="/"
              :class="
                state === 'expanded' ? 'font-semibold' : 'flex items-center justify-center p-2'
              "
            >
              <div
                class="flex size-8 items-center justify-center rounded-md bg-primary text-primary-foreground"
              >
                <MapPin class="size-4" />
              </div>
              <div v-if="state === 'expanded'" class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-semibold">GeoKrety Stats</span>
                <span class="truncate text-xs text-muted-foreground">Dashboard</span>
              </div>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>

    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>Navigation</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in items" :key="item.to">
              <SidebarMenuButton as-child :is-active="isActive(item.to)">
                <RouterLink :to="item.to">
                  <component :is="item.icon" class="size-4" />
                  <span>{{ item.title }}</span>
                </RouterLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>

    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton @click="cycleColorMode">
            <ThemeToggle :asMenuButton="true" />
          </SidebarMenuButton>
        </SidebarMenuItem>

        <SidebarMenuItem>
          <SidebarMenuButton>
            <Wifi class="size-4" />
            <span>Status</span>
            <div class="ml-auto">
              <LiveBadge :show-text="state === 'expanded'" />
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>

        <SidebarMenuItem>
          <SidebarMenuButton @click="toggleSidebar">
            <ChevronsLeft v-if="state === 'expanded'" class="size-4" />
            <ChevronsRight v-else class="size-4" />
            <span v-if="state === 'expanded'">Collapse</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>

    <SidebarRail />
  </Sidebar>
</template>
