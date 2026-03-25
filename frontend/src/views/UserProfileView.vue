<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserProfile } from '@/composables/useUserProfile'
import UserInfoHeader from '@/components/UserInfoHeader.vue'
import UserStatsGrid from '@/components/UserStatsGrid.vue'
import UserGeokretyList from '@/components/UserGeokretyList.vue'
import UserOwnedGeokretyMap from '@/components/UserOwnedGeokretyMap.vue'
import AppBreadcrumb from '@/components/AppBreadcrumb.vue'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { User } from 'lucide-vue-next'

const route = useRoute()
const { user, loading, error, fetchUser, fetchOwned, fetchFound } = useUserProfile()

const userId = computed(() => Number(route.params.id))

function getOwned(limit: number, cursor?: string) {
  return fetchOwned(userId.value, limit, cursor)
}

function getFound(limit: number, cursor?: string) {
  return fetchFound(userId.value, limit, cursor)
}

onMounted(() => fetchUser(userId.value))
</script>

<template>
  <main class="min-h-screen bg-background text-foreground pb-16">
    <div class="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 pt-10">
      <AppBreadcrumb :items="[
        { label: 'Users', to: '/users' },
        { label: user?.username ?? `User #${userId}` },
      ]" />

      <!-- Error state -->
      <div v-if="error" class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 mb-6">
        <p class="text-sm text-destructive">{{ error }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetchUser(userId)">Retry</Button>
      </div>

      <!-- Loading state -->
      <div v-if="loading" class="flex justify-center py-16">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
      </div>

      <!-- User profile -->
      <div v-else-if="user">
        <UserInfoHeader :user="user" />
        <UserStatsGrid :user="user" />
        <div class="mb-8">
          <UserOwnedGeokretyMap :fetch-fn="getOwned" />
        </div>

        <Tabs default-value="owned" class="w-full">
          <TabsList class="grid w-full grid-cols-2">
            <TabsTrigger value="owned">Owned GeoKrety</TabsTrigger>
            <TabsTrigger value="found">Found GeoKrety</TabsTrigger>
          </TabsList>
          <TabsContent value="owned" class="mt-4">
            <UserGeokretyList :fetch-fn="getOwned" empty-text="No owned GeoKrety." />
          </TabsContent>
          <TabsContent value="found" class="mt-4">
            <UserGeokretyList :fetch-fn="getFound" empty-text="No found GeoKrety." />
          </TabsContent>
        </Tabs>
      </div>

      <!-- Not found -->
      <div v-else-if="!loading" class="text-center py-16">
        <User class="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
        <p class="text-muted-foreground">User not found.</p>
      </div>
    </div>
  </main>
</template>
