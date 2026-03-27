<script setup lang="ts">
/**
 * MultilingualMarkdown — renders markdown that may contain per-language sections.
 *
 * Supported header formats:
 *   ## 🇬🇧 English           flag emoji + full language name
 *   ## 🇫🇷 Français          flag emoji + native language name
 *   ## 🇩🇪 DE                flag emoji + ISO 3166-1 alpha-2 code (any case)
 *   ## FR                    ISO alpha-2 code only (any case)
 *   ## fr                    lowercase ISO code
 *
 * When at least two such language headers are present the component shows a
 * language-switcher and defaults to the visitor's browser language (or English,
 * or the first section as a last resort).  A "Show all" button is always
 * available to display the full unfiltered content.
 */
import { computed, ref } from 'vue'
import MarkdownContent from '@/components/MarkdownContent.vue'
import { ALPHA2_TO_LANG, countryCodeToName, countryCodeToFlag } from '@/lib/countryFlag'

interface LanguageSection {
  flagEmoji: string
  countryCode: string
  languageCode: string
  label: string
  content: string
}

interface ParsedMarkdown {
  isMultilingual: boolean
  preamble?: string
  sections: LanguageSection[]
}

// Pattern for one flag emoji (two regional-indicator surrogate pairs)
const FLAG_PAT = '\uD83C[\uDDE6-\uDDFF]\uD83C[\uDDE6-\uDDFF]'

// Language display name (lowercase) → BCP 47 primary subtag (for full-name headers)
const NAME_TO_CODE: Record<string, string> = {
  english: 'en',
  français: 'fr', french: 'fr',
  deutsch: 'de', german: 'de',
  español: 'es', spanish: 'es',
  polski: 'pl', polish: 'pl',
  português: 'pt', portuguese: 'pt',
  italiano: 'it', italian: 'it',
  nederlands: 'nl', dutch: 'nl',
  'русский': 'ru', russian: 'ru',
  '日本語': 'ja', japanese: 'ja',
  '中文': 'zh', chinese: 'zh',
  '한국어': 'ko', korean: 'ko',
  svenska: 'sv', norsk: 'nb', dansk: 'da', suomi: 'fi',
  čeština: 'cs', slovenčina: 'sk', magyar: 'hu', română: 'ro',
  türkçe: 'tr', ελληνικά: 'el', slovenščina: 'sl',
}

function decodeFlag(flag: string): string {
  const cp1 = flag.codePointAt(0)
  const cp2 = flag.codePointAt(2)
  if (cp1 == null || cp2 == null) return ''
  return String.fromCharCode(65 + cp1 - 0x1F1E6, 65 + cp2 - 0x1F1E6)
}

/** Returns true when `s` is a valid ISO 3166-1 alpha-2 country code. */
function isAlpha2Code(s: string): boolean {
  return /^[A-Za-z]{2}$/.test(s) && s.toUpperCase() in ALPHA2_TO_LANG
}

function parseMission(source: string): ParsedMarkdown {
  // Combined regex: optional flag emoji then either an ISO-2 code or a full language name
  // Examples: "## 🇬🇧 English", "## 🇩🇪 DE", "## FR", "## fr"
  const rx = new RegExp(
    `^##\\s+((?:${FLAG_PAT})\\s+)?([A-Za-z]{2}(?![A-Za-z])|[^\\n]+)$`,
    'gm',
  )
  const headers: { index: number; flag: string; label: string }[] = []
  let m: RegExpExecArray | null
  while ((m = rx.exec(source)) !== null) {
    const rawFlag = (m[1] ?? '').trim()
    const rawLabel = m[2]!.trim()
    // Reject headers that are just markdown like "## Some long title"
    // Accept only: ISO-2-only, or flag+anything, or known language names
    const hasFlag = rawFlag.length > 0
    const isCode = isAlpha2Code(rawLabel)
    const isName = rawLabel.toLowerCase() in NAME_TO_CODE
    if (!hasFlag && !isCode && !isName) continue
    headers.push({ index: m.index, flag: rawFlag, label: rawLabel })
  }

  if (headers.length < 2) {
    return { isMultilingual: false, sections: [] }
  }

  const preamble = source.slice(0, headers[0]!.index).trim()

  const sections: LanguageSection[] = headers.map((h, i) => {
    const lineEnd = source.indexOf('\n', h.index)
    const start = lineEnd === -1 ? source.length : lineEnd + 1
    const end = i + 1 < headers.length ? headers[i + 1]!.index : source.length
    const content = source
      .slice(start, end)
      .replace(/\n---\s*$/, '')
      .trim()

    // Determine country code
    let countryCode: string
    if (h.flag) {
      countryCode = decodeFlag(h.flag)
    } else if (isAlpha2Code(h.label)) {
      countryCode = h.label.toUpperCase()
    } else {
      countryCode = ''
    }

    // Determine language code
    const langFromCountry = countryCode ? (ALPHA2_TO_LANG[countryCode] ?? '') : ''
    const langFromName = NAME_TO_CODE[h.label.toLowerCase()] ?? ''
    const languageCode = langFromCountry || langFromName || countryCode.toLowerCase()

    // Determine display label (prefer full English name for ISO codes)
    let displayLabel: string
    if (isAlpha2Code(h.label) && !h.flag) {
      // Pure ISO code header like "## DE" — show full country name
      displayLabel = countryCodeToName(h.label) || h.label.toUpperCase()
    } else {
      displayLabel = h.label
    }

    // Build flag emoji if not provided
    const flagEmoji = h.flag || (countryCode ? countryCodeToFlag(countryCode) : '')

    return { flagEmoji, countryCode, languageCode, label: displayLabel, content }
  })

  return { isMultilingual: true, preamble: preamble || undefined, sections }
}

const props = defineProps<{ source: string }>()

const parsed = computed(() => parseMission(props.source ?? ''))

const browserLang = navigator.language.split('-')[0]!.toLowerCase()

const defaultLang = computed<string | null>(() => {
  if (!parsed.value.isMultilingual) return null
  return (
    parsed.value.sections.find((s) => s.languageCode === browserLang)?.languageCode ??
    parsed.value.sections.find((s) => s.languageCode === 'en')?.languageCode ??
    parsed.value.sections[0]?.languageCode ??
    null
  )
})

const activeLang = ref<string | null>(null)
const showAll = ref(false)

const effectiveLang = computed(() => (showAll.value ? null : (activeLang.value ?? defaultLang.value)))

const visibleSections = computed(() => {
  if (!parsed.value.isMultilingual) return []
  return effectiveLang.value === null
    ? parsed.value.sections
    : parsed.value.sections.filter((s) => s.languageCode === effectiveLang.value)
})

function selectLang(code: string): void {
  activeLang.value = code
  showAll.value = false
}
</script>

<template>
  <!-- Plain markdown (no language sections detected) -->
  <MarkdownContent v-if="!parsed.isMultilingual" :source="source" />

  <!-- Multilingual markdown -->
  <div v-else class="space-y-3">
    <!-- Optional preamble / title above language sections -->
    <MarkdownContent v-if="parsed.preamble" :source="parsed.preamble" />

    <!-- Language switcher -->
    <div class="flex flex-wrap items-center gap-1">
      <button
        v-for="s in parsed.sections"
        :key="s.languageCode"
        :class="[
          'rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors',
          effectiveLang === s.languageCode && !showAll
            ? 'border-border bg-accent text-accent-foreground'
            : 'border-border bg-card text-muted-foreground hover:text-foreground',
        ]"
        :aria-pressed="effectiveLang === s.languageCode && !showAll"
        :title="s.label"
        @click="selectLang(s.languageCode)"
      >
        {{ s.flagEmoji }} {{ s.label }}
      </button>
      <button
        :class="[
          'rounded-full border px-2.5 py-0.5 text-xs transition-colors',
          showAll
            ? 'border-border bg-accent text-accent-foreground'
            : 'border-border bg-card text-muted-foreground hover:text-foreground',
        ]"
        :aria-pressed="showAll"
        @click="showAll = !showAll"
      >
        Show all
      </button>
    </div>

    <!-- Section bodies -->
    <div v-for="s in visibleSections" :key="s.languageCode" class="space-y-0.5">
      <!-- Language label shown when displaying all sections at once -->
      <p v-if="showAll" class="text-xs font-semibold text-muted-foreground">
        {{ s.flagEmoji }} {{ s.label }}
      </p>
      <MarkdownContent :source="s.content" />
    </div>
  </div>
</template>
