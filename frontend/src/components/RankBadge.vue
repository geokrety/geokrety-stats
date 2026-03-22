<script setup lang="ts">
import { Trophy } from 'lucide-vue-next'

defineProps<{
  rank: number
}>()

function medalClass(rank: number): string {
  if (rank === 1) return 'text-yellow-400'
  if (rank === 2) return 'text-muted-foreground'
  if (rank === 3) return 'text-amber-600'
  return ''
}
</script>

<template>
  <Trophy v-if="rank <= 3" class="h-5 w-5" :class="medalClass(rank)" />
  <span v-else class="font-medium text-muted-foreground">{{ rank }}</span>
</template>
