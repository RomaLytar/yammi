<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useReleasesStore } from '@/stores/releases'
import { useBoardStore } from '@/stores/board'
import type { Card } from '@/types/domain'
import TaskTable from './TaskTable.vue'
import BaseSelect from '@/components/shared/BaseSelect.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

interface Emits {
  (e: 'open-card', card: Card): void
}

const emit = defineEmits<Emits>()
const releasesStore = useReleasesStore()
const boardStore = useBoardStore()

const boardId = computed(() => boardStore.boardId || '')
const assigningCardId = ref<string | null>(null)

onMounted(() => {
  if (boardId.value) {
    releasesStore.fetchBacklog(boardId.value)
    if (releasesStore.releases.length === 0) releasesStore.fetchReleases(boardId.value)
  }
})

const releaseOptions = computed(() =>
  releasesStore.releases
    .filter(r => r.status !== 'completed')
    .map(r => ({ value: r.id, label: r.name, sublabel: r.status === 'active' ? 'активный' : 'черновик' }))
)

async function handleAssign(cardId: string) {
  if (releaseOptions.value.length === 1) {
    // Only one release — assign directly
    await doAssign(cardId, String(releaseOptions.value[0].value))
  } else {
    assigningCardId.value = assigningCardId.value === cardId ? null : cardId
  }
}

async function doAssign(cardId: string, releaseId: string) {
  try {
    await releasesStore.assignCard(boardId.value, releaseId, cardId)
    assigningCardId.value = null
  } catch (err) { console.error(err) }
}
</script>

<template>
  <div class="bp">
    <div class="bp__header">
      <h2 class="bp__title">Бэклог</h2>
      <span class="bp__count">{{ releasesStore.backlog.length }} задач</span>
    </div>

    <div v-if="releasesStore.loading" class="bp__empty"><BaseSpinner /></div>
    <div v-else-if="releasesStore.backlog.length === 0" class="bp__empty">
      <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--color-text-tertiary)" stroke-width="1" stroke-linecap="round">
        <line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/>
      </svg>
      <p>Бэклог пуст — все задачи распределены по релизам</p>
    </div>

    <div v-else class="bp__content">
      <TaskTable
        :cards="releasesStore.backlog"
        :show-assign="releaseOptions.length > 0"
        @click-card="(card) => emit('open-card', card)"
        @assign-card="handleAssign"
      />

      <!-- Release picker popup for assigning -->
      <Transition name="overlay">
        <div v-if="assigningCardId && releaseOptions.length > 1" class="bp__assign-overlay" @click.self="assigningCardId = null">
          <div class="bp__assign-popup">
            <div class="bp__assign-head">
              <h3>В какой релиз?</h3>
              <button class="bp__assign-close" @click="assigningCardId = null">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="bp__assign-list">
              <div
                v-for="opt in releaseOptions"
                :key="opt.value"
                class="bp__assign-item"
                @click="doAssign(assigningCardId!, String(opt.value))"
              >
                <span class="bp__assign-name">{{ opt.label }}</span>
                <span class="bp__assign-status">{{ opt.sublabel }}</span>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.bp { padding: 20px 24px; flex: 1; overflow-y: auto; }
.bp__header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; max-width: 960px; margin-left: auto; margin-right: auto; }
.bp__title { margin: 0; font-size: var(--font-size-lg, 20px); font-weight: 700; color: var(--color-text-primary, var(--color-text)); letter-spacing: var(--letter-spacing-tight, -0.02em); }
.bp__count { font-size: var(--font-size-xs, 12px); color: var(--color-text-tertiary); font-weight: 600; padding: 2px 10px; background: var(--color-surface-alt); border-radius: var(--radius-full, 9999px); }
.bp__empty { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 60px 0; color: var(--color-text-tertiary); gap: 12px; font-size: var(--font-size-sm, 14px); }
.bp__content { max-width: 960px; margin: 0 auto; }

/* Assign popup */
.bp__assign-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.3); z-index: 100; display: flex; align-items: center; justify-content: center; }
.bp__assign-popup { background: var(--color-surface); border-radius: var(--radius-lg, 18px); box-shadow: var(--shadow-xl); width: 360px; overflow: hidden; }
.bp__assign-head { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; border-bottom: 1px solid var(--color-border-light, var(--color-border)); }
.bp__assign-head h3 { margin: 0; font-size: var(--font-size-md, 16px); font-weight: 600; }
.bp__assign-close { background: none; border: none; color: var(--color-text-tertiary); cursor: pointer; padding: 4px; border-radius: 6px; display: flex; }
.bp__assign-close:hover { color: var(--color-text-primary); background: var(--color-surface-alt); }
.bp__assign-list { padding: 6px; }
.bp__assign-item { display: flex; align-items: center; justify-content: space-between; padding: 10px 14px; border-radius: var(--radius-sm, 8px); cursor: pointer; transition: background 80ms; }
.bp__assign-item:hover { background: var(--color-primary-soft, rgba(124,92,252,0.06)); }
.bp__assign-name { font-size: var(--font-size-sm, 14px); font-weight: 500; color: var(--color-text-primary, var(--color-text)); }
.bp__assign-status { font-size: 11px; color: var(--color-text-tertiary); }

.overlay-enter-active { transition: opacity 150ms ease-out; }
.overlay-leave-active { transition: opacity 100ms ease-in; }
.overlay-enter-from, .overlay-leave-to { opacity: 0; }
</style>
