<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import type { BoardTemplate } from '@/types/domain'
import * as boardsApi from '@/api/boards'

export interface PresetTemplate {
  id: string
  name: string
  description: string
  columns: string[]
}

const PRESET_TEMPLATES: PresetTemplate[] = [
  {
    id: 'preset-kanban',
    name: 'Kanban-доска',
    description: 'Классический рабочий процесс для команды',
    columns: ['Бэклог', 'В работе', 'На ревью', 'Тестирование', 'Готово'],
  },
  {
    id: 'preset-home',
    name: 'Домашние дела',
    description: 'Планирование задач по дому',
    columns: ['Надо сделать', 'В процессе', 'Покупки', 'Сделано'],
  },
  {
    id: 'preset-sprint',
    name: 'Спринт',
    description: 'Для Scrum-команд с двухнедельными спринтами',
    columns: ['Спринт-бэклог', 'В разработке', 'Код-ревью', 'QA', 'Done'],
  },
  {
    id: 'preset-crm',
    name: 'Воронка продаж',
    description: 'Отслеживание сделок и клиентов',
    columns: ['Лиды', 'Контакт', 'Предложение', 'Переговоры', 'Закрыто'],
  },
  {
    id: 'preset-content',
    name: 'Контент-план',
    description: 'Управление публикациями и контентом',
    columns: ['Идеи', 'Написание', 'Редактура', 'Дизайн', 'Опубликовано'],
  },
]

// SVG icon paths per preset (Lucide-style stroke icons)
const PRESET_ICONS: Record<string, string> = {
  'preset-kanban': '<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/>',
  'preset-home': '<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>',
  'preset-sprint': '<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>',
  'preset-crm': '<line x1="12" y1="20" x2="12" y2="10"/><line x1="18" y1="20" x2="18" y2="4"/><line x1="6" y1="20" x2="6" y2="16"/>',
  'preset-content': '<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>',
}

interface Emits {
  (e: 'close'): void
  (e: 'create', data: { title: string; description: string }): void
  (e: 'createFromTemplate', data: { templateId: string; title: string; description: string }): void
  (e: 'createFromPreset', data: { preset: PresetTemplate; title: string; description: string }): void
}

const emit = defineEmits<Emits>()

const title = ref('')
const description = ref('')
const loading = ref(false)

// --- Selection state ---
type Selection =
  | { type: 'empty' }
  | { type: 'preset'; preset: PresetTemplate }
  | { type: 'personal'; templateId: string }

const selection = ref<Selection>({ type: 'empty' })

// --- Board templates ---
const boardTemplates = ref<BoardTemplate[]>([])

onMounted(async () => {
  try {
    boardTemplates.value = await boardsApi.listBoardTemplates()
  } catch (err) {
    console.error('Failed to load board templates:', err)
  }
})

const selectedPresetId = computed(() =>
  selection.value.type === 'preset' ? selection.value.preset.id : null
)

function selectPreset(preset: PresetTemplate) {
  if (selectedPresetId.value === preset.id) {
    selection.value = { type: 'empty' }
  } else {
    selection.value = { type: 'preset', preset }
  }
}

function selectPersonal(templateId: string) {
  if (templateId) {
    selection.value = { type: 'personal', templateId }
  } else {
    selection.value = { type: 'empty' }
  }
}

const personalTemplateId = computed(() =>
  selection.value.type === 'personal' ? selection.value.templateId : ''
)

function handleCreate() {
  if (!title.value.trim()) return
  loading.value = true

  const sel = selection.value
  if (sel.type === 'preset') {
    emit('createFromPreset', {
      preset: sel.preset,
      title: title.value.trim(),
      description: description.value.trim(),
    })
  } else if (sel.type === 'personal') {
    emit('createFromTemplate', {
      templateId: sel.templateId,
      title: title.value.trim(),
      description: description.value.trim(),
    })
  } else {
    emit('create', {
      title: title.value.trim(),
      description: description.value.trim(),
    })
  }
}

function handleClose() {
  if (!loading.value) {
    emit('close')
  }
}

const submitLabel = computed(() => {
  if (selection.value.type === 'preset') return 'Создать из шаблона'
  if (selection.value.type === 'personal') return 'Создать из шаблона'
  return 'Создать'
})
</script>

<template>
  <BaseModal title="Создать доску" size="medium" @close="handleClose">
    <form @submit.prevent="handleCreate" class="cbm-form">
      <!-- Preset templates -->
      <div class="cbm-section">
        <label class="cbm-section-label">Готовые шаблоны</label>
        <div class="cbm-presets">
          <button
            v-for="preset in PRESET_TEMPLATES"
            :key="preset.id"
            type="button"
            class="cbm-preset"
            :class="{ 'cbm-preset--selected': selectedPresetId === preset.id }"
            :disabled="loading"
            @click="selectPreset(preset)"
          >
            <span class="cbm-preset__icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" v-html="PRESET_ICONS[preset.id]" />
            </span>
            <span class="cbm-preset__name">{{ preset.name }}</span>
            <span class="cbm-preset__desc">{{ preset.description }}</span>
            <div class="cbm-preset__cols">
              <span
                v-for="(col, i) in preset.columns"
                :key="i"
                class="cbm-preset__col"
              >{{ col }}</span>
            </div>
            <span v-if="selectedPresetId === preset.id" class="cbm-preset__check">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </span>
          </button>
        </div>
      </div>

      <!-- Personal templates -->
      <div v-if="boardTemplates.length > 0" class="cbm-section">
        <label class="cbm-section-label">Мои шаблоны</label>
        <select
          class="cbm-template-select"
          :value="personalTemplateId"
          :disabled="loading"
          @change="selectPersonal(($event.target as HTMLSelectElement).value)"
        >
          <option value="">Не выбран</option>
          <option v-for="t in boardTemplates" :key="t.id" :value="t.id">
            {{ t.name }}
            <template v-if="t.columnsData.length"> ({{ t.columnsData.length }} кол.)</template>
          </option>
        </select>
      </div>

      <!-- Divider -->
      <div class="cbm-divider" />

      <!-- Title -->
      <BaseInput
        v-model="title"
        label="Название доски"
        placeholder="Моя доска"
        :disabled="loading"
        required
        autofocus
      />

      <!-- Description — always visible -->
      <BaseInput
        v-model="description"
        label="Описание"
        placeholder="Описание доски (необязательно)"
        :disabled="loading"
        type="textarea"
      />

      <div class="modal-actions">
        <BaseButton
          type="button"
          variant="secondary"
          :disabled="loading"
          @click="handleClose"
        >
          Отмена
        </BaseButton>
        <BaseButton
          type="submit"
          :loading="loading"
          :disabled="!title.trim()"
        >
          {{ submitLabel }}
        </BaseButton>
      </div>
    </form>
  </BaseModal>
</template>

<style scoped>
.cbm-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.cbm-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cbm-section-label {
  font-size: var(--font-size-xs, 12px);
  font-weight: 600;
  color: var(--color-text-secondary);
  letter-spacing: 0.01em;
  text-transform: uppercase;
}

/* Preset grid */
.cbm-presets {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
  gap: 10px;
}

.cbm-preset {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: var(--color-surface-alt);
  border: 1.5px solid var(--color-border-light);
  border-radius: var(--radius-md, 12px);
  cursor: pointer;
  text-align: left;
  transition: all 0.2s;
  font-family: inherit;
}

.cbm-preset:hover {
  border-color: var(--color-primary);
  background: var(--color-surface);
  box-shadow: var(--shadow-xs);
}

.cbm-preset--selected {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  box-shadow: 0 0 0 1px var(--color-primary);
}

.cbm-preset:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cbm-preset__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--color-primary-soft);
  color: var(--color-primary);
  margin-bottom: 2px;
}

.cbm-preset--selected .cbm-preset__icon {
  background: var(--color-primary);
  color: white;
}

.cbm-preset__name {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.2;
}

.cbm-preset__desc {
  font-size: 11px;
  color: var(--color-text-tertiary);
  line-height: 1.3;
}

.cbm-preset__cols {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;
}

.cbm-preset__col {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--color-surface);
  border: 1px solid var(--color-border-light);
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.cbm-preset--selected .cbm-preset__col {
  background: var(--color-surface);
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.cbm-preset__check {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Personal templates */
.cbm-template-select {
  width: 100%;
  padding: 10px 14px;
  border: 1.5px solid var(--color-input-border);
  border-radius: var(--radius-md, 12px);
  background: var(--color-input-bg);
  color: var(--color-text);
  font-size: 14px;
  font-family: inherit;
  outline: none;
  transition: all 0.15s;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1l4 4 4-4' stroke='%239ca3af' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
}

.cbm-template-select:focus {
  border-color: var(--color-input-focus);
  box-shadow: var(--shadow-focus);
}

.cbm-divider {
  height: 1px;
  background: var(--color-border-light);
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 4px;
}
</style>
