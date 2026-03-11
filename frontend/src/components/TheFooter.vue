<script setup lang="ts">
import { MapPin, Github, ExternalLink } from 'lucide-vue-next'
import { Separator } from '@/components/ui/separator'
</script>

<template>
  <footer class="border-t border-border bg-background">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="grid gap-10 sm:grid-cols-3">
        <!-- Brand -->
        <div class="col-span-1">
          <div class="flex items-center gap-2 mb-3">
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
            >
              <MapPin class="h-5 w-5 text-primary-foreground" />
            </div>
            <span class="text-lg font-bold text-foreground">GeoKrety Stats</span>
          </div>
          <p class="text-sm text-muted-foreground leading-relaxed">
            Open-source statistics dashboard for the GeoKrety tracking platform.
          </p>
          <div
            class="mt-4 inline-flex items-center gap-1.5 rounded-full bg-card/5 px-3 py-1 text-xs text-muted-foreground ring-1 ring-border/10"
          >
            Powered by
            <a
              href="https://geokrety.org"
              target="_blank"
              rel="noopener noreferrer"
              class="text-foreground hover:text-foreground/80 transition-colors"
            >
              GeoKrety.org
            </a>
          </div>
        </div>

        <!-- Links -->
        <div>
          <h3 class="mb-4 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Platform
          </h3>
          <ul class="space-y-2.5">
            <li>
              <a
                href="https://geokrety.org"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                GeoKrety.org
                <ExternalLink class="h-3 w-3" />
              </a>
            </li>
            <li>
              <a
                href="https://geokrety.org/api"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                API Documentation
                <ExternalLink class="h-3 w-3" />
              </a>
            </li>
            <li>
              <a
                href="#stats"
                class="text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                Statistics
              </a>
            </li>
          </ul>
        </div>

        <div>
          <h3 class="mb-4 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Open Source
          </h3>
          <ul class="space-y-2.5">
            <li>
              <a
                href="https://github.com/geokrety/geokrety-stats"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                <Github class="h-4 w-4" />
                geokrety-stats
                <ExternalLink class="h-3 w-3" />
              </a>
            </li>
            <li>
              <a
                href="https://github.com/geokrety/geokrety-stats-api"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                <Github class="h-4 w-4" />
                geokrety-stats-api
                <ExternalLink class="h-3 w-3" />
              </a>
            </li>
            <li>
              <a
                href="https://github.com/geokrety/geokrety-website"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                <Github class="h-4 w-4" />
                geokrety-website
                <ExternalLink class="h-3 w-3" />
              </a>
            </li>
          </ul>
        </div>
      </div>

      <!-- Bottom bar -->
      <Separator class="mt-10 bg-border" />
      <div class="mt-6 flex flex-col sm:flex-row items-center justify-between gap-4 pt-0">
        <p class="text-xs text-muted-foreground">
          &copy; {{ new Date().getFullYear() }} GeoKrety Stats — Open Source Project
        </p>
        <p class="text-xs text-muted-foreground">
          Data provided by
          <a
            href="https://geokrety.org"
            target="_blank"
            rel="noopener noreferrer"
            class="text-muted-foreground transition-colors hover:text-foreground"
          >
            GeoKrety.org
          </a>
        </p>
      </div>
    </div>
  </footer>
</template>
