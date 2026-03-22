/**
 * useGkid — composable for GKID conversion in Vue components.
 *
 * Delegates to the pure utility functions in @/lib/gkid.
 */
import { intToGkid, gkidToInt, validateGkid } from '@/lib/gkid'

export function useGkid(): {
  intToGkid: (id: number) => string
  gkidToInt: (gkid: string) => number | null
  isGkid: (value: string) => boolean
} {
  return { intToGkid, gkidToInt, isGkid: validateGkid }
}
