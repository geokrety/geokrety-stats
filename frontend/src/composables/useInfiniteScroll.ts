/**
 * useInfiniteScroll — triggers a callback when a sentinel element enters the viewport.
 */
import { ref, onUnmounted, type Ref } from 'vue'

export function useInfiniteScroll(callback: () => void): {
  sentinel: Ref<HTMLElement | null>
  observe: () => void
  disconnect: () => void
} {
  const sentinel = ref<HTMLElement | null>(null)
  let observer: IntersectionObserver | null = null

  function observe(): void {
    disconnect()
    if (!sentinel.value) return
    observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0]
        if (entry?.isIntersecting) {
          callback()
        }
      },
      { rootMargin: '200px' },
    )
    observer.observe(sentinel.value)
  }

  function disconnect(): void {
    observer?.disconnect()
    observer = null
  }

  onUnmounted(() => disconnect())

  return { sentinel, observe, disconnect }
}
