<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useReleasesStore } from '@/stores/releases'
import { useBoardStore } from '@/stores/board'
import { useAuthStore } from '@/stores/auth'
import type { Release, Card } from '@/types/domain'
import ReleaseStatusBadge from './ReleaseStatusBadge.vue'
import TaskTable from './TaskTable.vue'
import CreateReleaseModal from './CreateReleaseModal.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

interface Props {
  selectedReleaseId?: string
}

interface Emits {
  (e: 'open-card', card: Card): void
  (e: 'open-release', releaseId: string): void
  (e: 'close-release'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const releasesStore = useReleasesStore()
const boardStore = useBoardStore()
const authStore = useAuthStore()

const boardId = computed(() => boardStore.boardId || '')
const isOwner = computed(() => boardStore.board?.ownerId === authStore.userId)

const showCreateModal = ref(false)
const showConfirmDelete = ref(false)
const showConfirmComplete = ref(false)
const pendingReleaseId = ref<string | null>(null)

// Views — driven by prop from URL
const selectedRelease = computed(() =>
  props.selectedReleaseId ? releasesStore.releases.find(r => r.id === props.selectedReleaseId) || null : null
)
const releaseCards = ref<Card[]>([])
const cardsLoading = ref(false)

// Card counts per release (fetched lazily)
const cardCounts = ref<Map<string, number>>(new Map())

// Backlog popup
const showBacklogPopup = ref(false)
const backlogCards = ref<Card[]>([])
const backlogLoading = ref(false)

onMounted(async () => {
  if (boardId.value) {
    await releasesStore.fetchReleases(boardId.value)
    // Fetch card counts for each release
    for (const r of releasesStore.releases) {
      try {
        const cards = await import('@/api/releases').then(m => m.getReleaseCards(boardId.value, r.id))
        cardCounts.value.set(r.id, cards.length)
      } catch { cardCounts.value.set(r.id, 0) }
    }
  }
})

function openRelease(release: Release) {
  emit('open-release', release.id)
}

function backToList() {
  emit('close-release')
  releaseCards.value = []
}

// Load cards when selected release changes (from URL)
watch(() => props.selectedReleaseId, async (id) => {
  if (!id) { releaseCards.value = []; return }
  cardsLoading.value = true
  try {
    await releasesStore.fetchReleaseCards(boardId.value, id)
    releaseCards.value = releasesStore.releaseCards
  } catch { releaseCards.value = [] }
  finally { cardsLoading.value = false }
}, { immediate: true })

async function openBacklogPopup() {
  showBacklogPopup.value = true
  backlogLoading.value = true
  try {
    await releasesStore.fetchBacklog(boardId.value)
    backlogCards.value = releasesStore.backlog
  } catch { backlogCards.value = [] }
  finally { backlogLoading.value = false }
}

async function assignFromBacklog(cardId: string) {
  if (!selectedRelease.value) return
  try {
    await releasesStore.assignCard(boardId.value, selectedRelease.value.id, cardId)
    backlogCards.value = backlogCards.value.filter(c => c.id !== cardId)
    await releasesStore.fetchReleaseCards(boardId.value, selectedRelease.value.id)
    releaseCards.value = releasesStore.releaseCards
    cardCounts.value.set(selectedRelease.value.id, releaseCards.value.length)
  } catch (err) { console.error(err) }
}

async function removeCard(cardId: string) {
  if (!selectedRelease.value) return
  try {
    await releasesStore.removeCard(boardId.value, selectedRelease.value.id, cardId)
    releaseCards.value = releaseCards.value.filter(c => c.id !== cardId)
    cardCounts.value.set(selectedRelease.value.id, releaseCards.value.length)
  } catch (err) { console.error(err) }
}

async function handleCreate(data: { name: string; description: string }) {
  try {
    const r = await releasesStore.createRelease(boardId.value, data.name, data.description)
    cardCounts.value.set(r.id, 0)
    showCreateModal.value = false
  } catch (err) { console.error(err) }
}

async function handleStart(e: Event, releaseId: string) {
  e.stopPropagation()
  try { await releasesStore.startRelease(boardId.value, releaseId) } catch (err) { console.error(err) }
}

function confirmComplete(e: Event, releaseId: string) {
  e.stopPropagation()
  pendingReleaseId.value = releaseId
  showConfirmComplete.value = true
}

async function handleComplete() {
  if (!pendingReleaseId.value) return
  try { await releasesStore.completeRelease(boardId.value, pendingReleaseId.value) } catch (err) { console.error(err) }
  finally { showConfirmComplete.value = false; pendingReleaseId.value = null }
}

function confirmDelete(e: Event, releaseId: string) {
  e.stopPropagation()
  pendingReleaseId.value = releaseId
  showConfirmDelete.value = true
}

async function handleDelete() {
  if (!pendingReleaseId.value) return
  try { await releasesStore.deleteRelease(boardId.value, pendingReleaseId.value) } catch (err) { console.error(err) }
  finally { showConfirmDelete.value = false; pendingReleaseId.value = null }
}

function formatDate(iso?: string): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

function getDaysLeft(endDate?: string): string {
  if (!endDate) return ''
  const days = Math.ceil((new Date(endDate).getTime() - Date.now()) / 86400000)
  if (days < 0) return 'просрочен'
  if (days === 0) return 'сегодня'
  if (days === 1) return '1 день'
  if (days < 5) return `${days} дня`
  return `${days} дней`
}

function getMemberName(userId: string): string {
  return boardStore.getMemberName(userId)
}

function getColumnName(columnId: string): string {
  return boardStore.columns.find(c => c.id === columnId)?.title || '—'
}

const priorityConfig: Record<string, { label: string; color: string }> = {
  critical: { label: 'Крит', color: '#ef4444' },
  high: { label: 'Выс', color: '#f59e0b' },
  medium: { label: 'Сред', color: '#7c5cfc' },
  low: { label: 'Низ', color: '#10b981' },
}

const typeLabels: Record<string, string> = { bug: 'Баг', feature: 'Фича', task: 'Задача', improvement: 'Улучшение' }
</script>

<template>
  <div class="rp">
    <!-- ═══ LIST VIEW ═══ -->
    <template v-if="!selectedRelease">
      <div class="rp-header">
        <h2 class="rp-header__title">Релизы</h2>
        <BaseButton v-if="isOwner" size="sm" @click="showCreateModal = true">+ Новый релиз</BaseButton>
      </div>

      <div v-if="releasesStore.loading" class="rp-empty"><BaseSpinner /></div>
      <div v-else-if="releasesStore.releases.length === 0" class="rp-empty">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--color-text-tertiary)" stroke-width="1" stroke-linecap="round"><path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"/><line x1="4" y1="22" x2="4" y2="15"/></svg>
        <p>Релизов пока нет</p>
        <BaseButton v-if="isOwner" @click="showCreateModal = true">Создать первый релиз</BaseButton>
      </div>

      <div v-else class="rp-list">
        <div
          v-for="release in releasesStore.releases"
          :key="release.id"
          class="rl-card"
          :class="{
            'rl-card--active': release.status === 'active',
            'rl-card--done': release.status === 'completed',
          }"
          @click="openRelease(release)"
        >
          <div class="rl-card__top">
            <div class="rl-card__info">
              <span class="rl-card__name">{{ release.name }}</span>
              <ReleaseStatusBadge :status="release.status" />
            </div>
            <div class="rl-card__stats">
              <span class="rl-card__count">{{ cardCounts.get(release.id) ?? '...' }} задач</span>
              <span v-if="release.status === 'active' && release.endDate" class="rl-card__deadline" :class="{ 'rl-card__deadline--warn': getDaysLeft(release.endDate) === 'просрочен' }">
                {{ getDaysLeft(release.endDate) }}
              </span>
            </div>
          </div>

          <p v-if="release.description" class="rl-card__desc">{{ release.description }}</p>

          <div class="rl-card__bottom">
            <div class="rl-card__dates">
              <span v-if="release.startDate || release.startedAt">{{ formatDate(release.startDate || release.startedAt) }}</span>
              <span v-if="release.endDate"> — {{ formatDate(release.endDate) }}</span>
              <span v-if="!release.startDate && !release.startedAt">Создан {{ formatDate(release.createdAt) }}</span>
            </div>
            <div v-if="isOwner" class="rl-card__actions">
              <button v-if="release.status === 'draft'" class="rl-action rl-action--start" @click.stop="handleStart($event, release.id)">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                Запустить
              </button>
              <button v-if="release.status === 'active'" class="rl-action rl-action--complete" @click.stop="confirmComplete($event, release.id)">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
                Завершить
              </button>
              <button v-if="release.status !== 'completed'" class="rl-action rl-action--delete" @click.stop="confirmDelete($event, release.id)" title="Удалить">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- ═══ DETAIL VIEW ═══ -->
    <template v-else>
      <div class="rp-header">
        <div class="rp-header__left">
          <button class="rp-back" @click="backToList">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="15 18 9 12 15 6"/></svg>
          </button>
          <h2 class="rp-header__title">{{ selectedRelease.name }}</h2>
          <ReleaseStatusBadge :status="selectedRelease.status" />
        </div>
        <BaseButton v-if="selectedRelease.status !== 'completed'" size="sm" variant="secondary" @click="openBacklogPopup">+ Из бэклога</BaseButton>
      </div>

      <div v-if="selectedRelease.startDate || selectedRelease.description" class="rp-detail-meta">
        <span v-if="selectedRelease.startDate">{{ formatDate(selectedRelease.startDate) }} — {{ formatDate(selectedRelease.endDate) }}</span>
        <span v-if="selectedRelease.description">{{ selectedRelease.description }}</span>
      </div>

      <div class="rd-summary">
        <span class="rd-summary__count">{{ releaseCards.length }} задач</span>
      </div>

      <div v-if="cardsLoading" class="rp-empty"><BaseSpinner /></div>
      <div v-else-if="releaseCards.length === 0" class="rp-empty rp-empty--sm">Нет задач в этом релизе</div>

      <div v-else class="rd-content">
        <TaskTable
          :cards="releaseCards"
          :show-remove="selectedRelease.status !== 'completed'"
          @click-card="(card) => emit('open-card', card)"
          @remove-card="removeCard"
        />
      </div>
    </template>

    <!-- Backlog popup -->
    <Transition name="overlay">
      <div v-if="showBacklogPopup" class="bl-overlay" @click.self="showBacklogPopup = false">
        <div class="bl-popup">
          <div class="bl-popup__head">
            <h3>Добавить из бэклога</h3>
            <button class="bl-popup__close" @click="showBacklogPopup = false">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div v-if="backlogLoading" class="rp-empty rp-empty--sm"><BaseSpinner /></div>
          <div v-else-if="backlogCards.length === 0" class="rp-empty rp-empty--sm">Бэклог пуст</div>
          <div v-else class="bl-popup__list">
            <div v-for="card in backlogCards" :key="card.id" class="bl-item" @click="assignFromBacklog(card.id)">
              <div class="bl-item__info">
                <span class="bl-item__title">{{ card.title }}</span>
                <span class="bl-item__meta">{{ typeLabels[card.taskType] || card.taskType }}</span>
              </div>
              <svg class="bl-item__plus" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <CreateReleaseModal v-if="showCreateModal" @close="showCreateModal = false" @create="handleCreate" />
    <ConfirmModal v-if="showConfirmComplete" title="Завершить релиз?" message="Незавершённые задачи будут перемещены в бэклог." confirm-text="Завершить" @confirm="handleComplete" @cancel="showConfirmComplete = false" />
    <ConfirmModal v-if="showConfirmDelete" title="Удалить релиз?" message="Задачи релиза будут перемещены в бэклог." confirm-text="Удалить" variant="danger" @confirm="handleDelete" @cancel="showConfirmDelete = false" />
  </div>
</template>

<style scoped>
.rp { padding: 20px 24px; flex: 1; overflow-y: auto; }

/* Header */
.rp-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; max-width: 960px; margin-left: auto; margin-right: auto; }
.rp-header__left { display: flex; align-items: center; gap: 10px; }
.rp-header__title { margin: 0; font-size: var(--font-size-lg); font-weight: 700; color: var(--color-text-primary, var(--color-text)); letter-spacing: var(--letter-spacing-tight); }

.rp-back {
  display: flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: var(--radius-sm);
  border: 1px solid var(--color-border); background: var(--color-surface);
  color: var(--color-text-secondary); cursor: pointer; transition: all var(--transition-fast);
}
.rp-back:hover { border-color: var(--color-primary); color: var(--color-primary); background: var(--color-primary-soft); }

.rp-detail-meta { max-width: 960px; margin: -8px auto 16px; display: flex; gap: 16px; font-size: var(--font-size-xs); color: var(--color-text-tertiary); }

/* Empty state */
.rp-empty { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 60px 0; color: var(--color-text-tertiary); gap: 12px; font-size: var(--font-size-sm); }
.rp-empty--sm { padding: 32px 0; }

/* ═══ Release list cards ═══ */
.rp-list { display: flex; flex-direction: column; gap: 8px; max-width: 960px; margin: 0 auto; }

.rl-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border-light, var(--color-border));
  border-radius: var(--radius-md);
  padding: 16px 20px;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.rl-card:hover { border-color: var(--color-border); box-shadow: var(--shadow-sm); }
.rl-card--active { border-left: 3px solid var(--color-success, #10b981); }
.rl-card--done { opacity: 0.55; }

.rl-card__top { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.rl-card__info { display: flex; align-items: center; gap: 10px; min-width: 0; }
.rl-card__name { font-size: var(--font-size-md); font-weight: 600; color: var(--color-text-primary, var(--color-text)); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.rl-card__stats { display: flex; align-items: center; gap: 12px; flex-shrink: 0; }
.rl-card__count { font-size: var(--font-size-xs); color: var(--color-text-tertiary); padding: 2px 8px; background: var(--color-surface-alt, var(--color-input-bg)); border-radius: var(--radius-full); }
.rl-card__deadline { font-size: var(--font-size-xs); font-weight: 500; color: var(--color-text-secondary); }
.rl-card__deadline--warn { color: var(--color-danger, #ef4444); }

.rl-card__desc { margin: 8px 0 0; font-size: var(--font-size-sm); color: var(--color-text-secondary); line-height: var(--line-height-normal); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.rl-card__bottom { display: flex; align-items: center; justify-content: space-between; margin-top: 12px; }
.rl-card__dates { font-size: var(--font-size-xs); color: var(--color-text-tertiary); }
.rl-card__actions { display: flex; gap: 6px; align-items: center; }

/* Action buttons */
.rl-action {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 14px; border-radius: var(--radius-sm, 8px);
  font-size: 13px; font-weight: 600; font-family: inherit;
  cursor: pointer; border: none; transition: all var(--transition-fast, 150ms);
  white-space: nowrap;
}
.rl-action--start {
  background: #10b981; color: white;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.25);
}
.rl-action--start:hover { background: #059669; box-shadow: 0 4px 12px rgba(16, 185, 129, 0.35); }

.rl-action--complete {
  background: var(--color-primary, #7c5cfc); color: white;
  box-shadow: 0 2px 8px rgba(124, 92, 252, 0.25);
}
.rl-action--complete:hover { box-shadow: 0 4px 12px rgba(124, 92, 252, 0.35); }

.rl-action--delete {
  background: none; color: var(--color-text-tertiary);
  padding: 6px 8px; border-radius: 6px;
}
.rl-action--delete:hover { color: var(--color-danger, #ef4444); background: var(--color-danger-soft, rgba(239, 68, 68, 0.08)); }

/* Release detail content */
.rd-summary { max-width: 960px; margin: 0 auto 8px; font-size: var(--font-size-xs); color: var(--color-text-tertiary); }
.rd-summary__count { font-weight: 600; }
.rd-content { max-width: 960px; margin: 0 auto; }

/* ═══ Backlog popup ═══ */
.bl-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.35); z-index: 100; display: flex; align-items: center; justify-content: center; }
.bl-popup { background: var(--color-surface); border-radius: var(--radius-lg); box-shadow: var(--shadow-xl); width: 480px; max-height: 65vh; display: flex; flex-direction: column; }
.bl-popup__head { display: flex; align-items: center; justify-content: space-between; padding: 18px 20px; border-bottom: 1px solid var(--color-border-light, var(--color-border)); }
.bl-popup__head h3 { margin: 0; font-size: var(--font-size-md); font-weight: 600; }
.bl-popup__close { background: none; border: none; color: var(--color-text-tertiary); cursor: pointer; padding: 4px; border-radius: 6px; display: flex; transition: all var(--transition-fast); }
.bl-popup__close:hover { color: var(--color-text-primary, var(--color-text)); background: var(--color-surface-alt); }
.bl-popup__list { overflow-y: auto; padding: 6px; }

.bl-item { display: flex; align-items: center; justify-content: space-between; padding: 10px 14px; border-radius: var(--radius-sm); cursor: pointer; transition: background 80ms ease; }
.bl-item:hover { background: var(--color-primary-soft, rgba(124,92,252,0.06)); }
.bl-item__info { flex: 1; min-width: 0; }
.bl-item__title { font-size: var(--font-size-sm); font-weight: 500; color: var(--color-text-primary, var(--color-text)); display: block; }
.bl-item__meta { font-size: var(--font-size-xs); color: var(--color-text-tertiary); }
.bl-item__plus { color: var(--color-primary); flex-shrink: 0; opacity: 0; transition: opacity var(--transition-fast); }
.bl-item:hover .bl-item__plus { opacity: 1; }

/* Transitions */
.overlay-enter-active { transition: opacity 150ms ease-out; }
.overlay-leave-active { transition: opacity 100ms ease-in; }
.overlay-enter-from, .overlay-leave-to { opacity: 0; }
</style>
