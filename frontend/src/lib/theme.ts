import { useColorMode } from '@vueuse/core'
import { createContext } from 'reka-ui'
import type { Ref } from 'vue'

const colorMode = useColorMode({ storageKey: 'geokrety-theme' } as any)

export const cycleColorMode = () => {
  if (colorMode.value === 'dark') colorMode.value = 'light'
  else if (colorMode.value === 'light') colorMode.value = 'auto'
  else colorMode.value = 'dark'
}

export const isDark = () => colorMode.value === 'dark'

export { colorMode }

// Context helpers for providing/consuming theme-related utilities
export const [useTheme, provideTheme] = createContext<{
  colorMode: Ref<string>
  cycleColorMode: () => void
  isDark: () => boolean
}>('Theme')
