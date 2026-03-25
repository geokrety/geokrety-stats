<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import type { GeokretListItem } from '@/types/api'
import type { UserGeokretyPage } from '@/services/api/users'
import GeokretListCard from '@/components/GeokretListCard.vue'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'

const props = defineProps<{
  fetchFn: (limit: number, cursor?: string) => Promise<UserGeokretyPage>
  emptyText?: string
}>()

const items = ref<GeokretListItem[]>([])
const loading = ref(false)
const cursor = ref<string | undefined>()
const hasMore = ref(false)

async function load(append = false): Promise<void> {
  loading.value = true
  try {
    const page = await props.fetchFn(10, append ? cursor.value : undefined)
    items.value = append ? [...items.value, ...page.data] : page.data
    cursor.value = page.nextCursor
    hasMore.value = page.hasMore
  } catch (e) {
    console.error('[UserGeokretyList]', e)
  } finally {
    loading.value = false
  }
}

const { sentinel, observe } = useInfiniteScroll(() => {
  if (!loading.value && hasMore.value) {
    load(true)
  }
})

onMounted(() => load())

watch([items, hasMore], () => {
  nextTick(() => {
    if (hasMore.value) observe()
  })
})
</script>

<template>
  <div>
    <div v-if="items.length === 0 && !loading" class="text-sm text-muted-foreground">
      {{ emptyText ?? 'No GeoKrety.' }}
    </div>
    <div class="divide-y divide-border">
      <GeokretListCard v-for="gk in items" :key="gk.id" :gk="gk" />
    </div>
    <div v-if="loading" class="flex justify-center py-2">
      <div class="h-5 w-5 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
    </div>
    <div ref="sentinel" class="h-1" />
  </div>
</template>
