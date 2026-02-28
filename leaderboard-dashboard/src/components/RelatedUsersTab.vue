<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchList } from '../composables/useApi.js'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import Pagination from './Pagination.vue'

const props = defineProps({
  endpoint: {
    type: String,
    required: true
  },
  title: {
    type: String,
    default: 'Related Users'
  }
})

const users = ref([])
const page = ref(1)
const meta = ref({})
const loading = ref(false)
const error = ref(null)

async function loadUsers() {
  loading.value = true
  error.value = null
  try {
    const { items, meta: m } = await fetchList(props.endpoint, { page: page.value, per_page: 25 })
    users.value = items
    meta.value = m
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadUsers()
})

watch(() => page.value, loadUsers)
watch(() => props.endpoint, () => {
  page.value = 1
  loadUsers()
})
</script>

<template>
  <div v-if="loading" class="text-center py-5">
    <div class="spinner-border"></div>
  </div>
  <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
  <div v-else-if="users.length === 0" class="alert alert-info">No related users found.</div>
  <div v-else>
    <div class="card shadow-sm">
      <div class="card-header"><b>{{ title }}</b></div>
      <div class="table-responsive">
        <table class="table table-sm table-hover mb-0">
          <thead class="table-light">
            <tr>
              <th>User</th>
              <th class="text-end">Shared GeoKrety</th>
              <th class="text-end">Total Points</th>
              <th class="text-end">Total Moves</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.user_id">
              <td>
                <span v-if="user.home_country" class="me-1">{{ getCountryFlag(user.home_country) }}</span>
                <RouterLink :to="`/users/${user.user_id}`" class="text-decoration-none">
                  {{ user.username }}
                </RouterLink>
                <br/>
                <small class="text-muted">#{{ user.user_id }}</small>
              </td>
              <td class="text-end">{{ user.shared_geokrety_count?.toLocaleString() }}</td>
              <td class="text-end"><strong>{{ user.total_points?.toLocaleString() }}</strong></td>
              <td class="text-end">{{ user.total_moves?.toLocaleString() }}</td>
              <td class="text-end">
                <RouterLink :to="`/users/${user.user_id}`" class="btn btn-xs btn-outline-secondary py-0 px-1" style="font-size:0.75rem">
                  <i class="bi bi-arrow-right"></i>
                </RouterLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="card-footer d-flex justify-content-between align-items-center">
        <small class="text-muted">Showing {{ users.length }} of {{ meta.total }}</small>
        <Pagination v-if="meta" :meta="meta" :page="page" @update:page="page = $event" />
      </div>
    </div>
  </div>
</template>
