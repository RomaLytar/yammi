<script setup lang="ts">
import { computed } from 'vue'
import type { Card } from '@/types/domain'
import { useBoardStore } from '@/stores/board'

interface Props {
  cards: Card[]
  showRemove?: boolean
  showAssign?: boolean
}

interface Emits {
  (e: 'click-card', card: Card): void
  (e: 'remove-card', cardId: string): void
  (e: 'assign-card', cardId: string): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()
const boardStore = useBoardStore()

function getColumnName(columnId: string): string {
  return boardStore.columns.find(c => c.id === columnId)?.title || '—'
}

function getMemberName(userId: string): string {
  return boardStore.getMemberName(userId)
}

const priorityConfig: Record<string, { label: string; color: string }> = {
  critical: { label: 'Критический', color: '#ef4444' },
  high: { label: 'Высокий', color: '#f59e0b' },
  medium: { label: 'Средний', color: '#7c5cfc' },
  low: { label: 'Низкий', color: '#10b981' },
}

const typeConfig: Record<string, { label: string; cls: string }> = {
  bug: { label: 'Баг', cls: 'tt-type--bug' },
  feature: { label: 'Фича', cls: 'tt-type--feature' },
  task: { label: 'Задача', cls: 'tt-type--task' },
  improvement: { label: 'Улучшение', cls: 'tt-type--improvement' },
}
</script>

<template>
  <div class="tt">
    <!-- Header -->
    <div class="tt-row tt-row--header">
      <div class="tt-cell tt-cell--type">Тип</div>
      <div class="tt-cell tt-cell--title">Задача</div>
      <div class="tt-cell tt-cell--status">Статус</div>
      <div class="tt-cell tt-cell--priority">Приоритет</div>
      <div class="tt-cell tt-cell--assignee">Исполнитель</div>
      <div class="tt-cell tt-cell--action"></div>
    </div>

    <!-- Rows -->
    <div
      v-for="card in cards"
      :key="card.id"
      class="tt-row"
      @click="emit('click-card', card)"
    >
      <!-- Type -->
      <div class="tt-cell tt-cell--type">
        <span class="tt-type-icon" :class="typeConfig[card.taskType]?.cls || 'tt-type--task'">
          <svg v-if="card.taskType === 'bug'" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M8 2l1.88 1.88M14.12 3.88L16 2M9 7.13v-1a3 3 0 1 1 6 0v1"/><path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6"/>
          </svg>
          <svg v-else-if="card.taskType === 'feature'" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
          </svg>
          <svg v-else-if="card.taskType === 'improvement'" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/>
          </svg>
          <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </span>
      </div>

      <!-- Title -->
      <div class="tt-cell tt-cell--title">
        <span class="tt-title">{{ card.title }}</span>
      </div>

      <!-- Status -->
      <div class="tt-cell tt-cell--status">
        <span class="tt-status">{{ getColumnName(card.columnId) }}</span>
      </div>

      <!-- Priority -->
      <div class="tt-cell tt-cell--priority">
        <span class="tt-pri" :style="{ '--c': priorityConfig[card.priority]?.color || '#888' }">
          <span class="tt-pri__dot" />
          {{ priorityConfig[card.priority]?.label || card.priority }}
        </span>
      </div>

      <!-- Assignee -->
      <div class="tt-cell tt-cell--assignee">
        <div v-if="card.assigneeId" class="tt-user">
          <span class="tt-user__avatar">{{ getMemberName(card.assigneeId).charAt(0).toUpperCase() }}</span>
          <span class="tt-user__name">{{ getMemberName(card.assigneeId) }}</span>
        </div>
        <span v-else class="tt-no-user">—</span>
      </div>

      <!-- Action -->
      <div class="tt-cell tt-cell--action" @click.stop>
        <button v-if="showRemove" class="tt-action-btn tt-action-btn--remove" title="Убрать из релиза" @click="emit('remove-card', card.id)">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
        <button v-if="showAssign" class="tt-action-btn tt-action-btn--assign" title="В релиз" @click="emit('assign-card', card.id)">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tt {
  border: 1px solid var(--color-border-light, var(--color-border));
  border-radius: var(--radius-md, 12px);
  overflow: hidden;
  background: var(--color-surface);
}

/* Row grid */
.tt-row {
  display: grid;
  grid-template-columns: 48px 1fr 140px 120px 160px 40px;
  align-items: center;
  padding: 0 16px;
  min-height: 48px;
  border-bottom: 1px solid var(--color-border-light, var(--color-border));
  cursor: pointer;
  transition: background 80ms ease;
}
.tt-row:last-child { border-bottom: none; }
.tt-row:not(.tt-row--header):hover { background: var(--color-surface-alt, var(--color-input-bg)); }

/* Header */
.tt-row--header {
  cursor: default;
  background: var(--color-surface-alt, var(--color-input-bg));
  min-height: 40px;
  border-bottom: 1px solid var(--color-border);
}
.tt-row--header .tt-cell {
  font-size: 11px;
  font-weight: 700;
  color: var(--color-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Cell base */
.tt-cell {
  padding: 10px 0;
  font-size: 14px;
  color: var(--color-text-primary, var(--color-text));
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tt-cell--type { display: flex; align-items: center; justify-content: center; }
.tt-cell--title { padding-right: 12px; }
.tt-cell--action { display: flex; justify-content: center; }

/* Type icon */
.tt-type-icon {
  display: flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; border-radius: 8px;
}
.tt-type--bug { color: #ef4444; background: rgba(239,68,68,0.08); }
.tt-type--feature { color: #f59e0b; background: rgba(245,158,11,0.08); }
.tt-type--task { color: var(--color-primary, #7c5cfc); background: var(--color-primary-soft, rgba(124,92,252,0.08)); }
.tt-type--improvement { color: #10b981; background: rgba(16,185,129,0.08); }

/* Title */
.tt-title { font-weight: 500; }

/* Status badge */
.tt-status {
  display: inline-block;
  font-size: 12px; font-weight: 500;
  padding: 3px 10px;
  border-radius: var(--radius-full, 9999px);
  background: var(--color-surface-alt, var(--color-input-bg));
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-light, var(--color-border));
  max-width: 100%;
  overflow: hidden; text-overflow: ellipsis;
}

/* Priority */
.tt-pri {
  display: inline-flex; align-items: center; gap: 6px;
  font-size: 12px; font-weight: 600; color: var(--c);
}
.tt-pri__dot {
  width: 8px; height: 8px; border-radius: 50%;
  background: var(--c); flex-shrink: 0;
}

/* Assignee */
.tt-user { display: flex; align-items: center; gap: 8px; }
.tt-user__avatar {
  width: 26px; height: 26px; border-radius: 50%;
  background: var(--gradient-primary); color: white;
  font-size: 11px; font-weight: 700;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.tt-user__name { font-size: 13px; color: var(--color-text-secondary); overflow: hidden; text-overflow: ellipsis; }
.tt-no-user { color: var(--color-text-tertiary); font-size: 13px; }

/* Action buttons */
.tt-action-btn {
  background: none; border: none; color: var(--color-text-tertiary);
  cursor: pointer; padding: 5px; border-radius: 6px;
  opacity: 0; transition: all var(--transition-fast, 150ms); display: flex;
}
.tt-row:hover .tt-action-btn { opacity: 1; }
.tt-action-btn--remove:hover { color: var(--color-danger, #ef4444); background: var(--color-danger-soft, rgba(239,68,68,0.08)); }
.tt-action-btn--assign:hover { color: var(--color-primary, #7c5cfc); background: var(--color-primary-soft, rgba(124,92,252,0.08)); }
</style>
