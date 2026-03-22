<script setup lang="ts">
import { onMounted, ref, watch, nextTick } from 'vue'
import { useGeokrety } from '@/composables/useGeokrety'
import { useGkid } from '@/composables/useGkid'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'
import GeokretyTypeBadge from '@/components/GeokretyTypeBadge.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Search, Package } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'

const { geokrety, loading, error, hasMore, fetch, loadMore, search: searchApi, searchMore, reset } = useGeokrety()
const { intToGkid } = useGkid()

const searchQuery = ref('')
const pageSize = 20

const { sentinel, observe } = useInfiniteScroll(() => {
  if (!loading.value && hasMore.value) doLoadMore()
})

onMounted(() => fetch(pageSize))

watch([geokrety, hasMore], () => {
  nextTick(() => observe())
})

function doSearch(): void {
  const q = searchQuery.value.trim()
  if (q) {
    searchApi(q, pageSize)
  } else {
    reset()
    fetch(pageSize)
  }
}

function doLoadMore(): void {
  const q = searchQuery.value.trim()
  if (q) {
    searchMore(q, pageSize)
  } else {
    loadMore(pageSize)
  }
}

function getGkid(gk: { gkid?: string | null; id: number }): string {
  return gk.gkid ?? intToGkid(gk.id)
}

function fmt(n: number | undefined): string {
  if (n === undefined || n === null) return '—'
  return n.toLocaleString('en')
}
</script>

<template>
  <main class="min-h-screen bg-background text-foreground pb-16">
    <div class="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 pt-10">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-6">
        <Package class="h-7 w-7 text-primary" />
        <h1 class="text-3xl font-bold tracking-tight">GeoKrety</h1>
      </div>

      <!-- Search bar -->
      <div class="flex gap-2 mb-6">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Search by name or GKID..."
            class="pl-9"
            @keydown.enter="doSearch"
          />
        </div>
        <Button @click="doSearch" :disabled="loading">Search</Button>
      </div>

      <!-- Error state -->
      <div v-if="error" class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 mb-6">
        <p class="text-sm text-destructive">{{ error }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetch(pageSize)">
          Retry
        </Button>
      </div>

      <!-- Loading initial -->
      <div v-if="loading && geokrety.length === 0" class="flex justify-center py-16">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
      </div>

      <!-- GeoKrety list -->
      <div v-else-if="geokrety.length > 0" class="space-y-3">
        <RouterLink
          v-for="gk in geokrety"
          :key="gk.id"
          :to="{ name: 'geokret-detail', params: { gkid: getGkid(gk) } }"
          class="block"
        >
          <Card class="hover:border-primary/30 transition-colors cursor-pointer">
            <CardContent class="flex items-center gap-4 p-4">
              <GeokretyTypeBadge :type="gk.type" icon-only />
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="font-semibold text-foreground truncate">{{ gk.name }}</span>
                  <span class="font-mono text-xs text-muted-foreground">
                    {{ getGkid(gk) }}
                  </span>
                </div>
                <div class="mt-0.5 text-sm text-muted-foreground">
                  <span v-if="gk.ownerUsername">Owner: {{ gk.ownerUsername }}</span>
                  <span v-if="gk.country"> · {{ gk.country }}</span>
                  <span v-if="gk.missing" class="text-destructive font-medium"> · Missing</span>
                </div>
              </div>
              <div class="flex-shrink-0 text-right text-sm text-muted-foreground">
                <div>❤️ {{ fmt(gk.lovesCount) }}</div>
                <div>📷 {{ fmt(gk.picturesCount) }}</div>
              </div>
            </CardContent>
          </Card>
        </RouterLink>

        <!-- Infinite scroll sentinel -->
        <div ref="sentinel" class="flex justify-center py-4">
          <div v-if="loading" class="h-6 w-6 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
          <p v-else-if="!hasMore" class="text-sm text-muted-foreground">All GeoKrety loaded.</p>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else class="text-center py-16">
        <Package class="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
        <p class="text-muted-foreground">No GeoKrety found.</p>
      </div>
    </div>
  </main>
</template>
