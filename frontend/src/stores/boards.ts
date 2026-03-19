import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Board } from '@/types/domain'

// Заготовка. Board API ещё не реализован на бэкенде.
export const useBoardsStore = defineStore('boards', () => {
  const boards = ref<Board[]>([])
  const loading = ref(false)

  async function fetchBoards(): Promise<void> {
    // TODO: await boardsApi.getBoards()
    loading.value = false
  }

  function clear(): void {
    boards.value = []
  }

  return { boards, loading, fetchBoards, clear }
})
