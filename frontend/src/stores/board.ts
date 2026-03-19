import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Board, Column } from '@/types/domain'

// Текущая открытая доска — aggregate root на фронте.
// Заготовка. Board API ещё не реализован на бэкенде.
export const useBoardStore = defineStore('board', () => {
  const board = ref<Board | null>(null)
  const columns = ref<Column[]>([])
  const version = ref(0)
  const loading = ref(false)

  async function fetchBoard(_boardId: string): Promise<void> {
    // TODO: const data = await boardsApi.getBoard(boardId)
    loading.value = false
  }

  // Optimistic move: snapshot → mutate → API → rollback on error
  async function moveCard(
    _cardId: string,
    _fromColumnId: string,
    _toColumnId: string,
    _newPosition: number,
  ): Promise<void> {
    const snapshot = structuredClone(columns.value)
    try {
      // TODO: applyCardMove locally + await boardsApi.moveCard(...)
      void snapshot // используется для rollback при ошибке
    } catch {
      columns.value = snapshot
    }
  }

  function clear(): void {
    board.value = null
    columns.value = []
    version.value = 0
  }

  return { board, columns, version, loading, fetchBoard, moveCard, clear }
})
