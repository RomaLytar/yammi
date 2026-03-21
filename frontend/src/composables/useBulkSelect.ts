import { ref, computed } from 'vue'

export function useBulkSelect(canSelect: (id: string) => boolean) {
  const selectMode = ref(false)
  const selectedIds = ref<Set<string>>(new Set())

  const selectedCount = computed(() => selectedIds.value.size)

  function toggleSelectMode() {
    selectMode.value = !selectMode.value
    selectedIds.value = new Set()
  }

  function toggleSelect(id: string, e?: Event) {
    if (e) e.stopPropagation()
    const s = new Set(selectedIds.value)
    if (s.has(id)) {
      s.delete(id)
    } else {
      s.add(id)
    }
    selectedIds.value = s
  }

  function toggleSelectAll(allIds: string[]) {
    const selectableIds = allIds.filter(canSelect)
    const allSelected = selectableIds.length > 0 && selectableIds.every(id => selectedIds.value.has(id))
    if (allSelected) {
      selectedIds.value = new Set()
    } else {
      selectedIds.value = new Set(selectableIds)
    }
  }

  function clearSelection() {
    selectedIds.value = new Set()
    selectMode.value = false
  }

  return {
    selectMode,
    selectedIds,
    selectedCount,
    toggleSelectMode,
    toggleSelect,
    toggleSelectAll,
    clearSelection,
  }
}
