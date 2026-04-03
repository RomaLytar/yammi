<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useReleasesStore } from '@/stores/releases'
import { useBoardStore } from '@/stores/board'
import { useAuthStore } from '@/stores/auth'
import ReleaseStatusBadge from '@/components/board/ReleaseStatusBadge.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

const route = useRoute()
const router = useRouter()
const releasesStore = useReleasesStore()
const boardStore = useBoardStore()
const authStore = useAuthStore()

const boardId = computed(() => route.params.boardId as string)
const releaseId = computed(() => route.params.releaseId as string)
const release = computed(() => releasesStore.currentRelease)
const isOwner = computed(() => boardStore.board?.ownerId === authStore.userId)
const isCompleted = computed(() => release.value?.status === 'completed')
const showConfirmRemove = ref(false)
const pendingRemoveCardId = ref<string | null>(null)

onMounted(async () => {
  if (!boardStore.board) {
    try {
      await boardStore.fetchBoard(boardId.value)
    } catch {
      router.push('/boards')
      return
    }
  }
  await Promise.all([
    releasesStore.fetchRelease(boardId.value, releaseId.value),
    releasesStore.fetchReleaseCards(boardId.value, releaseId.value),
  ])
})

function confirmRemoveCard(cardId: string) {
  pendingRemoveCardId.value = cardId
  showConfirmRemove.value = true
}

async function handleRemoveCard() {
  if (!pendingRemoveCardId.value) return
  try {
    await releasesStore.removeCard(boardId.value, releaseId.value, pendingRemoveCardId.value)
  } catch (err) {
    console.error('Failed to remove card:', err)
  } finally {
    showConfirmRemove.value = false
    pendingRemoveCardId.value = null
  }
}

function formatDate(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })
}

function getMemberName(userId: string): string {
  return boardStore.getMemberName(userId)
}
</script>

<template>
  <div class="release-detail">
    <div v-if="releasesStore.loading && !release" class="release-detail__loading">
      <BaseSpinner />
    </div>

    <template v-else-if="release">
      <div class="release-detail__header">
        <div>
          <div class="release-detail__title-row">
            <h1 class="release-detail__title">{{ release.name }}</h1>
            <ReleaseStatusBadge :status="release.status" />
          </div>
          <p v-if="release.description" class="release-detail__desc">{{ release.description }}</p>
          <div class="release-detail__meta">
            <span>Создал: {{ getMemberName(release.createdBy) }}</span>
            <span v-if="release.startedAt">Начат: {{ formatDate(release.startedAt) }}</span>
            <span v-if="release.completedAt">Завершён: {{ formatDate(release.completedAt) }}</span>
          </div>
        </div>
        <div class="release-detail__actions">
          <BaseButton variant="secondary" @click="router.push(`/boards/${boardId}/releases`)">
            Все релизы
          </BaseButton>
        </div>
      </div>

      <div class="release-detail__cards-header">
        <h2 class="release-detail__section-title">
          Задачи ({{ releasesStore.releaseCards.length }})
        </h2>
      </div>

      <div v-if="releasesStore.releaseCards.length === 0" class="release-detail__empty">
        В этом релизе пока нет задач
      </div>

      <div v-else class="release-cards">
        <div
          v-for="card in releasesStore.releaseCards"
          :key="card.id"
          class="release-task"
        >
          <div class="release-task__content">
            <h4 class="release-task__title">{{ card.title }}</h4>
            <div class="release-task__meta">
              <span class="release-task__type">{{ card.taskType }}</span>
              <span class="release-task__priority" :class="`release-task__priority--${card.priority}`">
                {{ card.priority }}
              </span>
              <span v-if="card.assigneeId" class="release-task__assignee">
                {{ getMemberName(card.assigneeId) }}
              </span>
            </div>
          </div>
          <button
            v-if="!isCompleted"
            class="release-task__remove"
            title="Убрать из релиза"
            @click="confirmRemoveCard(card.id)"
          >
            &times;
          </button>
        </div>
      </div>
    </template>

    <ConfirmModal
      v-if="showConfirmRemove"
      title="Убрать из релиза?"
      message="Задача будет перемещена в бэклог."
      confirm-text="Убрать"
      @confirm="handleRemoveCard"
      @cancel="showConfirmRemove = false"
    />
  </div>
</template>

<style scoped>
.release-detail {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
}
.release-detail__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
}
.release-detail__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  gap: 16px;
}
.release-detail__title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.release-detail__title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary);
}
.release-detail__desc {
  margin: 8px 0 0 0;
  font-size: 14px;
  color: var(--color-text-secondary);
  line-height: 1.5;
}
.release-detail__meta {
  display: flex;
  gap: 16px;
  margin-top: 8px;
  font-size: 12px;
  color: var(--color-text-tertiary);
}
.release-detail__actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.release-detail__cards-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.release-detail__section-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
}
.release-detail__empty {
  text-align: center;
  padding: 40px 0;
  color: var(--color-text-tertiary);
  font-size: 14px;
}
.release-cards {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.release-task {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--color-surface);
  border: 1px solid var(--color-border-light);
  border-radius: 10px;
  padding: 12px 16px;
  gap: 12px;
}
.release-task__content { flex: 1; min-width: 0; }
.release-task__title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
}
.release-task__meta {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-tertiary);
}
.release-task__priority--critical { color: #ef4444; }
.release-task__priority--high { color: #f59e0b; }
.release-task__priority--medium { color: #7c5cfc; }
.release-task__priority--low { color: #10b981; }
.release-task__remove {
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  font-size: 20px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.15s;
}
.release-task__remove:hover {
  background: var(--color-danger-soft);
  color: var(--color-danger);
}
</style>
