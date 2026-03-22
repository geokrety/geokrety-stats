/**
 * useUserProfile — composable for fetching user details.
 */
import { ref, type Ref } from 'vue'
import { getUserDetails, getUserOwnedGeokrety, getUserFoundGeokrety, type UserDetails, type UserGeokretyPage } from '@/services/api/users'
import { ApiError } from '@/services/api/client'

export function useUserProfile(): {
  user: Ref<UserDetails | null>
  loading: Ref<boolean>
  error: Ref<string | null>
  fetchUser: (userId: number) => Promise<void>
  fetchOwned: (userId: number, limit: number, cursor?: string) => Promise<UserGeokretyPage>
  fetchFound: (userId: number, limit: number, cursor?: string) => Promise<UserGeokretyPage>
} {
  const user = ref<UserDetails | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchUser(userId: number): Promise<void> {
    loading.value = true
    error.value = null
    try {
      user.value = await getUserDetails(userId)
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load user profile'
      console.error('[useUserProfile]', e)
    } finally {
      loading.value = false
    }
  }

  function fetchOwned(userId: number, limit: number, cursor?: string): Promise<UserGeokretyPage> {
    return getUserOwnedGeokrety(userId, limit, cursor)
  }

  function fetchFound(userId: number, limit: number, cursor?: string): Promise<UserGeokretyPage> {
    return getUserFoundGeokrety(userId, limit, cursor)
  }

  return { user, loading, error, fetchUser, fetchOwned, fetchFound }
}
