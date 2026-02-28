<script setup>
import { useUserCard } from '../composables/useUserCard.js'
import { RouterLink } from 'vue-router'

const props = defineProps({
  user: {
    type: Object,
    required: true
  }
})

const {
  username,
  userId,
  flag,
  stats,
  leaderboardUrl,
} = useUserCard(props.user)
</script>

<template>
  <div class="card h-100 shadow-sm border-0 user-card">
    <div class="card-body p-3">
      <div class="d-flex align-items-center gap-2 mb-3">
        <div class="avatar-circle bg-light border text-primary">
          {{ username?.[0]?.toUpperCase() }}
        </div>
        <div class="flex-grow-1 overflow-hidden">
          <RouterLink :to="leaderboardUrl" class="text-decoration-none d-block text-truncate fw-bold text-dark">
            <span v-if="flag" class="me-1">{{ flag }}</span> {{ username }}
          </RouterLink>
          <small class="text-muted">#{{ userId }}</small>
        </div>
      </div>

      <div class="row g-2 text-center small">
        <div class="col-4">
          <div class="text-muted" style="font-size: 0.65rem">POINTS</div>
          <div class="fw-bold">{{ stats.points?.toLocaleString() }}</div>
        </div>
        <div class="col-4 border-start border-end">
          <div class="text-muted" style="font-size: 0.65rem">MOVES</div>
          <div class="fw-bold">{{ stats.moves?.toLocaleString() }}</div>
        </div>
        <div class="col-4">
          <div class="text-muted" style="font-size: 0.65rem">GKs</div>
          <div class="fw-bold">{{ stats.geokrety?.toLocaleString() }}</div>
        </div>
      </div>
    </div>
    <div class="card-footer bg-white border-top-0 pt-0 pb-3 text-center">
      <RouterLink :to="leaderboardUrl" class="btn btn-sm btn-outline-primary w-100 rounded-pill py-1">
        View Profile <i class="bi bi-arrow-right ms-1"></i>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
.user-card {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.user-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 0.5rem 1rem rgba(0,0,0,0.08) !important;
}
.avatar-circle {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 1.2rem;
  flex-shrink: 0;
}
</style>
