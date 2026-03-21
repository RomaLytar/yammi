<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBoardStore } from '@/stores/board'
import { useAuthStore } from '@/stores/auth'
import { useRealtimeConnection } from '@/composables/useRealtimeBoard'
import { registerHandler, unregisterHandler } from '@/services/realtimeService'
import type { Card } from '@/types/domain'
import type {
  CardCreatedData, CardUpdatedData, CardDeletedData, CardMovedData,
  ColumnCreatedData, ColumnUpdatedData, ColumnDeletedData,
  BoardUpdatedData, BoardDeletedData, MemberRemovedData,
} from '@/types/events'
import BoardColumn from '@/components/board/BoardColumn.vue'
import CreateColumnModal from '@/components/board/CreateColumnModal.vue'
import CreateCardModal from '@/components/board/CreateCardModal.vue'
import EditCardModal from '@/components/board/EditCardModal.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

const route = useRoute()
const router = useRouter()
const boardStore = useBoardStore()
const authStore = useAuthStore()
const { subscribeBoard, unsubscribeBoard } = useRealtimeConnection()

const isOwner = computed(() => boardStore.board?.ownerId === authStore.userId)
const currentUserId = computed(() => authStore.userId || '')

const showCreateColumnModal = ref(false)
const showCreateCardModal = ref(false)
const showEditCardModal = ref(false)
const showConfirmDeleteColumn = ref(false)
const showBulkDeleteCards = ref(false)
const activeColumnId = ref<string | null>(null)
const activeCard = ref<Card | null>(null)
const pendingDeleteColumnId = ref<string | null>(null)

// Bulk card select
const cardSelectMode = ref(false)
const selectedCardIds = ref<Set<string>>(new Set())
const selectedCardCount = computed(() => selectedCardIds.value.size)

function canDeleteCard(card: Card): boolean {
  return isOwner.value || card.creatorId === currentUserId.value
}

function toggleCardSelectMode() {
  cardSelectMode.value = !cardSelectMode.value
  selectedCardIds.value = new Set()
}

function toggleCardSelect(cardId: string) {
  const s = new Set(selectedCardIds.value)
  if (s.has(cardId)) s.delete(cardId)
  else s.add(cardId)
  selectedCardIds.value = s
}

async function handleBulkDeleteCards() {
  if (selectedCardIds.value.size === 0) return
  try {
    await boardStore.deleteCards([...selectedCardIds.value])
    selectedCardIds.value = new Set()
    cardSelectMode.value = false
  } catch (err) {
    console.error('Failed to bulk delete cards:', err)
  } finally {
    showBulkDeleteCards.value = false
  }
}

// --- Real-time event handlers ---

function onCardCreated(data: unknown) {
  const d = data as CardCreatedData
  if (d.actor_id === authStore.userId) return
  const column = boardStore.columns.find(c => c.id === d.column_id)
  if (!column) return
  // Avoid duplicates
  if (column.cards.some(c => c.id === d.card_id)) return
  column.cards.push({
    id: d.card_id,
    title: d.title,
    description: d.description || '',
    position: d.position,
    columnId: d.column_id,
    creatorId: d.actor_id,
    version: 1,
    createdAt: new Date().toISOString(),
  })
  column.cards.sort((a, b) => a.position.localeCompare(b.position))
}

function onCardUpdated(data: unknown) {
  const d = data as CardUpdatedData
  if (d.actor_id === authStore.userId) return
  for (const column of boardStore.columns) {
    const card = column.cards.find(c => c.id === d.card_id)
    if (card) {
      card.title = d.title
      card.description = d.description
      if (d.assignee_id !== undefined) card.assigneeId = d.assignee_id
      break
    }
  }
}

function onCardDeleted(data: unknown) {
  const d = data as CardDeletedData
  if (d.actor_id === authStore.userId) return
  for (const column of boardStore.columns) {
    const idx = column.cards.findIndex(c => c.id === d.card_id)
    if (idx !== -1) {
      column.cards.splice(idx, 1)
      break
    }
  }
}

function onCardMoved(data: unknown) {
  const d = data as CardMovedData
  if (d.actor_id === authStore.userId) return
  const fromColumn = boardStore.columns.find(c => c.id === d.from_column_id)
  const toColumn = boardStore.columns.find(c => c.id === d.to_column_id)
  if (!fromColumn || !toColumn) return

  const cardIndex = fromColumn.cards.findIndex(c => c.id === d.card_id)
  if (cardIndex === -1) return

  const [card] = fromColumn.cards.splice(cardIndex, 1)
  card.position = d.new_position
  card.columnId = d.to_column_id
  toColumn.cards.push(card)
  toColumn.cards.sort((a, b) => a.position.localeCompare(b.position))
}

function onColumnCreated(data: unknown) {
  const d = data as ColumnCreatedData
  if (d.actor_id === authStore.userId) return
  if (boardStore.columns.some(c => c.id === d.column_id)) return
  boardStore.columns.push({
    id: d.column_id,
    title: d.title,
    position: d.position,
    cards: [],
  })
  boardStore.columns.sort((a, b) => a.position - b.position)
}

function onColumnUpdated(data: unknown) {
  const d = data as ColumnUpdatedData
  if (d.actor_id === authStore.userId) return
  const column = boardStore.columns.find(c => c.id === d.column_id)
  if (column) column.title = d.title
}

function onColumnDeleted(data: unknown) {
  const d = data as ColumnDeletedData
  if (d.actor_id === authStore.userId) return
  boardStore.columns = boardStore.columns.filter(c => c.id !== d.column_id)
}

function onBoardUpdated(data: unknown) {
  const d = data as BoardUpdatedData
  if (d.actor_id === authStore.userId) return
  if (boardStore.board) {
    boardStore.board.title = d.title
    boardStore.board.description = d.description
  }
}

function onBoardDeleted(data: unknown) {
  const d = data as BoardDeletedData
  if (boardStore.board?.id === d.board_id) {
    router.push('/boards')
  }
}

function onMemberRemoved(data: unknown) {
  const d = data as MemberRemovedData
  if (d.user_id === authStore.userId && d.board_id === boardStore.board?.id) {
    router.push('/boards')
  }
}

const realtimeHandlers: Array<[string, (data: unknown) => void]> = [
  ['card.created', onCardCreated],
  ['card.updated', onCardUpdated],
  ['card.deleted', onCardDeleted],
  ['card.moved', onCardMoved],
  ['column.created', onColumnCreated],
  ['column.updated', onColumnUpdated],
  ['column.deleted', onColumnDeleted],
  ['board.updated', onBoardUpdated],
  ['board.deleted', onBoardDeleted],
  ['member.removed', onMemberRemoved],
]

onMounted(async () => {
  const boardId = route.params.boardId as string
  try {
    await boardStore.fetchBoard(boardId)
  } catch (error) {
    console.error('Failed to load board:', error)
    router.push('/boards')
    return
  }

  // Subscribe to board updates via WebSocket
  subscribeBoard(boardId)

  // Register real-time handlers
  for (const [event, handler] of realtimeHandlers) {
    registerHandler(event, handler)
  }
})

onUnmounted(() => {
  const boardId = route.params.boardId as string

  // Unsubscribe from board updates
  if (boardId) unsubscribeBoard(boardId)

  // Unregister real-time handlers
  for (const [event, handler] of realtimeHandlers) {
    unregisterHandler(event, handler)
  }

  boardStore.clear()
})

async function handleCreateColumn(title: string) {
  try {
    await boardStore.createColumn(title)
    showCreateColumnModal.value = false
  } catch (error) {
    console.error('Failed to create column:', error)
  }
}

async function handleUpdateColumn(columnId: string, title: string) {
  try {
    await boardStore.updateColumn(columnId, title)
  } catch (error) {
    console.error('Failed to update column:', error)
  }
}

function handleDeleteColumn(columnId: string) {
  pendingDeleteColumnId.value = columnId
  showConfirmDeleteColumn.value = true
}

async function confirmDeleteColumn() {
  if (!pendingDeleteColumnId.value) return

  try {
    await boardStore.deleteColumn(pendingDeleteColumnId.value)
    showConfirmDeleteColumn.value = false
    pendingDeleteColumnId.value = null
  } catch (error) {
    console.error('Failed to delete column:', error)
  }
}

function cancelDeleteColumn() {
  showConfirmDeleteColumn.value = false
  pendingDeleteColumnId.value = null
}

function handleAddCard(columnId: string) {
  activeColumnId.value = columnId
  showCreateCardModal.value = true
}

async function handleCreateCard(data: { title: string; description: string }) {
  if (!activeColumnId.value) return

  try {
    await boardStore.createCard(activeColumnId.value, data.title, data.description)
    showCreateCardModal.value = false
    activeColumnId.value = null
  } catch (error) {
    console.error('Failed to create card:', error)
  }
}

function handleCardClick(card: Card) {
  activeCard.value = card
  showEditCardModal.value = true
}

async function handleUpdateCard(data: { title: string; description: string }) {
  if (!activeCard.value) return

  try {
    await boardStore.updateCard(activeCard.value.id, data.title, data.description)
    showEditCardModal.value = false
    activeCard.value = null
  } catch (error) {
    console.error('Failed to update card:', error)
  }
}

async function handleDeleteCard(cardId?: string) {
  const id = cardId || activeCard.value?.id
  if (!id) return

  try {
    await boardStore.deleteCards([id])

    if (showEditCardModal.value) {
      showEditCardModal.value = false
      activeCard.value = null
    }
  } catch (error) {
    console.error('Failed to delete card:', error)
  }
}

async function handleCardMove(event: { cardId: string; fromColumnId: string; toColumnId: string; newIndex: number }) {
  try {
    await boardStore.moveCard(event.cardId, event.fromColumnId, event.toColumnId, event.newIndex)
  } catch (error) {
    console.error('Failed to move card:', error)
  }
}

function closeCreateCardModal() {
  showCreateCardModal.value = false
  activeColumnId.value = null
}

function closeEditCardModal() {
  showEditCardModal.value = false
  activeCard.value = null
}
</script>

<template>
  <div class="board-page">
    <div v-if="boardStore.loading" class="board-page__loading">
      <BaseSpinner />
    </div>

    <div v-else-if="boardStore.error" class="board-page__error">
      <p>{{ boardStore.error }}</p>
      <BaseButton variant="secondary" @click="router.push('/boards')">
        Вернуться к доскам
      </BaseButton>
    </div>

    <template v-else-if="boardStore.board">
      <div class="board-page__header">
        <div>
          <h1 class="board-page__title">{{ boardStore.board.title }}</h1>
          <p v-if="boardStore.board.description" class="board-page__description">
            {{ boardStore.board.description }}
          </p>
        </div>
        <div class="board-page__actions">
          <button
            class="select-toggle"
            :class="{ 'select-toggle--active': cardSelectMode }"
            @click="toggleCardSelectMode"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="9 11 12 14 22 4" />
              <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />
            </svg>
            {{ cardSelectMode ? 'Отмена' : 'Выбрать' }}
          </button>
          <button
            v-if="cardSelectMode && selectedCardCount > 0"
            class="bulk-delete-btn"
            @click="showBulkDeleteCards = true"
          >
            Удалить ({{ selectedCardCount }})
          </button>
          <BaseButton v-if="isOwner && !cardSelectMode" @click="showCreateColumnModal = true">
            + Добавить колонку
          </BaseButton>
        </div>
      </div>

      <div class="board-page__content">
        <div class="board-columns">
          <BoardColumn
            v-for="column in boardStore.columns"
            :key="column.id"
            :column="column"
            :is-owner="isOwner"
            :current-user-id="currentUserId"
            :select-mode="cardSelectMode"
            :selected-ids="selectedCardIds"
            @add-card="handleAddCard(column.id)"
            @card-click="handleCardClick"
            @card-delete="handleDeleteCard"
            @card-move="handleCardMove"
            @card-toggle-select="toggleCardSelect"
            @update-title="(title) => handleUpdateColumn(column.id, title)"
            @delete="handleDeleteColumn(column.id)"
          />

          <div v-if="isOwner" class="board-columns__placeholder">
            <button
              class="add-column-button"
              @click="showCreateColumnModal = true"
            >
              + Добавить колонку
            </button>
          </div>
        </div>
      </div>
    </template>

    <CreateColumnModal
      v-if="showCreateColumnModal"
      @close="showCreateColumnModal = false"
      @create="handleCreateColumn"
    />

    <CreateCardModal
      v-if="showCreateCardModal"
      @close="closeCreateCardModal"
      @create="handleCreateCard"
    />

    <EditCardModal
      v-if="showEditCardModal && activeCard"
      :card="activeCard"
      :can-delete="isOwner || activeCard.creatorId === currentUserId"
      @close="closeEditCardModal"
      @update="handleUpdateCard"
      @delete="handleDeleteCard()"
    />

    <ConfirmModal
      v-if="showBulkDeleteCards"
      title="Удалить выбранные карточки"
      :message="`Удалить ${selectedCardCount} карточек? Это действие нельзя отменить.`"
      confirm-text="Удалить"
      variant="danger"
      @confirm="handleBulkDeleteCards"
      @cancel="showBulkDeleteCards = false"
    />

    <ConfirmModal
      v-if="showConfirmDeleteColumn"
      title="Удалить колонку?"
      message="Вы уверены, что хотите удалить эту колонку? Все карточки в ней также будут удалены. Это действие нельзя отменить."
      confirm-text="Удалить"
      cancel-text="Отмена"
      variant="danger"
      @confirm="confirmDeleteColumn"
      @cancel="cancelDeleteColumn"
    />
  </div>
</template>

<style scoped>
.board-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--gradient-auth-bg);
  overflow: hidden;
  position: relative;
}

.board-page::before {
  content: '';
  position: absolute;
  width: 800px;
  height: 800px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.06) 0%, transparent 70%);
  top: -200px;
  right: -200px;
  pointer-events: none;
  z-index: 0;
}

.board-page::after {
  content: '';
  position: absolute;
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.05) 0%, transparent 70%);
  bottom: -150px;
  left: -150px;
  pointer-events: none;
  z-index: 0;
}

.board-page__loading,
.board-page__error {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  color: white;
}

.board-page__error {
  flex-direction: column;
  gap: 16px;
}

.board-page__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(139, 92, 246, 0.1);
  box-shadow: 0 1px 3px rgba(139, 92, 246, 0.1);
  position: relative;
  z-index: 1;
}

.board-page__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.select-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border: 1px solid rgba(255,255,255,0.4);
  border-radius: 8px;
  font-size: 13px;
  color: #6b7280;
  background: rgba(255,255,255,0.5);
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}
.select-toggle:hover { border-color: #9ca3af; color: #374151; }
.select-toggle--active { border-color: #3b82f6; color: #3b82f6; background: rgba(219,234,254,0.6); }

.bulk-delete-btn {
  padding: 7px 14px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  background: #dc2626;
  color: white;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.15s;
}
.bulk-delete-btn:hover { background: #b91c1c; }

.board-page__title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  background: var(--gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.board-page__description {
  margin: 4px 0 0 0;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.board-page__content {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 20px 0;
  position: relative;
  z-index: 1;
}

.board-columns {
  display: flex;
  gap: 20px;
  min-height: 100%;
  align-items: flex-start;
  padding: 0 24px;
}

.board-columns__placeholder {
  min-width: 300px;
  flex-shrink: 0;
}

.add-column-button {
  width: 100%;
  padding: 14px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  border: 2px dashed rgba(99, 102, 241, 0.3);
  border-radius: 16px;
  color: var(--color-primary);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.add-column-button:hover {
  background: rgba(255, 255, 255, 0.95);
  border-color: var(--color-primary);
  transform: translateY(-2px);
  box-shadow: 0 4px 6px -1px rgba(99, 102, 241, 0.1), 0 2px 4px -1px rgba(99, 102, 241, 0.06);
}

.add-column-button:active {
  transform: translateY(0);
}

/* Кастомный scrollbar */
.board-page__content::-webkit-scrollbar {
  height: 12px;
}

.board-page__content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  margin: 0 24px;
}

.board-page__content::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
  border-radius: 6px;
  transition: background 0.2s;
}

.board-page__content::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.4);
}
</style>
