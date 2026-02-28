<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchList } from '../composables/useApi.js'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import Pagination from './Pagination.vue'
import UserCard from './UserCard.vue'

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
    <div class="row row-cols-1 row-cols-sm-2 row-cols-md-3 row-cols-xl-4 g-3 mb-4">
      <div v-for="user in users" :key="user.user_id" class="col">
        <UserCard :user="user" />
      </div>
    </div>
    <div class="d-flex justify-content-between align-items-center mt-4">
      <small class="text-muted">Showing {{ users.length }} of {{ meta.total }}</small>
      <Pagination v-if="meta" :meta="meta" :page="page" @update:page="page = $event" />
    </div>
  </div>
</template>
