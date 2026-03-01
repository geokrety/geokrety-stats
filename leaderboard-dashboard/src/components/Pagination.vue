<script setup>
import { computed } from 'vue'

const props = defineProps({
  meta: { type: Object, required: true },
  page: { type: Number, required: true },
})
const emit = defineEmits(['update:page'])

const totalPages = computed(() =>
  props.meta.per_page ? Math.ceil(props.meta.total / props.meta.per_page) : 1
)

function go(p) {
  if (p < 1 || p > totalPages.value) return
  emit('update:page', p)
}

const pages = computed(() => {
  const cur = props.page
  const last = totalPages.value
  const delta = 2
  const range = []
  for (let i = Math.max(1, cur - delta); i <= Math.min(last, cur + delta); i++) range.push(i)
  return range
})
</script>

<template>
  <nav aria-label="pagination">
    <ul class="pagination pagination-sm justify-content-center mb-0">
      <li class="page-item" :class="{ disabled: page <= 1 }">
        <button type="button" class="page-link" @click="go(1)">&laquo;</button>
      </li>
      <li class="page-item" :class="{ disabled: page <= 1 }">
        <button type="button" class="page-link" @click="go(page - 1)">&lsaquo;</button>
      </li>
      <li v-for="p in pages" :key="p" class="page-item" :class="{ active: p === page }">
        <button type="button" class="page-link" @click="go(p)">{{ p }}</button>
      </li>
      <li class="page-item" :class="{ disabled: page >= totalPages }">
        <button type="button" class="page-link" @click="go(page + 1)">&rsaquo;</button>
      </li>
      <li class="page-item" :class="{ disabled: page >= totalPages }">
        <button type="button" class="page-link" @click="go(totalPages)">&raquo;</button>
      </li>
    </ul>
    <p class="text-center text-muted small mt-1">
      Page {{ page }} of {{ totalPages }}
      <span v-if="meta.total">({{ meta.total?.toLocaleString() }} entries)</span>
    </p>
  </nav>
</template>
