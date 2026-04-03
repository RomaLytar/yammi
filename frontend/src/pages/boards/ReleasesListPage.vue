<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useReleasesStore } from '@/stores/releases'
import { useBoardStore } from '@/stores/board'
import { useAuthStore } from '@/stores/auth'
import type { Release } from '@/types/domain'
import ReleaseStatusBadge from '@/components/board/ReleaseStatusBadge.vue'
import CreateReleaseModal from '@/components/board/CreateReleaseModal.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

const route = useRoute()
const router = useRouter()
const releasesStore = useReleasesStore()
const boardStore = useBoardStore()
const authStore = useAuthStore()

const boardId = computed(() => route.params.boardId as string)
const isOwner = computed(() => boardStore.board?.ownerId === authStore.userId)

const showCreateModal = ref(false)
const showConfirmDelete = ref(false)
const showConfirmComplete = ref(false)
const pendingReleaseId = ref<string | null>(null)

onMounted(async () => {
  if (!boardStore.board) {
    try {
      await boardStore.fetchBoard(boardId.value)
    } catch {
      router.push('/boards')
      return
    }
  }
  await releasesStore.fetchReleases(boardId.value)
})

async function handleCreate(data: { name: string; description: string }) {
  try {
    await releasesStore.createRelease(boardId.value, data.name, data.description)
    showCreateModal.value = false
  } catch (err) {
    console.error('Failed to create release:', err)
  }
}

async function handleStart(releaseId: string) {
  try {
    await releasesStore.startRelease(boardId.value, releaseId)
  } catch (err) {
    console.error('Failed to start release:', err)
  }
}

function confirmComplete(releaseId: string) {
  pendingReleaseId.value = releaseId
  showConfirmComplete.value = true
}

async function handleComplete() {
  if (!pendingReleaseId.value) return
  try {
    await releasesStore.completeRelease(boardId.value, pendingReleaseId.value)
  } catch (err) {
    console.error('Failed to complete release:', err)
  } finally {
    showConfirmComplete.value = false
    pendingReleaseId.value = null
  }
}

function confirmDelete(releaseId: string) {
  pendingReleaseId.value = releaseId
  showConfirmDelete.value = true
}

async function handleDelete() {
  if (!pendingReleaseId.value) return
  try {
    await releasesStore.deleteRelease(boardId.value, pendingReleaseId.value)
  } catch (err) {
    console.error('Failed to delete release:', err)
  } finally {
    showConfirmDelete.value = false
    pendingReleaseId.value = null
  }
}

function openRelease(release: Release) {
  router.push(`/boards/${boardId.value}/releases/${release.id}`)
}

function formatDate(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>

<template>
  <div class="releases-page">
    <div class="releases-page__header">
      <div>
        <h1 class="releases-page__title">Релизы</h1>
        <p class="releases-page__subtitle">{{ boardStore.board?.title }}</p>
      </div>
      <div class="releases-page__actions">
        <BaseButton v-if="isOwner" @click="showCreateModal = true">
          + Создать релиз
        </BaseButton>
      </div>
    </div>

    <div v-if="releasesStore.loading" class="releases-page__loading">
      <BaseSpinner />
    </div>

    <div v-else-if="releasesStore.releases.length === 0" class="releases-page__empty">
      <p>Релизов пока нет</p>
      <BaseButton v-if="isOwner" @click="showCreateModal = true">Создать первый релиз</BaseButton>
    </div>

    <div v-else class="releases-list">
      <div
        v-for="release in releasesStore.releases"
        :key="release.id"
        class="release-card"
        :class="{ 'release-card--active': release.status === 'active', 'release-card--completed': release.status === 'completed' }"
        @click="openRelease(release)"
      >
        <div class="release-card__header">
          <h3 class="release-card__name">{{ release.name }}</h3>
          <ReleaseStatusBadge :status="release.status" />
        </div>
        <p v-if="release.description" class="release-card__desc">{{ release.description }}</p>
        <div class="release-card__meta">
          <span v-if="release.startedAt">Начат: {{ formatDate(release.startedAt) }}</span>
          <span v-if="release.completedAt">Завершён: {{ formatDate(release.completedAt) }}</span>
          <span v-if="!release.startedAt">Создан: {{ formatDate(release.createdAt) }}</span>
        </div>
        <div v-if="isOwner" class="release-card__actions" @click.stop>
          <button
            v-if="release.status === 'draft'"
            class="release-card__action release-card__action--start"
            @click="handleStart(release.id)"
          >
            Запустить
          </button>
          <button
            v-if="release.status === 'active'"
            class="release-card__action release-card__action--complete"
            @click="confirmComplete(release.id)"
          >
            Завершить
          </button>
          <button
            v-if="release.status !== 'completed'"
            class="release-card__action release-card__action--delete"
            @click="confirmDelete(release.id)"
          >
            Удалить
          </button>
        </div>
      </div>
    </div>

    <CreateReleaseModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @create="handleCreate"
    />

    <ConfirmModal
      v-if="showConfirmComplete"
      title="Завершить релиз?"
      message="Незавершённые задачи (не в done-колонке) будут перемещены в бэклог."
      confirm-text="Завершить"
      @confirm="handleComplete"
      @cancel="showConfirmComplete = false"
    />

    <ConfirmModal
      v-if="showConfirmDelete"
      title="Удалить релиз?"
      message="Все задачи релиза будут перемещены в бэклог. Это действие нельзя отменить."
      confirm-text="Удалить"
      variant="danger"
      @confirm="handleDelete"
      @cancel="showConfirmDelete = false"
    />
  </div>
</template>

<style scoped>
.releases-page {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 56px);
}
.releases-page__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  max-width: 900px;
  width: 100%;
  margin-left: auto;
  margin-right: auto;
  padding: 24px 24px 0;
}
.releases-page__title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  background: var(--gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.releases-page__subtitle {
  margin: 4px 0 0 0;
  font-size: 14px;
  color: var(--color-text-secondary);
}
.releases-page__actions {
  display: flex;
  gap: 8px;
}
.releases-page__loading,
.releases-page__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: var(--color-text-secondary);
  gap: 16px;
}

.releases-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.release-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border-light);
  border-radius: 12px;
  padding: 16px 20px;
  cursor: pointer;
  transition: all 0.15s;
}
.release-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: var(--color-border);
}
.release-card--active {
  border-left: 4px solid #10b981;
}
.release-card--completed {
  opacity: 0.7;
}
.release-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.release-card__name {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
}
.release-card__desc {
  margin: 8px 0 0 0;
  font-size: 14px;
  color: var(--color-text-secondary);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.release-card__meta {
  display: flex;
  gap: 16px;
  margin-top: 12px;
  font-size: 12px;
  color: var(--color-text-tertiary);
}
.release-card__actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}
.release-card__action {
  padding: 5px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  background: var(--color-surface-alt);
  color: var(--color-text-secondary);
  transition: all 0.15s;
}
.release-card__action:hover {
  border-color: var(--color-text-tertiary);
  color: var(--color-text-primary);
}
.release-card__action--start {
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.3);
}
.release-card__action--start:hover {
  background: rgba(16, 185, 129, 0.08);
  border-color: #10b981;
}
.release-card__action--complete {
  color: var(--color-primary);
  border-color: var(--color-primary);
}
.release-card__action--complete:hover {
  background: var(--color-primary-light);
}
.release-card__action--delete {
  color: var(--color-danger, #dc2626);
  border-color: rgba(220, 38, 38, 0.2);
}
.release-card__action--delete:hover {
  background: var(--color-danger-soft, rgba(239, 68, 68, 0.06));
  border-color: var(--color-danger, #dc2626);
}
</style>
