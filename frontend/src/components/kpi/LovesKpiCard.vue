<script setup lang="ts">
import BaseKpiCard from './BaseKpiCard.vue'

withDefaults(defineProps<{
  label: string
  value: string | number
  decimals?: number
}>(), {
  decimals: 0,
})
</script>

<template>
  <BaseKpiCard stat="loves" :label="label" :value="value" :decimals="decimals" />
</template>
