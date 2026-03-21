<script setup lang="ts">
import { ref } from 'vue'
import type { Column, Card } from '@/types/domain'
import BoardCard from './BoardCard.vue'
import Draggable from 'vuedraggable'

interface Props {
  column: Column
  isOwner: boolean
  currentUserId: string
  selectMode?: boolean
  selectedIds?: Set<string>
}

interface Emits {
  (e: 'add-card'): void
  (e: 'card-click', card: Card): void
  (e: 'card-delete', cardId: string): void
  (e: 'card-move', event: { cardId: string; fromColumnId: string; toColumnId: string; newIndex: number }): void
  (e: 'update-title', title: string): void
  (e: 'delete'): void
  (e: 'card-toggle-select', cardId: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const isEditingTitle = ref(false)
const editedTitle = ref(props.column.title)

function startEdit() {
  isEditingTitle.value = true
  editedTitle.value = props.column.title
}

function saveTitle() {
  if (editedTitle.value.trim() && editedTitle.value !== props.column.title) {
    emit('update-title', editedTitle.value.trim())
  }
  isEditingTitle.value = false
}

function cancelEdit() {
  editedTitle.value = props.column.title
  isEditingTitle.value = false
}

// Vuedraggable events
function onDragChange(event: any) {
  console.log('[BoardColumn]', props.column.title, '- onDragChange:', Object.keys(event))

  // Карточка перемещена внутри этой колонки
  if (event.moved) {
    const card = event.moved.element as Card
    const newIndex = event.moved.newIndex
    console.log('[BoardColumn] MOVED within column:', card.title, 'to index', newIndex)
    emit('card-move', {
      cardId: card.id,
      fromColumnId: props.column.id,
      toColumnId: props.column.id,
      newIndex
    })
  }

  // Карточка добавлена в эту колонку (из другой колонки)
  if (event.added) {
    const card = event.added.element as Card
    const newIndex = event.added.newIndex
    const fromColumnId = card.columnId

    console.log('[BoardColumn] ADDED to column:', card.title)
    console.log('[BoardColumn]   from:', fromColumnId, 'to:', props.column.id, 'index:', newIndex)

    emit('card-move', {
      cardId: card.id,
      fromColumnId,
      toColumnId: props.column.id,
      newIndex
    })
  }
}
</script>

<template>
  <div class="board-column">
    <div class="board-column__header">
      <input
        v-if="isEditingTitle"
        v-model="editedTitle"
        class="board-column__title-input"
        @blur="saveTitle"
        @keydown.enter="saveTitle"
        @keydown.esc="cancelEdit"
        autofocus
      />
      <h3
        v-else
        class="board-column__title"
        @dblclick="startEdit"
      >
        {{ column.title }}
        <span class="board-column__count">{{ column.cards.length }}</span>
      </h3>

      <div v-if="isOwner" class="board-column__actions">
        <button
          class="board-column__action"
          title="Удалить колонку"
          @click="$emit('delete')"
        >
          ×
        </button>
      </div>
    </div>

    <Draggable
      v-model="column.cards"
      group="cards"
      :animation="200"
      class="board-column__cards"
      item-key="id"
      ghost-class="ghost-card"
      @change="onDragChange"
    >
      <template #item="{ element }">
        <BoardCard
          :key="element.id"
          :card="element"
          :can-delete="isOwner || element.creatorId === currentUserId"
          :select-mode="selectMode"
          :selected="selectedIds?.has(element.id)"
          :can-select="isOwner || element.creatorId === currentUserId"
          @click="selectMode ? (isOwner || element.creatorId === currentUserId) && $emit('card-toggle-select', element.id) : $emit('card-click', element)"
          @delete="$emit('card-delete', element.id)"
          @toggle-select="$emit('card-toggle-select', element.id)"
        />
      </template>
    </Draggable>

    <button class="board-column__add-card" @click="$emit('add-card')">
      + Добавить карточку
    </button>
  </div>
</template>

<style scoped>
.board-column {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 16px;
  min-width: 300px;
  max-width: 300px;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 180px);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
}

.board-column:hover {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.board-column__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  gap: 8px;
}

.board-column__title {
  flex: 1;
  font-size: 15px;
  font-weight: 700;
  color: var(--color-text-primary, #111827);
  margin: 0;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: background 0.2s;
}

.board-column__title:hover {
  background: var(--color-surface, #f3f4f6);
}

.board-column__count {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-tertiary, #6b7280);
  background: var(--color-surface, #e5e7eb);
  padding: 3px 10px;
  border-radius: 14px;
  min-width: 24px;
  text-align: center;
}

.board-column__title-input {
  flex: 1;
  font-size: 15px;
  font-weight: 700;
  padding: 8px 10px;
  border: 2px solid var(--color-primary, #3b82f6);
  border-radius: 8px;
  outline: none;
  background: white;
}

.board-column__actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.board-column:hover .board-column__actions {
  opacity: 1;
}

.board-column__action {
  background: none;
  border: none;
  color: var(--color-text-tertiary, #9ca3af);
  font-size: 22px;
  cursor: pointer;
  padding: 4px;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s;
}

.board-column__action:hover {
  background: var(--color-danger-light, #fee2e2);
  color: var(--color-danger, #dc2626);
}

.board-column__cards {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  min-height: 120px;
  padding: 2px;
  margin: 0 -4px;
  padding: 0 4px;
}

.board-column__cards::-webkit-scrollbar {
  width: 8px;
}

.board-column__cards::-webkit-scrollbar-track {
  background: transparent;
}

.board-column__cards::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 4px;
}

.board-column__cards::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

.board-column__add-card {
  margin-top: 10px;
  padding: 10px 14px;
  background: none;
  border: 2px dashed var(--color-border, #d1d5db);
  border-radius: 10px;
  color: var(--color-text-tertiary, #6b7280);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.board-column__add-card:hover {
  border-color: var(--color-primary, #3b82f6);
  color: var(--color-primary, #3b82f6);
  background: var(--color-primary-light, #eff6ff);
  transform: translateY(-1px);
}

.board-column__add-card:active {
  transform: translateY(0);
}

.ghost-card {
  opacity: 0.4;
  background: var(--color-primary-light, #dbeafe);
  border: 2px dashed var(--color-primary, #3b82f6);
  border-radius: 12px;
}
</style>
