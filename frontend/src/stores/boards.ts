import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Board } from '@/types/domain'
import * as boardsApi from '@/api/boards'
import { ApiError } from '@/api/client'

export const useBoardsStore = defineStore('boards', () => {
  const boards = ref<Board[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const nextCursor = ref<string | undefined>(undefined)
  const hasMore = ref(true)

  // Filters
  const ownerOnly = ref(false)
  const search = ref('')
  const sortBy = ref<'updated_at' | 'created_at' | 'title'>('updated_at')

  async function fetchBoards(reset = false): Promise<void> {
    if (loading.value) return

    try {
      loading.value = true
      error.value = null

      const cursor = reset ? undefined : nextCursor.value
      const result = await boardsApi.getBoards(20, cursor, ownerOnly.value, search.value, sortBy.value)

      if (reset) {
        boards.value = result.boards
      } else {
        boards.value.push(...result.boards)
      }

      nextCursor.value = result.nextCursor
      hasMore.value = !!result.nextCursor
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка загрузки досок'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createBoard(title: string, description: string): Promise<Board> {
    try {
      error.value = null
      const board = await boardsApi.createBoard({ title, description })
      boards.value.unshift(board)
      return board
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка создания доски'
      throw err
    }
  }

  async function deleteBoards(boardIds: string[]): Promise<void> {
    try {
      error.value = null
      await boardsApi.deleteBoards(boardIds)
      const idsSet = new Set(boardIds)
      boards.value = boards.value.filter((b) => !idsSet.has(b.id))
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка удаления досок'
      throw err
    }
  }

  function clear(): void {
    boards.value = []
    nextCursor.value = undefined
    hasMore.value = true
    error.value = null
  }

  return {
    boards,
    loading,
    error,
    hasMore,
    ownerOnly,
    search,
    sortBy,
    fetchBoards,
    createBoard,
    deleteBoards,
    clear,
  }
})
