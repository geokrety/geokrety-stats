<script setup lang="ts">
import { Sun, Moon } from 'lucide-vue-next'
import { useTheme } from '@/lib/theme'
import { Button } from '@/components/ui/button'
import type { HTMLAttributes } from 'vue'

const { cycleColorMode, isDark } = useTheme()
const props = withDefaults(
  defineProps<{ asMenuButton?: boolean; class?: HTMLAttributes['class'] }>(),
  {
    asMenuButton: false,
  },
)
</script>

<template>
  <template v-if="props.asMenuButton">
    <Sun v-if="isDark()" class="size-4" />
    <Moon v-else class="size-4" />
    <span>{{ isDark() ? 'Switch to light mode' : 'Switch to dark mode' }}</span>
  </template>

  <template v-else>
    <Button
      variant="ghost"
      size="icon"
      :class="props.class"
      :aria-label="isDark() ? 'Switch to light mode' : 'Switch to dark mode'"
      @click="cycleColorMode"
    >
      <Sun v-if="isDark()" class="h-4 w-4" />
      <Moon v-else class="h-4 w-4" />
    </Button>
  </template>
</template>
