import { computed, ref, watch } from 'vue'
import {
  DEFAULT_THEME_ID,
  getThemeById,
  THEME_STORAGE_KEY,
  THEMES,
} from '@/config/themes'
import {
  applyFaviconToDocument,
  roundedFaviconDataUrl,
  themeIconSrc,
} from '@/utils/faviconRounded'

function readStoredThemeId() {
  if (typeof window === 'undefined') return DEFAULT_THEME_ID
  const raw = Number.parseInt(localStorage.getItem(THEME_STORAGE_KEY) || '', 10)
  return THEMES.some((t) => t.id === raw) ? raw : DEFAULT_THEME_ID
}

const themeId = ref(readStoredThemeId())
let faviconGeneration = 0

function applyThemeToDocument(theme) {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  root.dataset.theme = String(theme.id)
  root.style.setProperty('--ios-blue', theme.accent)
  root.style.setProperty('--primary', theme.accent)
  root.style.setProperty('--theme-accent', theme.accent)
  root.style.setProperty('--theme-accent-soft', theme.accentSoft)
  root.style.setProperty('--theme-bg-base', theme.bgBase)
  root.style.setProperty('--theme-wallpaper', theme.wallpaper)
  root.style.setProperty('--theme-mesh-1', theme.mesh1)
  root.style.setProperty('--theme-mesh-2', theme.mesh2)
  root.style.setProperty('--theme-mesh-3', theme.mesh3)
  root.style.setProperty('--theme-mesh-4', theme.mesh4)
  root.style.setProperty('--theme-mesh-opacity', String(theme.meshOpacity ?? 0.58))
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta) meta.setAttribute('content', theme.bgBase)
  updateFavicon(theme.id)
}

async function updateFavicon(id) {
  const gen = ++faviconGeneration
  const src = themeIconSrc(id)

  try {
    const [iconHref, appleHref] = await Promise.all([
      roundedFaviconDataUrl(src, 64, `theme-${id}`),
      roundedFaviconDataUrl(src, 180, `theme-${id}`),
    ])
    if (gen !== faviconGeneration) return
    applyFaviconToDocument({ iconHref, appleHref })
  } catch {
    if (gen !== faviconGeneration) return
    applyFaviconToDocument({ iconHref: src, appleHref: src })
  }
}

applyThemeToDocument(getThemeById(themeId.value))

export function useTheme() {
  const currentThemeId = computed({
    get: () => themeId.value,
    set: (id) => {
      const theme = getThemeById(id)
      themeId.value = theme.id
      localStorage.setItem(THEME_STORAGE_KEY, String(theme.id))
      applyThemeToDocument(theme)
    },
  })

  const currentTheme = computed(() => getThemeById(themeId.value))

  watch(currentTheme, (theme) => applyThemeToDocument(theme), { immediate: true })

  return {
    themes: THEMES,
    currentThemeId,
    currentTheme,
    setTheme(id) {
      currentThemeId.value = id
    },
  }
}
