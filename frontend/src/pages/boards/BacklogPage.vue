<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useReleasesStore } from '@/stores/releases'
import { useBoardStore } from '@/stores/board'
import type { Card, Release } from '@/types/domain'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

const route = useRoute()
const router = useRouter()
const releasesStore = useReleasesStore()
const boardStore = useBoardStore()

const boardId = computed(() => route.params.boardId as string)
const showAssignDropdown = ref<string | null>(null)

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
    releasesStore.fetchBacklog(boardId.value),
    releasesStore.fetchReleases(boardId.value),
  ])
})

const draftAndActiveReleases = computed(() =>
  releasesStore.releases.filter(r => r.status !== 'completed')
)

async function assignToRelease(cardId: string, releaseId: string) {
  try {
    await releasesStore.assignCard(boardId.value, releaseId, cardId)
    showAssignDropdown.value = null
  } catch (err) {
    console.error('Failed to assign card to release:', err)
  }
}

function getReleaseName(releaseId: string): string {
  return releasesStore.releases.find(r => r.id === releaseId)?.name || ''
}
</script>

<template>
  <div class="backlog-page">
    <div class="backlog-page__header">
      <div>
        <h1 class="backlog-page__title">Бэклог</h1>
        <p class="backlog-page__subtitle">{{ boardStore.board?.title }} — задачи без релиза</p>
      </div>
    </div>

    <div v-if="releasesStore.loading" class="backlog-page__loading">
      <BaseSpinner />
    </div>

    <div v-else-if="releasesStore.backlog.length === 0" class="backlog-page__empty">
      <p>Бэклог пуст — все задачи распределены по релизам</p>
    </div>

    <div v-else class="backlog-list">
      <div
        v-for="card in releasesStore.backlog"
        :key="card.id"
        class="backlog-card"
      >
        <div class="backlog-card__content">
          <h4 class="backlog-card__title">{{ card.title }}</h4>
          <div class="backlog-card__meta">
            <span class="backlog-card__type">{{ card.taskType }}</span>
            <span class="backlog-card__priority" :class="`backlog-card__priority--${card.priority}`">
              {{ card.priority }}
            </span>
          </div>
        </div>
        <div v-if="draftAndActiveReleases.length > 0" class="backlog-card__assign">
          <div class="assign-wrap">
            <button
              class="assign-btn"
              @click="showAssignDropdown = showAssignDropdown === card.id ? null : card.id"
            >
              Назначить в релиз
            </button>
            <div v-if="showAssignDropdown === card.id" class="assign-dropdown">
              <div
                v-for="release in draftAndActiveReleases"
                :key="release.id"
                class="assign-dropdown__item"
                @click="assignToRelease(card.id, release.id)"
              >
                {{ release.name }}
                <span class="assign-dropdown__status">{{ release.status === 'active' ? 'активный' : 'черновик' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.backlog-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
}
.backlog-page__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.backlog-page__title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  background: var(--gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.backlog-page__subtitle {
  margin: 4px 0 0 0;
  font-size: 14px;
  color: var(--color-text-secondary);
}
.backlog-page__actions {
  display: flex;
  gap: 8px;
}
.backlog-page__loading,
.backlog-page__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: var(--color-text-secondary);
}
.backlog-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.backlog-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--color-surface);
  border: 1px solid var(--color-border-light);
  border-radius: 10px;
  padding: 12px 16px;
  gap: 12px;
}
.backlog-card__content { flex: 1; min-width: 0; }
.backlog-card__title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
}
.backlog-card__meta {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-tertiary);
}
.backlog-card__priority--critical { color: #ef4444; }
.backlog-card__priority--high { color: #f59e0b; }
.backlog-card__priority--medium { color: #7c5cfc; }
.backlog-card__priority--low { color: #10b981; }

.assign-wrap { position: relative; }
.assign-btn {
  padding: 5px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 12px;
  background: var(--color-surface-alt);
  color: var(--color-text-secondary);
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}
.assign-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}
.assign-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  padding: 4px;
  z-index: 50;
  min-width: 180px;
}
.assign-dropdown__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--color-text-primary);
  transition: background 0.1s;
}
.assign-dropdown__item:hover { background: var(--color-surface-alt); }
.assign-dropdown__status {
  font-size: 11px;
  color: var(--color-text-tertiary);
}
</style>
