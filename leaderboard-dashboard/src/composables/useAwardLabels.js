import { ref, watch } from 'vue'
import { fetchList } from './useApi.js'
import { CHAIN_AWARD_LABELS } from './chainAwardLabels.js'

export function useAwardLabels(userIdRef) {
  const labels = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function loadLabels() {
    if (!userIdRef?.value) return
    loading.value = true
    error.value = null
    try {
      const { items = [] } = await fetchList(`/users/${userIdRef.value}/points/awards`, {
        per_page: 500,
        sort: 'label',
      })
      const seen = new Set()
      for (const award of items) {
        if (award.label) seen.add(award.label)
      }
      for (const chainLabel of CHAIN_AWARD_LABELS) {
        seen.add(chainLabel)
      }
      labels.value = [...seen].sort()
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  watch(userIdRef, () => {
    labels.value = []
    loadLabels()
  }, { immediate: true })

  return { labels, loading, error, reload: loadLabels }
}
