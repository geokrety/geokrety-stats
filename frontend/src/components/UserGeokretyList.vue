<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { GeokretListItem } from '@/types/api'
import type { UserGeokretyPage } from '@/services/api/users'
import GeokretListCard from '@/components/GeokretListCard.vue'
import { Button } from '@/components/ui/button'

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

onMounted(() => load())
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
    <Button v-if="hasMore && !loading" variant="ghost" size="sm" class="w-full mt-2" @click="load(true)">
      Load more
    </Button>
  </div>
</template>
