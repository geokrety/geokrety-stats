<script setup lang="ts">
import { onMounted, computed, nextTick, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useGeokretDetail } from '@/composables/useGeokretDetail'
import { useGkid } from '@/composables/useGkid'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'
import GeokretyTypeBadge from '@/components/GeokretyTypeBadge.vue'
import GeokretDetailKpis from '@/components/GeokretDetailKpis.vue'
import GeokretMovesMap from '@/components/GeokretMovesMap.vue'
import MultilingualMarkdown from '@/components/MultilingualMarkdown.vue'
import MarkdownContent from '@/components/MarkdownContent.vue'
import MoveTypeBadge from '@/components/MoveTypeBadge.vue'
import AppBreadcrumb from '@/components/AppBreadcrumb.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { formatNumber } from '@/lib/format'
import { User, Calendar } from 'lucide-vue-next'
import { formatDateTime, relativeTime } from '@/lib/dates'
import { countryCodeToFlag } from '@/lib/countryFlag'

const route = useRoute()
const { intToGkid } = useGkid()
const { geokret, loading, error, moves, movesLoading, movesHasMore, fetchDetails, fetchMoves, loadMoreMoves } = useGeokretDetail()

const gkidParam = computed(() => route.params.gkid as string)

const displayGkid = computed(() => {
  if (!geokret.value) return gkidParam.value
  return geokret.value.gkid ?? intToGkid(geokret.value.id)
})

function doLoadMore(): void {
  if (movesHasMore.value && !movesLoading.value) {
    loadMoreMoves(gkidParam.value).then(() => {
      nextTick(() => { if (movesHasMore.value) observe() })
    })
  }
}

const { sentinel, observe } = useInfiniteScroll(doLoadMore)

onMounted(async () => {
  fetchDetails(gkidParam.value)
  await fetchMoves(gkidParam.value)
  await nextTick()
  if (movesHasMore.value) observe()
})

watch([moves, movesHasMore], () => {
  nextTick(() => {
    if (movesHasMore.value) observe()
  })
})

</script>

<template>
  <main class="min-h-screen bg-background text-foreground pb-16">
    <div class="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 pt-10">
      <AppBreadcrumb :items="[
        { label: 'GeoKrety', to: '/geokrety' },
        { label: geokret?.name ?? gkidParam },
      ]" />
      <!-- Error state -->
      <div v-if="error" class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 mb-6">
        <p class="text-sm text-destructive">{{ error }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetchDetails(gkidParam)">Retry</Button>
      </div>

      <!-- Loading state -->
      <div v-if="loading" class="flex justify-center py-16">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
      </div>

      <!-- Detail content -->
      <div v-else-if="geokret">
        <!-- Header -->
        <div class="flex items-start gap-4 mb-8">
          <GeokretyTypeBadge :type="geokret.type" icon-only class="mt-1" />
          <div class="min-w-0 flex-1">
            <h1 class="text-3xl font-bold tracking-tight">{{ geokret.name }}</h1>
            <div class="flex items-center gap-2 mt-1">
              <span class="font-mono text-sm text-muted-foreground">{{ displayGkid }}</span>
              <span class="text-sm text-muted-foreground">· {{ geokret.typeName }}</span>
              <span v-if="geokret.missing" class="text-sm text-destructive font-medium">· Missing</span>
            </div>
          </div>
        </div>

        <GeokretDetailKpis :geokret="geokret" />

        <!-- Details card -->
        <Card class="mb-6">
          <CardHeader>
            <CardTitle class="text-lg">Details</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
              <div v-if="geokret.ownerUsername" class="flex items-center gap-2">
                <User class="h-4 w-4 text-muted-foreground" />
                <span class="text-muted-foreground">Owner:</span>
                <RouterLink
                  v-if="geokret.ownerId"
                  :to="{ name: 'user-profile', params: { id: geokret.ownerId } }"
                  class="font-medium text-primary hover:underline"
                >
                  {{ geokret.ownerUsername }}
                </RouterLink>
                <span v-else class="font-medium">{{ geokret.ownerUsername }}</span>
              </div>
              <div v-if="geokret.holderUsername" class="flex items-center gap-2">
                <User class="h-4 w-4 text-muted-foreground" />
                <span class="text-muted-foreground">Held by:</span>
                <RouterLink
                  v-if="geokret.holderId"
                  :to="{ name: 'user-profile', params: { id: geokret.holderId } }"
                  class="font-medium text-primary hover:underline"
                >
                  {{ geokret.holderUsername }}
                </RouterLink>
                <span v-else class="font-medium">{{ geokret.holderUsername }}</span>
              </div>
              <div v-if="geokret.country" class="flex items-center gap-2">
                <MapPin class="h-4 w-4 text-muted-foreground" />
                <span class="text-muted-foreground">Country:</span>
                <span class="font-medium">{{ geokret.country }}</span>
              </div>
              <div v-if="geokret.waypoint" class="flex items-center gap-2">
                <MapPin class="h-4 w-4 text-muted-foreground" />
                <span class="text-muted-foreground">Waypoint:</span>
                <span class="font-mono text-xs">{{ geokret.waypoint }}</span>
              </div>
              <div v-if="geokret.bornAt" class="flex items-center gap-2">
                <Calendar class="h-4 w-4 text-muted-foreground" />
                <span class="text-muted-foreground">Born:</span>
                <span>{{ formatDateTime(geokret.bornAt) }}</span>
              </div>
              <div v-if="geokret.lastMoveAt" class="flex items-center gap-2">
                <Calendar class="h-4 w-4 text-muted-foreground" />
                <span class="text-muted-foreground">Last move:</span>
                <span>{{ relativeTime(geokret.lastMoveAt) }}</span>
              </div>
            </div>
            <div v-if="geokret.mission" class="text-sm">
              <span class="text-muted-foreground">Mission:</span>
              <MultilingualMarkdown class="mt-1" :source="geokret.mission" />
            </div>
          </CardContent>
        </Card>

        <!-- Movement Map (U17) -->
        <Card v-if="moves.length > 0" class="mb-6">
          <CardHeader>
            <CardTitle class="text-lg">Movement Map</CardTitle>
          </CardHeader>
          <CardContent>
            <GeokretMovesMap :moves="moves" />
          </CardContent>
        </Card>

        <!-- Move History (U18) -->
        <Card class="mb-6">
          <CardHeader>
            <CardTitle class="text-lg">Move History</CardTitle>
          </CardHeader>
          <CardContent>
            <div v-if="moves.length === 0 && !movesLoading" class="text-sm text-muted-foreground">No moves recorded yet.</div>
            <ul class="divide-y divide-border">
              <li v-for="m in moves" :key="m.id" class="py-3">
                <div class="flex items-start justify-between gap-2">
                  <div class="flex items-start gap-2 min-w-0">
                    <MoveTypeBadge :type="m.moveTypeName" class="mt-0.5 shrink-0" />
                    <div class="min-w-0">
                      <div class="text-sm font-medium">
                        <span v-if="m.waypoint" class="font-mono text-xs text-muted-foreground">{{ m.waypoint }}</span>
                        <span v-if="m.country" class="text-xs text-muted-foreground ml-1">{{ countryCodeToFlag(m.country) }} {{ m.country }}</span>
                      </div>
                      <div class="text-xs text-muted-foreground mt-0.5">
                        {{ formatDateTime(m.movedOn) }}
                        <span v-if="m.username">
                          · by
                          <RouterLink
                            v-if="m.authorId"
                            :to="{ name: 'user-profile', params: { id: m.authorId } }"
                            class="text-primary hover:underline"
                          >
                            {{ m.username }}
                          </RouterLink>
                          <span v-else>{{ m.username }}</span>
                        </span>
                      </div>
                      <MarkdownContent
                        v-if="m.comment && !m.commentHidden"
                        class="mt-1 line-clamp-2 text-xs text-muted-foreground"
                        :source="m.comment"
                      />
                    </div>
                  </div>
                  <div v-if="m.kmDistance" class="text-xs text-muted-foreground whitespace-nowrap">
                    {{ formatNumber(m.kmDistance, 1) }} km
                  </div>
                </div>
              </li>
            </ul>
            <div v-if="movesLoading" class="flex justify-center py-3">
              <div class="h-5 w-5 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
            </div>
            <div ref="sentinel" class="h-1" />
          </CardContent>
        </Card>

        <!-- Back link -->
      </div>

      <!-- Not found -->
      <div v-else-if="!loading" class="text-center py-16">
        <Package class="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
        <p class="text-muted-foreground">GeoKret not found.</p>
      </div>
    </div>
  </main>
</template>

