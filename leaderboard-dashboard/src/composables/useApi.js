import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: { Accept: 'application/json' },
})

// Unwrap data+meta from JSON:API subset envelope
api.interceptors.response.use(
  (r) => r,
  (err) => {
    const msg = err.response?.data?.error || err.message
    return Promise.reject(new Error(msg))
  }
)

/**
 * Fetch a paginated list (returns { items, meta })
 */
export async function fetchList(path, params = {}) {
  const { data } = await api.get(path, { params })
  return { items: data.data ?? [], meta: data.meta ?? {} }
}

/**
 * Fetch a single record (returns data.data)
 */
export async function fetchOne(path, params = {}) {
  const { data } = await api.get(path, { params })
  return data.data ?? data
}

export default api
