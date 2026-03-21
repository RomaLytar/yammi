import { ref, watch } from 'vue'

export function useDebouncedSearch(delay = 300) {
  const searchInput = ref('')
  const debouncedValue = ref('')
  let searchTimer: ReturnType<typeof setTimeout> | null = null

  watch(searchInput, (val) => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      debouncedValue.value = val.trim()
    }, delay)
  })

  function clearSearch() {
    searchInput.value = ''
    debouncedValue.value = ''
  }

  return {
    searchInput,
    debouncedValue,
    clearSearch,
  }
}
