<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useBoardsStore } from '@/stores/boards'
import CreateBoardModal from '@/components/board/CreateBoardModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

const router = useRouter()
const boardsStore = useBoardsStore()

const showCreateModal = ref(false)
const creatingBoard = ref(false)

onMounted(async () => {
  await boardsStore.fetchBoards(true)
})

async function handleCreateBoard(data: { title: string; description: string }) {
  try {
    creatingBoard.value = true
    const board = await boardsStore.createBoard(data.title, data.description)
    showCreateModal.value = false
    router.push(`/boards/${board.id}`)
  } catch (error) {
    console.error('Failed to create board:', error)
  } finally {
    creatingBoard.value = false
  }
}

function openBoard(boardId: string) {
  router.push(`/boards/${boardId}`)
}

async function loadMore() {
  if (!boardsStore.loading && boardsStore.hasMore) {
    await boardsStore.fetchBoards(false)
  }
}
</script>

<template>
  <div class="board-list-page">
    <div class="board-list-header">
      <h1>Мои доски</h1>
      <BaseButton @click="showCreateModal = true">
        + Создать доску
      </BaseButton>
    </div>

    <div v-if="boardsStore.loading && boardsStore.boards.length === 0" class="board-list-empty">
      <BaseSpinner />
    </div>

    <div v-else-if="boardsStore.error" class="board-list-error">
      <p>{{ boardsStore.error }}</p>
      <BaseButton variant="secondary" @click="boardsStore.fetchBoards(true)">
        Повторить
      </BaseButton>
    </div>

    <div v-else-if="boardsStore.boards.length === 0" class="board-list-empty">
      <div class="empty-state">
        <div class="empty-state__icon">📋</div>
        <h2>У вас пока нет досок</h2>
        <p>Создайте первую доску для управления задачами</p>
        <BaseButton @click="showCreateModal = true">
          Создать доску
        </BaseButton>
      </div>
    </div>

    <div v-else class="board-list">
      <div
        v-for="board in boardsStore.boards"
        :key="board.id"
        class="board-item"
        @click="openBoard(board.id)"
      >
        <div class="board-item__header">
          <h3>{{ board.title }}</h3>
        </div>
        <p v-if="board.description" class="board-item__description">
          {{ board.description }}
        </p>
        <div class="board-item__footer">
          <span class="board-item__date">
            {{ new Date(board.createdAt).toLocaleDateString('ru-RU') }}
          </span>
        </div>
      </div>

      <div v-if="boardsStore.hasMore" class="board-list-footer">
        <BaseButton
          variant="secondary"
          :loading="boardsStore.loading"
          @click="loadMore"
        >
          Загрузить ещё
        </BaseButton>
      </div>
    </div>

    <CreateBoardModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @create="handleCreateBoard"
    />
  </div>
</template>

<style scoped>
.board-list-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.board-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.board-list-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
  color: var(--color-text-primary, #111827);
}

.board-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.board-item {
  background: var(--color-surface, #fff);
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.board-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
  border-color: var(--color-primary, #3b82f6);
}

.board-item__header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-primary, #111827);
}

.board-item__description {
  margin: 12px 0 0 0;
  font-size: 14px;
  color: var(--color-text-secondary, #6b7280);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.board-item__footer {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border, #e5e7eb);
}

.board-item__date {
  font-size: 12px;
  color: var(--color-text-tertiary, #9ca3af);
}

.board-list-empty,
.board-list-error {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}

.board-list-error {
  flex-direction: column;
  gap: 16px;
}

.empty-state {
  text-align: center;
  max-width: 400px;
}

.empty-state__icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-state h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--color-text-primary, #111827);
}

.empty-state p {
  margin: 0 0 24px 0;
  font-size: 16px;
  color: var(--color-text-secondary, #6b7280);
}

.board-list-footer {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
