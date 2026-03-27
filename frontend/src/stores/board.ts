import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Board, Column, Card, Label, BoardSettings, UserLabel, CardTemplate } from '@/types/domain'
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
  const labels = ref<Label[]>([])
  const boardSettings = ref<BoardSettings | null>(null)
  const globalLabels = ref<UserLabel[]>([])
  const cardTemplates = ref<CardTemplate[]>([])
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

      // Загружаем метки доски (board + global) и настройки
      await Promise.all([fetchAvailableLabels(id), fetchCardTemplates(id)])

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
    if (!boardId.value || !board.value) return

    try {
      error.value = null
      const updated = await boardsApi.updateBoard(boardId.value, { title, description, version: board.value.version })
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

  async function fetchLabels(id: string): Promise<void> {
    try {
      labels.value = await boardsApi.listLabels(id)
    } catch (err) {
      console.error('Failed to load labels:', err)
      labels.value = []
    }
  }

  async function fetchAvailableLabels(id: string): Promise<void> {
    try {
      const result = await boardsApi.getAvailableLabels(id)
      labels.value = result.boardLabels
      globalLabels.value = result.globalLabels
      boardSettings.value = {
        boardId: id,
        useBoardLabelsOnly: result.useBoardLabelsOnly,
      }
    } catch {
      // Fallback: load only board labels if available-labels endpoint not ready
      try {
        labels.value = await boardsApi.listLabels(id)
      } catch (err2) {
        console.error('Failed to load labels:', err2)
        labels.value = []
      }
      globalLabels.value = []
      boardSettings.value = { boardId: id, useBoardLabelsOnly: false }
    }
  }

  async function fetchCardTemplates(id: string): Promise<void> {
    try {
      cardTemplates.value = await boardsApi.listCardTemplates(id)
    } catch (err) {
      console.error('Failed to load card templates:', err)
      cardTemplates.value = []
    }
  }

  const allAvailableLabels = computed(() => {
    if (boardSettings.value?.useBoardLabelsOnly) {
      return labels.value.map(l => ({ ...l, isGlobal: false as const }))
    }
    return [
      ...labels.value.map(l => ({ ...l, isGlobal: false as const })),
      ...globalLabels.value.map(l => ({
        id: l.id,
        boardId: '',
        name: l.name,
        color: l.color,
        createdAt: l.createdAt,
        isGlobal: true as const,
      })),
    ]
  })

  async function createBoardLabel(name: string, color: string): Promise<void> {
    if (!boardId.value) return
    try {
      const label = await boardsApi.createLabel(boardId.value, name, color)
      labels.value.push(label)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка создания метки'
      throw err
    }
  }

  async function updateBoardLabel(labelId: string, name: string, color: string): Promise<void> {
    if (!boardId.value) return
    try {
      const updated = await boardsApi.updateLabel(boardId.value, labelId, name, color)
      const idx = labels.value.findIndex(l => l.id === labelId)
      if (idx !== -1) labels.value[idx] = updated
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка обновления метки'
      throw err
    }
  }

  async function deleteBoardLabel(labelId: string): Promise<void> {
    if (!boardId.value) return
    try {
      await boardsApi.deleteLabel(boardId.value, labelId)
      labels.value = labels.value.filter(l => l.id !== labelId)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка удаления метки'
      throw err
    }
  }

  async function saveBoardSettings(useBoardLabelsOnly: boolean): Promise<void> {
    if (!boardId.value) return
    try {
      const updated = await boardsApi.updateBoardSettings(boardId.value, useBoardLabelsOnly)
      boardSettings.value = updated
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка сохранения настроек'
      throw err
    }
  }

  async function createCard(
    columnId: string,
    title: string,
    description: string,
    opts?: { dueDate?: string; priority?: string; taskType?: string },
  ): Promise<void> {
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
        due_date: opts?.dueDate ? new Date(opts.dueDate).toISOString() : undefined,
        priority: opts?.priority,
        task_type: opts?.taskType,
      })

      column.cards.push(card)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка создания карточки'
      throw err
    }
  }

  async function updateCard(
    cardId: string,
    title: string,
    description: string,
    opts?: { dueDate?: string; priority?: string; taskType?: string },
  ): Promise<void> {
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
        due_date: opts?.dueDate ? new Date(opts.dueDate).toISOString() : undefined,
        priority: opts?.priority,
        task_type: opts?.taskType,
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
    labels.value = []
    boardSettings.value = null
    globalLabels.value = []
    cardTemplates.value = []
    error.value = null
  }

  return {
    board,
    columns,
    members,
    memberProfiles,
    labels,
    boardSettings,
    globalLabels,
    cardTemplates,
    allAvailableLabels,
    loading,
    error,
    boardId,
    fetchBoard,
    fetchLabels,
    fetchAvailableLabels,
    fetchCardTemplates,
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
    createBoardLabel,
    updateBoardLabel,
    deleteBoardLabel,
    saveBoardSettings,
    clear,
  }
})
