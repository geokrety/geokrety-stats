<script setup>
import { computed } from 'vue'
import LineChart from '../LineChart.vue'
import PointsBreakdownChart from '../PointsBreakdownChart.vue'
import PointsValue from '../PointsValue.vue'

const props = defineProps({
  userId: { type: [String, Number], required: true },
  user: { type: Object, default: () => ({}) },
  timeline: { type: Array, default: () => [] },
  breakdown: { type: Array, default: () => [] }
})

const emit = defineEmits(['openAwards'])

const today = new Date().toISOString().slice(0, 10)

function openAwards(source) {
  emit('openAwards', source)
}
</script>

<template>
  <div>
    <div class="card mb-4 shadow-sm">
      <div class="card-header d-flex justify-content-between align-items-center">
        <b>Points per Day</b>
      </div>
      <div class="card-body">
        <LineChart
          v-if="timeline.length"
          :data="timeline"
          x-key="day"
          y-key="points"
          color="#0d6efd"
          :height="220"
          :startDate="user?.joined_at?.slice(0, 10)"
          :endDate="today"
          :showRangeButtons="true"
        />
        <p v-else class="text-muted text-center py-3">No timeline data.</p>
      </div>
    </div>

    <div class="card mb-4 shadow-sm" v-if="breakdown.length">
      <div class="card-header"><b>Points by Bonus Type</b></div>
      <div class="card-body">
        <PointsBreakdownChart :data="breakdown" :height="300" />
      </div>
    </div>

    <div class="card shadow-sm">
      <div class="card-header"><b>Points Breakdown</b></div>
      <div class="card-body pb-0">
        <p class="text-muted small mb-2">This panel summarizes where this user earns points, grouped by reward source. Use the eye action or row click to open matching entries in Points Log.</p>
      </div>
      <div class="table-responsive border-0 mb-0">
        <table class="table table-sm table-hover mb-0 align-middle">
          <thead class="table-light">
            <tr>
              <th title="Activity or bonus type that awarded points to this user">Source</th>
              <th class="text-end" title="Total points earned from this source">Points</th>
              <th class="text-end d-none d-sm-table-cell" title="Number of times this reward was earned">Count</th>
              <th class="text-end" style="width: 50px" title="Open this source in Points Log"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="b in breakdown" :key="b.source" @click="openAwards(b.source)" style="cursor: pointer">
              <td class="fw-medium">{{ b.source }}</td>
              <td class="text-end fw-bold"><PointsValue :value="b.points" /></td>
              <td class="text-end d-none d-sm-table-cell text-muted">{{ b.count?.toLocaleString() }}</td>
              <td class="text-end">
                <button type="button" class="btn btn-sm btn-outline-secondary py-0 px-1" style="font-size:0.75rem" title="View matching point awards" @click.stop="openAwards(b.source)"><i class="bi bi-eye"></i></button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="card-footer text-end">
        <button type="button" class="btn btn-sm btn-outline-primary shadow-sm" @click="openAwards()">
          <i class="bi bi-list-stars me-1"></i>View all point awards
        </button>
      </div>
    </div>
  </div>
</template>
