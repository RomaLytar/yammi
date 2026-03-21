<script setup lang="ts">
import type { Card } from '@/types/domain'

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

defineProps<Props>()
defineEmits<Emits>()
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
      <p v-if="card.description" class="board-card__description">
        {{ card.description }}
      </p>
      <div v-if="card.assigneeId" class="board-card__footer">
        <div class="board-card__assignee">
          <span>👤</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.board-card {
  background: white;
  border: 1px solid transparent;
  border-radius: 12px;
  padding: 14px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-bottom: 10px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06);
  position: relative;
  overflow: hidden;
  display: flex;
  gap: 10px;
}

.board-card--selected {
  border-color: #3b82f6;
  background: #eff6ff;
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  opacity: 0;
  transition: opacity 0.3s;
}

.board-card:hover {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.board-card:not(.board-card--select-mode):hover::before {
  opacity: 1;
}

.board-card:active:not(.board-card--select-mode) {
  transform: translateY(0);
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
  accent-color: #3b82f6;
  cursor: pointer;
}

.checkbox-placeholder {
  width: 16px;
  height: 16px;
  border: 1.5px solid #d1d5db;
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
  margin-bottom: 6px;
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
  transform: scale(1.1);
}

.board-card__description {
  font-size: 13px;
  color: var(--color-text-secondary, #6b7280);
  margin: 6px 0 0 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.board-card__footer {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--color-border, #e5e7eb);
  display: flex;
  align-items: center;
  gap: 8px;
}

.board-card__assignee {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--color-text-tertiary, #6b7280);
  background: var(--color-surface-alt, #f3f4f6);
  padding: 4px 10px;
  border-radius: 12px;
}
</style>
