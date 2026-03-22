/**
 * Base HTTP client for v3 API with error handling and retry logic.
 */
import type { ApiResponse, ApiErrorResponse } from '@/types/api'
import { getErrorMessage } from '@/lib/errorMessages'

const API_BASE = import.meta.env?.VITE_API_URL ?? 'http://192.168.130.65:7415'

const MAX_RETRIES = 3
const BASE_DELAY_MS = 1000
const REQUEST_TIMEOUT_MS = 30_000

export class ApiError extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly status?: number,
    public readonly requestId?: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }

  get userMessage(): string {
    return getErrorMessage(this.code)
  }
}

/**
 * Compute delay with exponential backoff and jitter.
 */
function retryDelay(attempt: number): number {
  const exponential = BASE_DELAY_MS * Math.pow(2, attempt)
  const jitter = Math.random() * 100
  return exponential + jitter
}

/**
 * Execute a fetch request with timeout support.
 */
async function fetchWithTimeout(url: string, init: RequestInit): Promise<Response> {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)
  try {
    return await fetch(url, { ...init, signal: controller.signal })
  } finally {
    clearTimeout(timeoutId)
  }
}

/**
 * Make a GET request to a v3 API endpoint.
 * Automatically retries on 5xx errors with exponential backoff.
 */
export async function apiGet<T>(path: string, params?: Record<string, string | number>): Promise<ApiResponse<T>> {
  const url = new URL(`${API_BASE}${path}`)
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      url.searchParams.set(key, String(value))
    }
  }

  let lastError: Error | undefined

  for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
    try {
      const res = await fetchWithTimeout(url.toString(), {
        method: 'GET',
        credentials: 'same-origin',
        headers: { Accept: 'application/json' },
      })

      if (res.ok) {
        return (await res.json()) as ApiResponse<T>
      }

      // Parse error response
      if (res.headers.get('content-type')?.includes('application/json')) {
        const errorBody = (await res.json()) as ApiErrorResponse
        const apiErr = new ApiError(
          errorBody.error?.code ?? 'UNKNOWN',
          errorBody.error?.message ?? res.statusText,
          res.status,
          errorBody.requestId,
        )

        // Only retry on 5xx
        if (res.status >= 500 && attempt < MAX_RETRIES) {
          lastError = apiErr
          await new Promise((resolve) => setTimeout(resolve, retryDelay(attempt)))
          continue
        }

        throw apiErr
      }

      throw new ApiError('UNKNOWN', `API ${res.status}: ${res.statusText}`, res.status)
    } catch (err) {
      if (err instanceof ApiError) throw err

      if (err instanceof DOMException && err.name === 'AbortError') {
        throw new ApiError('TIMEOUT', 'Request timed out. Please try again.')
      }

      if (err instanceof TypeError) {
        // Network error (fetch failed)
        if (attempt < MAX_RETRIES) {
          lastError = err
          await new Promise((resolve) => setTimeout(resolve, retryDelay(attempt)))
          continue
        }
        throw new ApiError('NETWORK_ERROR', 'Unable to connect to the server.')
      }

      throw err
    }
  }

  throw lastError ?? new ApiError('UNKNOWN', 'Request failed after retries')
}
