<script setup lang="ts">
import { ref } from 'vue'
import type { Column, Card } from '@/types/domain'
import BoardCard from './BoardCard.vue'
import Draggable from 'vuedraggable'

interface Props {
  column: Column
}

interface Emits {
  (e: 'add-card'): void
  (e: 'card-click', card: Card): void
  (e: 'card-delete', cardId: string): void
  (e: 'card-move', event: { cardId: string; fromColumnId: string; toColumnId: string; newIndex: number }): void
  (e: 'update-title', title: string): void
  (e: 'delete'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const isEditingTitle = ref(false)
const editedTitle = ref(props.column.title)

// Track card being removed for cross-column drag
let removedCardInfo: { cardId: string; columnId: string } | null = null

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
  console.log('[BoardColumn] onDragChange called on column:', props.column.id, props.column.title)
  console.log('[BoardColumn] event keys:', Object.keys(event))

  // Карточка удалена из этой колонки (перемещается в другую)
  if (event.removed) {
    const card = event.removed.element as Card
    console.log('[BoardColumn] event.removed: card', card.id, 'removed from column', props.column.id)
    // Сохраняем инфо для следующего event.added в целевой колонке
    removedCardInfo = { cardId: card.id, columnId: props.column.id }
  }

  // Карточка добавлена в эту колонку (из другой колонки)
  if (event.added) {
    const card = event.added.element as Card
    const newIndex = event.added.newIndex

    // Сначала пытаемся использовать removedCardInfo (самый надежный способ)
    // Затем fallback на card.columnId
    const fromColumnId = removedCardInfo?.cardId === card.id
      ? removedCardInfo.columnId
      : card.columnId

    console.log('[BoardColumn] event.added detected')
    console.log('[BoardColumn]   card.id:', card.id)
    console.log('[BoardColumn]   fromColumnId (determined):', fromColumnId)
    console.log('[BoardColumn]   card.columnId:', card.columnId)
    console.log('[BoardColumn]   removedCardInfo:', removedCardInfo)
    console.log('[BoardColumn]   toColumnId:', props.column.id)
    console.log('[BoardColumn]   newIndex:', newIndex)

    if (!fromColumnId) {
      console.error('[BoardColumn] ERROR: fromColumnId is null/undefined!')
      return
    }

    emit('card-move', { cardId: card.id, fromColumnId, toColumnId: props.column.id, newIndex })
    console.log('[BoardColumn] emitted card-move event')

    // Очищаем removedCardInfo после использования
    removedCardInfo = null
  }

  // Карточка перемещена внутри этой колонки
  if (event.moved) {
    const card = event.moved.element as Card
    const newIndex = event.moved.newIndex
    console.log('[BoardColumn] event.moved detected')
    console.log('[BoardColumn]   card.id:', card.id)
    console.log('[BoardColumn]   column.id:', props.column.id)
    console.log('[BoardColumn]   newIndex:', newIndex)

    emit('card-move', { cardId: card.id, fromColumnId: props.column.id, toColumnId: props.column.id, newIndex })
    console.log('[BoardColumn] emitted card-move event')
  }

  if (!event.added && !event.moved && !event.removed) {
    console.warn('[BoardColumn] onDragChange called but no recognized event type:', Object.keys(event))
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

      <div class="board-column__actions">
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
          @click="$emit('card-click', element)"
          @delete="$emit('card-delete', element.id)"
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
  background: var(--color-surface-alt, #f9fafb);
  border-radius: 12px;
  padding: 12px;
  min-width: 280px;
  max-width: 280px;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 200px);
}

.board-column__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  gap: 8px;
}

.board-column__title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary, #111827);
  margin: 0;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}

.board-column__title:hover {
  background: var(--color-surface, #fff);
}

.board-column__count {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-tertiary, #9ca3af);
  background: var(--color-surface, #fff);
  padding: 2px 8px;
  border-radius: 12px;
}

.board-column__title-input {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  padding: 6px 8px;
  border: 2px solid var(--color-primary, #3b82f6);
  border-radius: 4px;
  outline: none;
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
  font-size: 20px;
  cursor: pointer;
  padding: 4px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.board-column__action:hover {
  background: var(--color-danger-light, #fee2e2);
  color: var(--color-danger, #dc2626);
}

.board-column__cards {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  min-height: 100px;
  padding: 2px;
}

.board-column__add-card {
  margin-top: 8px;
  padding: 8px 12px;
  background: none;
  border: 2px dashed var(--color-border, #e5e7eb);
  border-radius: 8px;
  color: var(--color-text-tertiary, #9ca3af);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.board-column__add-card:hover {
  border-color: var(--color-primary, #3b82f6);
  color: var(--color-primary, #3b82f6);
  background: var(--color-primary-light, #eff6ff);
}

.ghost-card {
  opacity: 0.5;
  background: var(--color-primary-light, #eff6ff);
  border: 2px dashed var(--color-primary, #3b82f6);
}
</style>
