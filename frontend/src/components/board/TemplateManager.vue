<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { CardTemplate, ColumnTemplate, BoardTemplate } from '@/types/domain'
import { useBoardStore } from '@/stores/board'
import * as boardsApi from '@/api/boards'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

const boardStore = useBoardStore()

// --- Card Templates ---
const cardTemplatesLoading = ref(false)
const deletingCardTemplateId = ref<string | null>(null)

async function deleteCardTemplate(templateId: string) {
  if (!boardStore.boardId) return
  deletingCardTemplateId.value = templateId
  try {
    await boardsApi.deleteCardTemplate(boardStore.boardId, templateId)
    boardStore.cardTemplates = boardStore.cardTemplates.filter(t => t.id !== templateId)
  } catch (err) {
    console.error('Failed to delete card template:', err)
  } finally {
    deletingCardTemplateId.value = null
  }
}

// --- Column Templates ---
const columnTemplates = ref<ColumnTemplate[]>([])
const columnTemplatesLoading = ref(false)
const deletingColumnTemplateId = ref<string | null>(null)
const savingColumnTemplate = ref(false)
const columnTemplateName = ref('')
const showColumnTemplateInput = ref(false)

async function loadColumnTemplates() {
  if (!boardStore.boardId) return
  columnTemplatesLoading.value = true
  try {
    columnTemplates.value = await boardsApi.listColumnTemplates(boardStore.boardId)
  } catch (err) {
    console.error('Failed to load column templates:', err)
  } finally {
    columnTemplatesLoading.value = false
  }
}

async function saveColumnsAsTemplate() {
  if (!boardStore.boardId || !columnTemplateName.value.trim()) return
  savingColumnTemplate.value = true
  try {
    const columnsData = boardStore.columns.map((col, i) => ({
      title: col.title,
      position: i,
    }))
    const tpl = await boardsApi.createColumnTemplate(boardStore.boardId, {
      name: columnTemplateName.value.trim(),
      columns_data: columnsData,
    })
    columnTemplates.value.push(tpl)
    showColumnTemplateInput.value = false
    columnTemplateName.value = ''
  } catch (err) {
    console.error('Failed to save column template:', err)
  } finally {
    savingColumnTemplate.value = false
  }
}

async function deleteColumnTemplate(templateId: string) {
  if (!boardStore.boardId) return
  deletingColumnTemplateId.value = templateId
  try {
    await boardsApi.deleteColumnTemplate(boardStore.boardId, templateId)
    columnTemplates.value = columnTemplates.value.filter(t => t.id !== templateId)
  } catch (err) {
    console.error('Failed to delete column template:', err)
  } finally {
    deletingColumnTemplateId.value = null
  }
}

// --- Board Templates ---
const boardTemplates = ref<BoardTemplate[]>([])
const boardTemplatesLoading = ref(false)
const deletingBoardTemplateId = ref<string | null>(null)
const savingBoardTemplate = ref(false)
const boardTemplateName = ref('')
const showBoardTemplateInput = ref(false)

async function loadBoardTemplates() {
  boardTemplatesLoading.value = true
  try {
    boardTemplates.value = await boardsApi.listBoardTemplates()
  } catch (err) {
    console.error('Failed to load board templates:', err)
  } finally {
    boardTemplatesLoading.value = false
  }
}

async function saveBoardAsTemplate() {
  if (!boardStore.board || !boardTemplateName.value.trim()) return
  savingBoardTemplate.value = true
  try {
    const columnsData = boardStore.columns.map((col, i) => ({
      title: col.title,
      position: i,
    }))
    const labelsData = boardStore.labels.map(l => ({
      name: l.name,
      color: l.color,
    }))
    const tpl = await boardsApi.createBoardTemplate({
      name: boardTemplateName.value.trim(),
      description: boardStore.board.description || '',
      columns_data: columnsData,
      labels_data: labelsData,
    })
    boardTemplates.value.push(tpl)
    showBoardTemplateInput.value = false
    boardTemplateName.value = ''
  } catch (err) {
    console.error('Failed to save board template:', err)
  } finally {
    savingBoardTemplate.value = false
  }
}

async function deleteBoardTemplate(templateId: string) {
  deletingBoardTemplateId.value = templateId
  try {
    await boardsApi.deleteBoardTemplate(templateId)
    boardTemplates.value = boardTemplates.value.filter(t => t.id !== templateId)
  } catch (err) {
    console.error('Failed to delete board template:', err)
  } finally {
    deletingBoardTemplateId.value = null
  }
}

function priorityLabel(p: string): string {
  switch (p) {
    case 'low': return 'Низкий'
    case 'medium': return 'Средний'
    case 'high': return 'Высокий'
    case 'critical': return 'Критический'
    default: return p
  }
}

onMounted(() => {
  loadColumnTemplates()
  loadBoardTemplates()
})
</script>

<template>
  <div class="tpl-manager">
    <!-- Card Templates -->
    <div class="tpl-section">
      <h3 class="tpl-section__title">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="3" y="3" width="18" height="18" rx="2" /><path d="M7 7h10M7 12h10M7 17h6" />
        </svg>
        Шаблоны карточек
      </h3>

      <div v-if="cardTemplatesLoading" class="tpl-section__center">
        <BaseSpinner size="sm" />
      </div>
      <div v-else-if="boardStore.cardTemplates.length === 0" class="tpl-section__empty">
        Нет сохранённых шаблонов карточек
      </div>
      <div v-else class="tpl-list">
        <div v-for="tpl in boardStore.cardTemplates" :key="tpl.id" class="tpl-item">
          <div class="tpl-item__body">
            <span class="tpl-item__name">{{ tpl.name }}</span>
            <span class="tpl-item__meta">{{ tpl.title }} &middot; {{ priorityLabel(tpl.priority) }}</span>
          </div>
          <button
            class="tpl-item__delete"
            :disabled="deletingCardTemplateId === tpl.id"
            title="Удалить"
            @click="deleteCardTemplate(tpl.id)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
        </div>
      </div>
      <p class="tpl-hint">Сохранить карточку как шаблон можно в окне редактирования карточки.</p>
    </div>

    <!-- Column Templates -->
    <div class="tpl-section">
      <h3 class="tpl-section__title">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <rect x="3" y="3" width="7" height="18" rx="1" /><rect x="14" y="3" width="7" height="18" rx="1" />
        </svg>
        Шаблоны колонок
      </h3>

      <div v-if="columnTemplatesLoading" class="tpl-section__center">
        <BaseSpinner size="sm" />
      </div>
      <div v-else-if="columnTemplates.length === 0" class="tpl-section__empty">
        Нет сохранённых шаблонов колонок
      </div>
      <div v-else class="tpl-list">
        <div v-for="tpl in columnTemplates" :key="tpl.id" class="tpl-item">
          <div class="tpl-item__body">
            <span class="tpl-item__name">{{ tpl.name }}</span>
            <span class="tpl-item__meta">{{ tpl.columnsData.length }} колонок</span>
          </div>
          <button
            class="tpl-item__delete"
            :disabled="deletingColumnTemplateId === tpl.id"
            title="Удалить"
            @click="deleteColumnTemplate(tpl.id)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
        </div>
      </div>

      <div v-if="showColumnTemplateInput" class="tpl-save-form">
        <input
          v-model="columnTemplateName"
          class="tpl-save-form__input"
          placeholder="Название шаблона..."
          :disabled="savingColumnTemplate"
          @keydown.enter="saveColumnsAsTemplate"
          @keydown.escape="showColumnTemplateInput = false; columnTemplateName = ''"
        />
        <div class="tpl-save-form__actions">
          <BaseButton size="sm" :loading="savingColumnTemplate" :disabled="!columnTemplateName.trim()" @click="saveColumnsAsTemplate">
            Сохранить
          </BaseButton>
          <BaseButton size="sm" variant="secondary" @click="showColumnTemplateInput = false; columnTemplateName = ''">
            Отмена
          </BaseButton>
        </div>
      </div>
      <BaseButton v-else variant="secondary" size="sm" @click="showColumnTemplateInput = true">
        Сохранить текущие колонки как шаблон
      </BaseButton>
    </div>

    <!-- Board Templates -->
    <div class="tpl-section">
      <h3 class="tpl-section__title">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="3" y="3" width="18" height="18" rx="2" /><line x1="9" y1="3" x2="9" y2="21" /><line x1="3" y1="9" x2="21" y2="9" />
        </svg>
        Шаблоны досок
      </h3>

      <div v-if="boardTemplatesLoading" class="tpl-section__center">
        <BaseSpinner size="sm" />
      </div>
      <div v-else-if="boardTemplates.length === 0" class="tpl-section__empty">
        Нет сохранённых шаблонов досок
      </div>
      <div v-else class="tpl-list">
        <div v-for="tpl in boardTemplates" :key="tpl.id" class="tpl-item">
          <div class="tpl-item__body">
            <span class="tpl-item__name">{{ tpl.name }}</span>
            <span class="tpl-item__meta">
              {{ tpl.columnsData.length }} колонок
              <template v-if="tpl.labelsData.length"> &middot; {{ tpl.labelsData.length }} меток</template>
            </span>
          </div>
          <button
            class="tpl-item__delete"
            :disabled="deletingBoardTemplateId === tpl.id"
            title="Удалить"
            @click="deleteBoardTemplate(tpl.id)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
        </div>
      </div>

      <div v-if="showBoardTemplateInput" class="tpl-save-form">
        <input
          v-model="boardTemplateName"
          class="tpl-save-form__input"
          placeholder="Название шаблона..."
          :disabled="savingBoardTemplate"
          @keydown.enter="saveBoardAsTemplate"
          @keydown.escape="showBoardTemplateInput = false; boardTemplateName = ''"
        />
        <div class="tpl-save-form__actions">
          <BaseButton size="sm" :loading="savingBoardTemplate" :disabled="!boardTemplateName.trim()" @click="saveBoardAsTemplate">
            Сохранить
          </BaseButton>
          <BaseButton size="sm" variant="secondary" @click="showBoardTemplateInput = false; boardTemplateName = ''">
            Отмена
          </BaseButton>
        </div>
      </div>
      <BaseButton v-else variant="secondary" size="sm" @click="showBoardTemplateInput = true">
        Сохранить доску как шаблон
      </BaseButton>
    </div>
  </div>
</template>

<style scoped>
.tpl-manager {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.tpl-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tpl-section__title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
}

.tpl-section__center {
  display: flex;
  justify-content: center;
  padding: 12px 0;
}

.tpl-section__empty {
  font-size: 13px;
  color: var(--color-text-tertiary);
  padding: 8px 0;
}

.tpl-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tpl-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  background: var(--color-surface-alt);
  border: 1px solid var(--color-border-light);
  transition: all 0.15s;
}

.tpl-item:hover {
  border-color: var(--color-border);
}

.tpl-item__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tpl-item__name {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tpl-item__meta {
  font-size: 11px;
  color: var(--color-text-tertiary);
}

.tpl-item__delete {
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  flex-shrink: 0;
}

.tpl-item__delete:hover {
  color: var(--color-danger);
  background: var(--color-danger-soft);
}

.tpl-item__delete:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.tpl-save-form {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tpl-save-form__input {
  width: 100%;
  padding: 8px 10px;
  border: 1.5px solid var(--color-input-border);
  border-radius: var(--radius-sm);
  background: var(--color-input-bg);
  color: var(--color-text);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: all 0.15s;
  box-sizing: border-box;
}

.tpl-save-form__input:focus {
  border-color: var(--color-input-focus);
  box-shadow: var(--shadow-focus);
}

.tpl-save-form__actions {
  display: flex;
  gap: 6px;
}

.tpl-hint {
  font-size: 11px;
  color: var(--color-text-tertiary);
  margin: 0;
  font-style: italic;
}
</style>
