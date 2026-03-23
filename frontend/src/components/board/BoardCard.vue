<script setup lang="ts">
import { computed } from 'vue'
import type { Card, Label } from '@/types/domain'
import { useBoardStore } from '@/stores/board'

interface Props {
  card: Card
  canDelete?: boolean
  selectMode?: boolean
  selected?: boolean
  canSelect?: boolean
}

interface Emits {
  (e: 'click'): void
  (e: 'delete'): void
  (e: 'toggle-select'): void
}

const props = defineProps<Props>()
defineEmits<Emits>()

const boardStore = useBoardStore()

// --- Priority ---
const priorityColor = computed(() => {
  switch (props.card.priority) {
    case 'low': return 'var(--color-success, #10b981)'
    case 'medium': return 'var(--color-primary, #7c5cfc)'
    case 'high': return '#f59e0b'
    case 'critical': return 'var(--color-danger, #ef4444)'
    default: return 'var(--color-primary, #7c5cfc)'
  }
})

const priorityLabel = computed(() => {
  switch (props.card.priority) {
    case 'low': return 'Низкий'
    case 'medium': return 'Средний'
    case 'high': return 'Высокий'
    case 'critical': return 'Критический'
    default: return ''
  }
})

// --- Due date ---
const dueDateFormatted = computed(() => {
  if (!props.card.dueDate) return null
  const d = new Date(props.card.dueDate)
  const months = ['янв', 'фев', 'мар', 'апр', 'май', 'июн', 'июл', 'авг', 'сен', 'окт', 'ноя', 'дек']
  return `${d.getDate()} ${months[d.getMonth()]}`
})

const dueDateClass = computed(() => {
  if (!props.card.dueDate) return ''
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const due = new Date(props.card.dueDate)
  due.setHours(0, 0, 0, 0)
  const diffDays = (due.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  if (diffDays < 0) return 'board-card__due--overdue'
  if (diffDays <= 3) return 'board-card__due--soon'
  return 'board-card__due--ok'
})

// --- Labels (show as colored dots, max 3 + overflow) ---
const cardLabels = computed<Label[]>(() => props.card.labels || [])
const visibleLabels = computed(() => cardLabels.value.slice(0, 3))
const extraLabelCount = computed(() => Math.max(0, cardLabels.value.length - 3))

// --- Checklist stats ---
const checklistStats = computed(() => props.card.checklistStats)
const checklistPercent = computed(() => {
  if (!checklistStats.value || checklistStats.value.total === 0) return 0
  return Math.round((checklistStats.value.checked / checklistStats.value.total) * 100)
})
</script>

<template>
  <div
    class="board-card"
    :class="{ 'board-card--selected': selected, 'board-card--select-mode': selectMode }"
    @click="$emit('click')"
  >
    <div v-if="selectMode" class="board-card__checkbox" @click.stop>
      <input
        v-if="canSelect"
        type="checkbox"
        :checked="selected"
        @change="$emit('toggle-select')"
      />
      <span v-else class="checkbox-placeholder" />
    </div>
    <div class="board-card__content">
      <div class="board-card__header">
        <h4 class="board-card__title">{{ card.title }}</h4>
        <button
          v-if="canDelete && !selectMode"
          class="board-card__delete"
          title="Удалить"
          @click.stop="$emit('delete')"
        >
          ×
        </button>
      </div>
      <div v-if="card.description" class="board-card__description" v-html="card.description" />

      <!-- Labels row -->
      <div v-if="cardLabels.length > 0" class="board-card__labels">
        <span
          v-for="label in visibleLabels"
          :key="label.id"
          class="board-card__label"
          :style="{ background: label.color }"
          :title="label.name"
        />
        <span v-if="extraLabelCount > 0" class="board-card__label-extra">+{{ extraLabelCount }}</span>
      </div>

      <!-- Footer indicators -->
      <div class="board-card__footer">
        <!-- Task type icon -->
        <span class="board-card__type" :title="card.taskType">
          <svg v-if="card.taskType === 'bug'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M8 2l1.88 1.88M14.12 3.88L16 2M9 7.13v-1a3.003 3.003 0 1 1 6 0v1"/>
            <path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6"/>
            <path d="M12 20v-9M6.53 9C4.6 8.8 3 7.1 3 5M6 13H2M6 17l-4 1M17.47 9c1.93-.2 3.53-1.9 3.53-4M18 13h4M18 17l4 1"/>
          </svg>
          <svg v-else-if="card.taskType === 'feature'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
          </svg>
          <svg v-else-if="card.taskType === 'improvement'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/>
          </svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </span>

        <!-- Priority dot -->
        <span
          class="board-card__priority"
          :style="{ background: priorityColor }"
          :title="priorityLabel"
        />

        <!-- Due date -->
        <span v-if="dueDateFormatted" class="board-card__due" :class="dueDateClass">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/>
          </svg>
          {{ dueDateFormatted }}
        </span>

        <!-- Checklist progress -->
        <span v-if="checklistStats && checklistStats.total > 0" class="board-card__checklist">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <polyline points="9 11 12 14 22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
          </svg>
          {{ checklistStats.checked }}/{{ checklistStats.total }}
          <span class="board-card__checklist-bar">
            <span class="board-card__checklist-fill" :style="{ width: checklistPercent + '%' }" />
          </span>
        </span>
      </div>
    </div>
    <!-- Assignee avatar (right side) -->
    <div class="board-card__avatar-wrap">
      <div
        v-if="card.assigneeId"
        class="board-card__avatar board-card__avatar--assigned"
        :title="boardStore.getMemberName(card.assigneeId)"
      >
        {{ boardStore.getMemberName(card.assigneeId).charAt(0).toUpperCase() }}
      </div>
      <div
        v-else
        class="board-card__avatar board-card__avatar--empty"
        title="Не назначен"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
          <circle cx="12" cy="7" r="4"/>
        </svg>
      </div>
    </div>
  </div>
</template>

<style scoped>
.board-card {
  background: var(--color-surface-alt);
  border: 1px solid var(--color-border-light);
  border-radius: 12px;
  padding: 14px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-bottom: 10px;
  box-shadow: var(--shadow-xs);
  position: relative;
  overflow: hidden;
  display: flex;
  gap: 10px;
  align-items: center;
}

.board-card--selected {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.board-card--select-mode {
  cursor: default;
}

.board-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 100%;
  background: var(--gradient-primary);
  opacity: 0;
  transition: opacity 0.3s;
}

.board-card:hover {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.board-card:not(.board-card--select-mode):hover::before {
  opacity: 1;
}

.board-card__checkbox {
  display: flex;
  align-items: flex-start;
  padding-top: 2px;
  flex-shrink: 0;
}

.board-card__checkbox input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--color-primary);
  cursor: pointer;
}

.checkbox-placeholder {
  width: 16px;
  height: 16px;
  border: 1.5px solid var(--color-border);
  border-radius: 3px;
  opacity: 0.3;
}

.board-card__content {
  flex: 1;
  min-width: 0;
}

.board-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 10px;
}

.board-card__title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary, #111827);
  margin: 0;
  word-break: break-word;
  line-height: 1.5;
}

.board-card__delete {
  background: none;
  border: none;
  color: var(--color-text-tertiary, #9ca3af);
  font-size: 22px;
  line-height: 1;
  cursor: pointer;
  padding: 2px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  opacity: 0;
  transition: all 0.2s;
  flex-shrink: 0;
}

.board-card:hover .board-card__delete {
  opacity: 1;
}

.board-card__delete:hover {
  background: var(--color-danger-light, #fee2e2);
  color: var(--color-danger, #dc2626);
}

.board-card__description {
  font-size: 13px;
  color: var(--color-text-secondary, #6b7280);
  margin: 6px 0 0 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Labels */
.board-card__labels {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 8px;
}

.board-card__label {
  width: 24px;
  height: 6px;
  border-radius: 3px;
  display: inline-block;
}

.board-card__label-extra {
  font-size: 10px;
  color: var(--color-text-tertiary, #9ca3af);
  font-weight: 600;
}

/* Footer indicators */
.board-card__footer {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.board-card__type {
  display: flex;
  align-items: center;
  color: var(--color-text-tertiary, #9ca3af);
}

.board-card__priority {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.board-card__due {
  display: flex;
  align-items: center;
  gap: 3px;
  font-size: 11px;
  font-weight: 500;
  padding: 1px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.board-card__due--ok {
  color: var(--color-success, #10b981);
  background: var(--color-success-soft, rgba(16, 185, 129, 0.08));
}

.board-card__due--soon {
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.08);
}

.board-card__due--overdue {
  color: var(--color-danger, #ef4444);
  background: var(--color-danger-soft, rgba(239, 68, 68, 0.08));
}

.board-card__checklist {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 500;
  color: var(--color-text-tertiary, #9ca3af);
}

.board-card__checklist-bar {
  width: 30px;
  height: 4px;
  background: var(--color-border, #e5e7eb);
  border-radius: 2px;
  overflow: hidden;
  display: inline-block;
}

.board-card__checklist-fill {
  height: 100%;
  background: var(--color-success, #10b981);
  border-radius: 2px;
  display: block;
  transition: width 0.2s;
}

/* Assignee avatar */
.board-card__avatar-wrap {
  flex-shrink: 0;
  align-self: flex-start;
  margin-top: 2px;
}

.board-card__avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  cursor: default;
}

.board-card__avatar--assigned {
  background: var(--gradient-primary);
  color: white;
}

.board-card__avatar--empty {
  background: var(--color-input-bg, #f3f4f6);
  color: var(--color-text-tertiary, #9ca3af);
  border: 1.5px dashed var(--color-border, #d1d5db);
}
</style>
