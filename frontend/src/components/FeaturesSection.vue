<script setup lang="ts">
import { Radio, BarChart2, Github, Users2 } from 'lucide-vue-next'

interface Feature {
  icon: unknown
  title: string
  description: string
  color: string
  bgColor: string
}

const features: Feature[] = [
  {
    icon: Radio,
    title: 'Real-time Tracking',
    description:
      'Movements are ingested via MQTT as they happen. Statistics update within seconds of every drop, grab, or log.',
    color: 'text-emerald-400',
    bgColor: 'bg-emerald-500/10',
  },
  {
    icon: BarChart2,
    title: 'Global Statistics',
    description:
      'From individual GeoKret journeys to country-level aggregations, every move contributes to a living world map.',
    color: 'text-cyan-400',
    bgColor: 'bg-cyan-500/10',
  },
  {
    icon: Github,
    title: 'Open Source',
    description:
      'Fully open-source under a free license. Inspect the code, file issues, or contribute on GitHub — community first.',
    color: 'text-violet-400',
    bgColor: 'bg-violet-500/10',
  },
  {
    icon: Users2,
    title: 'Community Driven',
    description:
      'Leaderboards, milestones, and activity feeds keep the GeoKrety community engaged and motivated to explore more.',
    color: 'text-amber-400',
    bgColor: 'bg-amber-500/10',
  },
]
</script>

<template>
  <section class="bg-background py-20">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div class="mb-14 text-center">
        <h2 class="text-3xl font-bold text-foreground sm:text-4xl">Why GeoKrety Stats?</h2>
        <p class="mt-3 text-muted-foreground max-w-xl mx-auto">
          Purpose-built to surface the metrics that matter most to the geocaching community
        </p>
      </div>

      <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <div
          v-for="feature in features"
          :key="feature.title"
          class="group rounded-2xl border border-border bg-card/5 p-6 transition-all hover:border-border/20 hover:-translate-y-1"
        >
          <div
            class="mb-5 inline-flex h-12 w-12 items-center justify-center rounded-2xl"
            :class="[feature.bgColor]"
          >
            <component :is="feature.icon" class="h-6 w-6" :class="feature.color" />
          </div>
          <h3 class="mb-2 text-lg font-semibold text-foreground">{{ feature.title }}</h3>
          <p class="text-sm leading-relaxed text-muted-foreground">{{ feature.description }}</p>
        </div>
      </div>
    </div>
  </section>
</template>
