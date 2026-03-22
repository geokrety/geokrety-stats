/**
 * useUsers — composable for fetching and searching users with cursor-based pagination.
 */
import { ref, type Ref } from 'vue'
import { listUsers, searchUsers, type UsersPage } from '@/services/api/users'
import { ApiError } from '@/services/api/client'
import type { UserListItem } from '@/types/api'

export function useUsers(): {
  users: Ref<UserListItem[]>
  loading: Ref<boolean>
  error: Ref<string | null>
  hasMore: Ref<boolean>
  fetch: (limit?: number) => Promise<void>
  loadMore: (limit?: number) => Promise<void>
  search: (query: string, limit?: number) => Promise<void>
  reset: () => void
} {
  const users = ref<UserListItem[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const hasMore = ref(false)
  let nextCursor: string | undefined
  let lastQuery = ''

  async function fetch(limit = 20): Promise<void> {
    loading.value = true
    error.value = null
    nextCursor = undefined
    lastQuery = ''
    try {
      const page: UsersPage = await listUsers(limit)
      users.value = page.data
      nextCursor = page.nextCursor
      hasMore.value = page.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load users'
      console.error('[useUsers]', e)
    } finally {
      loading.value = false
    }
  }

  async function loadMore(limit = 20): Promise<void> {
    if (loading.value || !hasMore.value) return
    loading.value = true
    try {
      const page: UsersPage = lastQuery
        ? await searchUsers(lastQuery, limit, nextCursor)
        : await listUsers(limit, nextCursor)
      users.value = [...users.value, ...page.data]
      nextCursor = page.nextCursor
      hasMore.value = page.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load more users'
      console.error('[useUsers]', e)
    } finally {
      loading.value = false
    }
  }

  async function search(query: string, limit = 20): Promise<void> {
    loading.value = true
    error.value = null
    nextCursor = undefined
    lastQuery = query
    try {
      const page: UsersPage = query
        ? await searchUsers(query, limit)
        : await listUsers(limit)
      users.value = page.data
      nextCursor = page.nextCursor
      hasMore.value = page.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to search users'
      console.error('[useUsers]', e)
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    users.value = []
    loading.value = false
    error.value = null
    hasMore.value = false
    nextCursor = undefined
    lastQuery = ''
  }

  return { users, loading, error, hasMore, fetch, loadMore, search, reset }
}
