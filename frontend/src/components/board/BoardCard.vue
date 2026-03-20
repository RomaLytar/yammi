<script setup lang="ts">
import type { Card } from '@/types/domain'

interface Props {
  card: Card
}

interface Emits {
  (e: 'click'): void
  (e: 'delete'): void
}

defineProps<Props>()
defineEmits<Emits>()
</script>

<template>
  <div class="board-card" @click="$emit('click')">
    <div class="board-card__header">
      <h4 class="board-card__title">{{ card.title }}</h4>
      <button
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
</template>

<style scoped>
.board-card {
  background: var(--color-surface, #fff);
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 8px;
}

.board-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
}

.board-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 4px;
}

.board-card__title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary, #111827);
  margin: 0;
  word-break: break-word;
}

.board-card__delete {
  background: none;
  border: none;
  color: var(--color-text-tertiary, #9ca3af);
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  opacity: 0;
  transition: opacity 0.2s;
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
  margin: 4px 0 0 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.board-card__footer {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--color-border, #e5e7eb);
}

.board-card__assignee {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--color-text-tertiary, #9ca3af);
}
</style>
