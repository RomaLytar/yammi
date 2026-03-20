<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBoardStore } from '@/stores/board'
import type { Card } from '@/types/domain'
import BoardColumn from '@/components/board/BoardColumn.vue'
import CreateColumnModal from '@/components/board/CreateColumnModal.vue'
import CreateCardModal from '@/components/board/CreateCardModal.vue'
import EditCardModal from '@/components/board/EditCardModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

const route = useRoute()
const router = useRouter()
const boardStore = useBoardStore()

const showCreateColumnModal = ref(false)
const showCreateCardModal = ref(false)
const showEditCardModal = ref(false)
const activeColumnId = ref<string | null>(null)
const activeCard = ref<Card | null>(null)

onMounted(async () => {
  const boardId = route.params.boardId as string
  try {
    await boardStore.fetchBoard(boardId)
  } catch (error) {
    console.error('Failed to load board:', error)
    router.push('/boards')
  }
})

onUnmounted(() => {
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

async function handleDeleteColumn(columnId: string) {
  if (confirm('Удалить колонку и все карточки в ней?')) {
    try {
      await boardStore.deleteColumn(columnId)
    } catch (error) {
      console.error('Failed to delete column:', error)
    }
  }
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
    await boardStore.deleteCard(id)
    if (showEditCardModal.value) {
      showEditCardModal.value = false
      activeCard.value = null
    }
  } catch (error) {
    console.error('Failed to delete card:', error)
  }
}

async function handleCardMove(event: { cardId: string; fromColumnId: string; toColumnId: string; newIndex: number }) {
  console.log('[BoardPage] handleCardMove:', event)
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
        <BaseButton @click="showCreateColumnModal = true">
          + Добавить колонку
        </BaseButton>
      </div>

      <div class="board-page__content">
        <div class="board-columns">
          <BoardColumn
            v-for="column in boardStore.columns"
            :key="column.id"
            :column="column"
            @add-card="handleAddCard(column.id)"
            @card-click="handleCardClick"
            @card-delete="handleDeleteCard"
            @card-move="handleCardMove"
            @update-title="(title) => handleUpdateColumn(column.id, title)"
            @delete="handleDeleteColumn(column.id)"
          />

          <div class="board-columns__placeholder">
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
      @close="closeEditCardModal"
      @update="handleUpdateCard"
      @delete="handleDeleteCard()"
    />
  </div>
</template>

<style scoped>
.board-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
  align-items: flex-start;
  padding: 24px 32px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.board-page__title {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  color: white;
}

.board-page__description {
  margin: 4px 0 0 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

.board-page__content {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 24px 32px;
}

.board-columns {
  display: flex;
  gap: 16px;
  min-height: 100%;
  align-items: flex-start;
}

.board-columns__placeholder {
  min-width: 280px;
}

.add-column-button {
  width: 100%;
  padding: 12px;
  background: rgba(255, 255, 255, 0.1);
  border: 2px dashed rgba(255, 255, 255, 0.3);
  border-radius: 12px;
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.add-column-button:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.5);
}
</style>
