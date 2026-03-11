<script setup lang="ts">
import { LayoutGrid, Table2 } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

type View = 'table' | 'cards'

const props = defineProps<{ modelValue: View }>()
const emit = defineEmits<{ 'update:modelValue': [value: View] }>()
</script>

<template>
  <div class="flex items-center gap-1 rounded-lg border bg-background p-1">
    <Button
      :variant="modelValue === 'table' ? 'default' : 'ghost'"
      size="sm"
      class="gap-1.5"
      title="Table view"
      aria-label="Switch to table view"
      @click="emit('update:modelValue', 'table')"
    >
      <Table2 class="h-4 w-4" />
      <span class="hidden sm:inline">Table</span>
    </Button>
    <Button
      :variant="modelValue === 'cards' ? 'default' : 'ghost'"
      size="sm"
      class="gap-1.5"
      title="Card view"
      aria-label="Switch to card view"
      @click="emit('update:modelValue', 'cards')"
    >
      <LayoutGrid class="h-4 w-4" />
      <span class="hidden sm:inline">Cards</span>
    </Button>
  </div>
</template>
