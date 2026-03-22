<script setup lang="ts">
import { computed } from 'vue'
import VueMarkdown from 'vue-markdown-render'
import DOMPurify from 'dompurify'

const props = withDefaults(defineProps<{
  source: string
  inline?: boolean
}>(), {
  inline: false,
})

const sanitizedSource = computed(() => DOMPurify.sanitize(props.source ?? ''))
const markdownOptions = {
  html: false,
  linkify: true,
  breaks: true,
}
</script>

<template>
  <div :class="inline ? 'markdown-inline' : 'markdown-content'">
    <VueMarkdown :source="sanitizedSource" :options="markdownOptions" />
  </div>
</template>

<style scoped>
.markdown-content :deep(p) {
  margin: 0.5rem 0;
}

.markdown-content :deep(p:first-child) {
  margin-top: 0;
}

.markdown-content :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown-content :deep(a),
.markdown-inline :deep(a) {
  color: hsl(var(--primary));
  text-decoration: underline;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 0.5rem 0;
  padding-left: 1.25rem;
}

.markdown-content :deep(code),
.markdown-inline :deep(code) {
  border-radius: 0.25rem;
  background: hsl(var(--muted));
  padding: 0.1rem 0.25rem;
  font-size: 0.875em;
}

.markdown-inline :deep(p) {
  display: inline;
  margin: 0;
}
</style>
