import { ref, watch } from 'vue'

export type ThemeId = 'light' | 'dark' | 'olive' | 'ocean'

export interface ThemeMeta {
  id: ThemeId
  label: string
  color: string // preview swatch color
}

export const themes: ThemeMeta[] = [
  { id: 'light', label: 'Светлая', color: '#7c5cfc' },
  { id: 'dark', label: 'Тёмная', color: '#111827' },
  { id: 'olive', label: 'Оливковая', color: '#6b7c4e' },
  { id: 'ocean', label: 'Океан', color: '#0891b2' },
]

const STORAGE_KEY = 'yammi-theme'

function getSavedTheme(): ThemeId {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved && themes.some(t => t.id === saved)) return saved as ThemeId
  return 'light'
}

const currentTheme = ref<ThemeId>(getSavedTheme())

export function useTheme() {
  function setTheme(id: ThemeId) {
    currentTheme.value = id
  }

  watch(currentTheme, (id) => {
    localStorage.setItem(STORAGE_KEY, id)
    document.documentElement.setAttribute('data-theme', id)
  }, { immediate: true })

  return { currentTheme, setTheme, themes }
}
