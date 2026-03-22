<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { RouterLink } from 'vue-router'
import { useUsers } from '@/composables/useUsers'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'
import AppBreadcrumb from '@/components/AppBreadcrumb.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Users, Search, Calendar } from 'lucide-vue-next'
import { relativeTime } from '@/lib/dates'

const { users, loading, error, hasMore, fetch, loadMore: fetchMore, search } = useUsers()
const searchQuery = ref('')
const isSearching = ref(false)

function handleLoadMore(): void {
  if (hasMore.value && !loading.value) {
    loadMoreUsers().then(() => nextTick(() => { if (hasMore.value) observe() }))
  }
}

const { sentinel, observe, disconnect } = useInfiniteScroll(handleLoadMore)

async function loadMoreUsers(): Promise<void> {
  await fetchMore(20)
}

function onSearch(): void {
  disconnect()
  if (searchQuery.value.trim().length >= 2) {
    isSearching.value = true
  } else {
    isSearching.value = false
  }
  search(searchQuery.value.trim(), 20).then(() => nextTick(() => { if (hasMore.value) observe() }))
}

function clearSearch(): void {
  searchQuery.value = ''
  isSearching.value = false
  disconnect()
  fetch(20).then(() => nextTick(() => { if (hasMore.value) observe() }))
}

// Debounced search on input
let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(searchQuery, (val) => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    if (val.trim().length >= 2 || val.trim().length === 0) {
      onSearch()
    }
  }, 300)
})

// Initial load
fetch(20).then(() => nextTick(() => { if (hasMore.value) observe() }))
</script>

<template>
  <main class="min-h-screen bg-background text-foreground pb-16">
    <div class="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 pt-10">
      <AppBreadcrumb :items="[{ label: 'Users' }]" />

      <div class="flex items-center gap-3 mb-6">
        <Users class="h-7 w-7 text-primary" />
        <h1 class="text-3xl font-bold tracking-tight">Users</h1>
      </div>

      <!-- Search bar -->
      <div class="flex gap-2 mb-6">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Search users…"
            class="pl-9"
          />
        </div>
        <Button v-if="isSearching" variant="ghost" size="sm" @click="clearSearch">Clear</Button>
      </div>

      <!-- Error state -->
      <div v-if="error" class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 mb-6">
        <p class="text-sm text-destructive">{{ error }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetch(20)">Retry</Button>
      </div>

      <!-- Users list -->
      <div class="space-y-2">
        <Card v-for="u in users" :key="u.id">
          <CardContent class="p-4 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="h-9 w-9 rounded-full bg-muted flex items-center justify-center text-sm font-semibold text-muted-foreground">
                {{ u.username.substring(0, 2).toUpperCase() }}
              </div>
              <div>
                <RouterLink :to="{ name: 'user-profile', params: { id: u.id } }" class="text-sm font-medium text-primary hover:underline">
                  {{ u.username }}
                </RouterLink>
                <div class="text-xs text-muted-foreground flex items-center gap-1">
                  <Calendar class="h-3 w-3" />
                  Joined {{ relativeTime(u.joinedAt) }}
                  <span v-if="u.homeCountry"> · {{ u.homeCountry }}</span>
                </div>
              </div>
            </div>
            <div v-if="u.lastMoveAt" class="text-xs text-muted-foreground">
              Last active {{ relativeTime(u.lastMoveAt) }}
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Empty state -->
      <div v-if="!loading && users.length === 0 && !error" class="text-center py-16">
        <Users class="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
        <p class="text-muted-foreground">{{ isSearching ? 'No users match your search.' : 'No users found.' }}</p>
      </div>

      <!-- Loading indicator -->
      <div v-if="loading" class="flex justify-center py-6">
        <div class="h-6 w-6 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
      </div>

      <!-- Infinite scroll sentinel -->
      <div ref="sentinel" class="h-1" />
    </div>
  </main>
</template>
