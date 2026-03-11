<script setup lang="ts">
defineProps<{
  msg: string
}>()
</script>

<template>
  <div class="text-center lg:text-left">
    <h1 class="relative -top-2 text-4xl font-medium text-primary">{{ msg }}</h1>
    <h3 class="text-lg text-muted-foreground">
      You’ve successfully created a project with
      <a href="https://vite.dev/" target="_blank" rel="noopener">Vite</a> +
      <a href="https://vuejs.org/" target="_blank" rel="noopener">Vue 3</a>. What's next?
    </h3>
  </div>
</template>
