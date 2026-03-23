import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Board, Column, Card } from '@/types/domain'
import type { MemberResponse } from '@/types/api'
import * as boardsApi from '@/api/boards'
import { ApiError } from '@/api/client'
import { generatePosition } from '@/utils/lexorank'

export interface MemberWithProfile {
  userId: string
  role: 'owner' | 'member'
  name: string
  email: string
}

export const useBoardStore = defineStore('board', () => {
  const board = ref<Board | null>(null)
  const columns = ref<Column[]>([])
  const members = ref<MemberResponse[]>([])
  const memberProfiles = ref<Map<string, MemberWithProfile>>(new Map())
  const loading = ref(false)
  const error = ref<string | null>(null)

  const boardId = computed(() => board.value?.id)

  async function fetchBoard(id: string): Promise<void> {
    try {
      loading.value = true
      error.value = null

      const result = await boardsApi.getBoard(id)
      board.value = result.board
      columns.value = result.columns

      // Загружаем участников доски (профили уже включены в ответ)
      members.value = await boardsApi.getMembers(id)
      buildMemberProfiles()

      // Загружаем карточки для каждой колонки
      await Promise.all(
        columns.value.map(async (column) => {
          const cards = await boardsApi.getCards(column.id, id)
          // Lexorank: строковая сортировка ("a" < "am" < "b" < "bm" < "c")
          column.cards = cards.sort((a, b) => a.position.localeCompare(b.position))
        }),
      )
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка загрузки доски'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateBoardInfo(title: string, description: string): Promise<void> {
    if (!boardId.value) return

    try {
      error.value = null
      const updated = await boardsApi.updateBoard(boardId.value, { title, description })
      board.value = updated
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка обновления доски'
      throw err
    }
  }

  async function createColumn(title: string): Promise<void> {
    if (!boardId.value) return

    try {
      error.value = null
      const position = columns.value.length
      const column = await boardsApi.createColumn(boardId.value, { title, position })

      columns.value.push({
        ...column,
        cards: [],
      })
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка создания колонки'
      throw err
    }
  }

  async function updateColumn(columnId: string, title: string): Promise<void> {
    try {
      error.value = null
      await boardsApi.updateColumn(columnId, { title })

      const column = columns.value.find((c) => c.id === columnId)
      if (column) {
        column.title = title
      }
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка обновления колонки'
      throw err
    }
  }

  async function deleteColumn(columnId: string): Promise<void> {
    if (!boardId.value) return
    try {
      error.value = null
      await boardsApi.deleteColumn(columnId, boardId.value)
      columns.value = columns.value.filter((c) => c.id !== columnId)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка удаления колонки'
      throw err
    }
  }

  async function createCard(columnId: string, title: string, description: string): Promise<void> {
    if (!boardId.value) return

    try {
      error.value = null

      const column = columns.value.find((c) => c.id === columnId)
      if (!column) return

      // Генерируем позицию для конца списка
      const lastCard = column.cards[column.cards.length - 1]
      const position = lastCard ? lastCard.position + 'm' : 'a'

      const card = await boardsApi.createCard(columnId, {
        board_id: boardId.value,
        title,
        description,
        position,
      })

      column.cards.push(card)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка создания карточки'
      throw err
    }
  }

  async function updateCard(cardId: string, title: string, description: string): Promise<void> {
    if (!boardId.value) return
    try {
      error.value = null
      // Сохраняем текущий assignee_id чтобы не обнулить его при обновлении title/description
      let currentAssignee: string | undefined
      for (const col of columns.value) {
        const card = col.cards.find(c => c.id === cardId)
        if (card) { currentAssignee = card.assigneeId; break }
      }
      const updated = await boardsApi.updateCard(cardId, {
        board_id: boardId.value,
        title,
        description,
        assignee_id: currentAssignee,
      })

      // Обновляем карточку в store
      for (const column of columns.value) {
        const cardIndex = column.cards.findIndex((c) => c.id === cardId)
        if (cardIndex !== -1) {
          column.cards[cardIndex] = updated
          break
        }
      }
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка обновления карточки'
      throw err
    }
  }

  async function deleteCards(cardIds: string[]): Promise<void> {
    if (!boardId.value || cardIds.length === 0) return

    try {
      error.value = null

      await boardsApi.deleteCards({
        card_ids: cardIds,
        board_id: boardId.value,
      })

      const idsSet = new Set(cardIds)
      for (const column of columns.value) {
        column.cards = column.cards.filter((c) => !idsSet.has(c.id))
      }
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка удаления карточек'
      throw err
    }
  }

  // Move card - vuedraggable УЖЕ обновил UI, просто сохраняем на бэк
  async function moveCard(
    cardId: string,
    fromColumnId: string,
    toColumnId: string,
    newIndex: number,
  ): Promise<void> {
    if (!boardId.value) return

    // Snapshot для rollback при ошибке (JSON клонирование для избежания проблем с Vue Proxy)
    const snapshot = JSON.parse(JSON.stringify(columns.value))

    try {
      error.value = null

      // Находим карточку в новой позиции (vuedraggable уже переместил)
      const toColumn = columns.value.find((c) => c.id === toColumnId)
      if (!toColumn) return

      const card = toColumn.cards.find((c) => c.id === cardId)
      if (!card) return

      // Генерируем lexorank позицию между соседними карточками
      const prevCard = toColumn.cards[newIndex - 1]
      const nextCard = toColumn.cards[newIndex + 1]
      const position = generatePosition(
        prevCard?.position,
        nextCard?.position,
      )

      // Обновляем позицию и columnId
      card.position = position
      card.columnId = toColumnId

      // Отправляем на бэк
      await boardsApi.moveCard(cardId, {
        board_id: boardId.value,
        from_column_id: fromColumnId,
        to_column_id: toColumnId,
        position: position,
        version: card.version,
      })
    } catch (err) {
      // Rollback при ошибке
      columns.value = snapshot
      error.value = err instanceof ApiError ? err.message : 'Ошибка перемещения карточки'
      throw err
    }
  }

  async function assignCard(cardId: string, assigneeId: string): Promise<void> {
    if (!boardId.value) return
    try {
      error.value = null
      const updated = await boardsApi.assignCard(cardId, boardId.value, assigneeId)
      updateCardInStore(cardId, updated)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка назначения карточки'
      throw err
    }
  }

  async function unassignCard(cardId: string): Promise<void> {
    if (!boardId.value) return
    try {
      error.value = null
      const updated = await boardsApi.unassignCard(cardId, boardId.value)
      updateCardInStore(cardId, updated)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка снятия назначения'
      throw err
    }
  }

  function updateCardInStore(cardId: string, updated: Card): void {
    for (const column of columns.value) {
      const idx = column.cards.findIndex((c) => c.id === cardId)
      if (idx !== -1) {
        column.cards[idx] = updated
        break
      }
    }
  }

  // Строим кеш профилей из обогащённого ответа API
  function buildMemberProfiles(): void {
    const profiles = new Map<string, MemberWithProfile>()
    for (const m of members.value) {
      profiles.set(m.user_id, {
        userId: m.user_id,
        role: m.role,
        name: m.name || m.email || m.user_id.slice(0, 8),
        email: m.email || '',
      })
    }
    memberProfiles.value = profiles
  }

  // Получить имя пользователя по ID (из кеша профилей)
  function getMemberName(userId: string): string {
    const profile = memberProfiles.value.get(userId)
    return profile?.name || userId.slice(0, 8)
  }

  function getMemberEmail(userId: string): string {
    const profile = memberProfiles.value.get(userId)
    return profile?.email || ''
  }

  function clear(): void {
    board.value = null
    columns.value = []
    members.value = []
    memberProfiles.value = new Map()
    error.value = null
  }

  return {
    board,
    columns,
    members,
    memberProfiles,
    loading,
    error,
    boardId,
    fetchBoard,
    updateBoardInfo,
    createColumn,
    updateColumn,
    deleteColumn,
    getMemberName,
    getMemberEmail,
    createCard,
    updateCard,
    deleteCards,
    moveCard,
    assignCard,
    unassignCard,
    clear,
  }
})
