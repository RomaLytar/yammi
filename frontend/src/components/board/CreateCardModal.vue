<script setup lang="ts">
import { ref, computed } from 'vue'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSearchSelect from '@/components/shared/BaseSearchSelect.vue'
import RichTextEditor from '@/components/shared/RichTextEditor.vue'
import { useBoardStore } from '@/stores/board'

import type { Priority, TaskType } from '@/types/domain'

interface Emits {
  (e: 'close'): void
  (e: 'create', data: {
    title: string; description: string; assigneeId?: string; files?: File[];
    dueDate?: string; priority?: string; taskType?: string
  }): void
}

const emit = defineEmits<Emits>()
const boardStore = useBoardStore()

const title = ref('')
const description = ref('')
const selectedAssignee = ref('')
const selectedPriority = ref<Priority>('medium')
const selectedTaskType = ref<TaskType>('task')
const selectedDueDate = ref('')
const loading = ref(false)
const isDragging = ref(false)

const assigneeOptions = computed(() =>
  boardStore.members.map(m => ({
    value: m.user_id,
    label: boardStore.getMemberName(m.user_id),
    sublabel: m.role === 'owner' ? 'владелец' : boardStore.getMemberEmail(m.user_id),
  }))
)

const selectedAssigneeName = computed(() =>
  selectedAssignee.value ? boardStore.getMemberName(selectedAssignee.value) : ''
)

// --- File uploads with local preview ---
interface PendingFile {
  file: File
  previewUrl: string | null
}
const pendingFiles = ref<PendingFile[]>([])

function isImage(file: File): boolean {
  return file.type.startsWith('image/')
}

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (!files) return
  for (const file of Array.from(files)) {
    pendingFiles.value.push({ file, previewUrl: isImage(file) ? URL.createObjectURL(file) : null })
  }
  target.value = ''
}

function removePendingFile(index: number) {
  const pf = pendingFiles.value[index]
  if (pf.previewUrl) URL.revokeObjectURL(pf.previewUrl)
  pendingFiles.value.splice(index, 1)
}

function handleDrop(event: DragEvent) {
  isDragging.value = false
  const files = event.dataTransfer?.files
  if (!files) return
  for (const file of Array.from(files)) {
    pendingFiles.value.push({ file, previewUrl: isImage(file) ? URL.createObjectURL(file) : null })
  }
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' Б'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' КБ'
  return (bytes / (1024 * 1024)).toFixed(1) + ' МБ'
}

function handleCreate() {
  if (!title.value.trim()) return
  loading.value = true
  emit('create', {
    title: title.value.trim(),
    description: description.value,
    assigneeId: selectedAssignee.value || undefined,
    files: pendingFiles.value.length ? pendingFiles.value.map(pf => pf.file) : undefined,
    dueDate: selectedDueDate.value || undefined,
    priority: selectedPriority.value,
    taskType: selectedTaskType.value,
  })
}

function handleClose() {
  if (!loading.value) {
    for (const pf of pendingFiles.value) {
      if (pf.previewUrl) URL.revokeObjectURL(pf.previewUrl)
    }
    emit('close')
  }
}
</script>

<template>
  <BaseModal size="medium" @close="handleClose">
    <template #header>
      <div class="ccm-header">
        <div class="ccm-header__icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="3" width="18" height="18" rx="2" />
            <line x1="12" y1="8" x2="12" y2="16" />
            <line x1="8" y1="12" x2="16" y2="12" />
          </svg>
        </div>
        <h2 class="ccm-header__title">Новая карточка</h2>
      </div>
    </template>

    <div class="ccm-body">
      <!-- Title -->
      <BaseInput
        v-model="title"
        placeholder="Название задачи..."
        :disabled="loading"
        autofocus
      />

      <!-- Description -->
      <div class="ccm-section">
        <div class="ccm-section__label">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <line x1="17" y1="10" x2="3" y2="10" /><line x1="21" y1="6" x2="3" y2="6" /><line x1="21" y1="14" x2="3" y2="14" /><line x1="17" y1="18" x2="3" y2="18" />
          </svg>
          Описание
        </div>
        <RichTextEditor
          v-model="description"
          placeholder="Подробности задачи..."
          :disabled="loading"
        />
      </div>

      <!-- Details row: assignee -->
      <div class="ccm-details">
        <div class="ccm-details__label">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
          </svg>
          Исполнитель
        </div>
        <div class="ccm-assignee">
          <BaseSearchSelect
            v-model="selectedAssignee"
            :options="assigneeOptions"
            placeholder="Выберите участника..."
            :disabled="loading"
            clearable
          />
        </div>
      </div>

      <!-- Metadata: Priority, Type, Due date -->
      <div class="ccm-meta-row">
        <div class="ccm-meta-group">
          <span class="ccm-meta-label">Приоритет</span>
          <div class="ccm-priority-btns">
            <button
              class="ccm-priority-btn"
              :class="{ 'ccm-priority-btn--active': selectedPriority === 'low' }"
              style="--btn-color: var(--color-success, #10b981)"
              :disabled="loading"
              @click="selectedPriority = 'low'"
            >
              <span class="ccm-priority-dot" style="background: var(--color-success, #10b981)" />
              Низкий
            </button>
            <button
              class="ccm-priority-btn"
              :class="{ 'ccm-priority-btn--active': selectedPriority === 'medium' }"
              style="--btn-color: var(--color-primary, #7c5cfc)"
              :disabled="loading"
              @click="selectedPriority = 'medium'"
            >
              <span class="ccm-priority-dot" style="background: var(--color-primary, #7c5cfc)" />
              Средний
            </button>
            <button
              class="ccm-priority-btn"
              :class="{ 'ccm-priority-btn--active': selectedPriority === 'high' }"
              style="--btn-color: #f59e0b"
              :disabled="loading"
              @click="selectedPriority = 'high'"
            >
              <span class="ccm-priority-dot" style="background: #f59e0b" />
              Высокий
            </button>
            <button
              class="ccm-priority-btn"
              :class="{ 'ccm-priority-btn--active': selectedPriority === 'critical' }"
              style="--btn-color: var(--color-danger, #ef4444)"
              :disabled="loading"
              @click="selectedPriority = 'critical'"
            >
              <span class="ccm-priority-dot" style="background: var(--color-danger, #ef4444)" />
              Крит.
            </button>
          </div>
        </div>

        <div class="ccm-meta-group">
          <span class="ccm-meta-label">Тип</span>
          <div class="ccm-type-btns">
            <button
              class="ccm-type-btn"
              :class="{ 'ccm-type-btn--active': selectedTaskType === 'task' }"
              :disabled="loading"
              title="Задача"
              @click="selectedTaskType = 'task'"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
            </button>
            <button
              class="ccm-type-btn"
              :class="{ 'ccm-type-btn--active': selectedTaskType === 'bug' }"
              :disabled="loading"
              title="Баг"
              @click="selectedTaskType = 'bug'"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M8 2l1.88 1.88M14.12 3.88L16 2M9 7.13v-1a3.003 3.003 0 1 1 6 0v1"/>
                <path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6"/>
                <path d="M12 20v-9M6.53 9C4.6 8.8 3 7.1 3 5M6 13H2M6 17l-4 1M17.47 9c1.93-.2 3.53-1.9 3.53-4M18 13h4M18 17l4 1"/>
              </svg>
            </button>
            <button
              class="ccm-type-btn"
              :class="{ 'ccm-type-btn--active': selectedTaskType === 'feature' }"
              :disabled="loading"
              title="Фича"
              @click="selectedTaskType = 'feature'"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
            </button>
            <button
              class="ccm-type-btn"
              :class="{ 'ccm-type-btn--active': selectedTaskType === 'improvement' }"
              :disabled="loading"
              title="Улучшение"
              @click="selectedTaskType = 'improvement'"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/></svg>
            </button>
          </div>
        </div>

        <div class="ccm-meta-group">
          <span class="ccm-meta-label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
            Дедлайн
          </span>
          <div class="ccm-date-wrap">
            <input
              v-model="selectedDueDate"
              type="date"
              class="ccm-date-input"
              :disabled="loading"
            />
            <button
              v-if="selectedDueDate"
              class="ccm-date-clear"
              type="button"
              @click="selectedDueDate = ''"
            >
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Files -->
      <div class="ccm-section">
        <div class="ccm-section__label">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
          </svg>
          Вложения
          <span v-if="pendingFiles.length" class="ccm-badge">{{ pendingFiles.length }}</span>
        </div>

        <label
          class="ccm-upload"
          :class="{ 'ccm-upload--drag': isDragging }"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="handleDrop"
        >
          <input type="file" class="ccm-upload__input" multiple :disabled="loading" @change="handleFileSelect" />
          <div class="ccm-upload__content">
            <div class="ccm-upload__icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline points="17 8 12 3 7 8" /><line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            </div>
            <span class="ccm-upload__text">
              {{ isDragging ? 'Отпустите для загрузки' : 'Перетащите файлы или нажмите' }}
            </span>
          </div>
        </label>

        <!-- Image grid -->
        <div v-if="pendingFiles.some(f => f.previewUrl)" class="ccm-previews">
          <div v-for="(pf, i) in pendingFiles" :key="i" class="ccm-preview" v-show="pf.previewUrl">
            <img :src="pf.previewUrl!" :alt="pf.file.name" class="ccm-preview__img" />
            <div class="ccm-preview__overlay">
              <span class="ccm-preview__name">{{ pf.file.name }}</span>
              <span class="ccm-preview__size">{{ formatFileSize(pf.file.size) }}</span>
            </div>
            <button class="ccm-preview__remove" @click="removePendingFile(i)">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" /></svg>
            </button>
          </div>
        </div>

        <!-- Non-image files -->
        <div v-if="pendingFiles.some(f => !f.previewUrl)" class="ccm-filelist">
          <div v-for="(pf, i) in pendingFiles" :key="i" class="ccm-fileitem" v-show="!pf.previewUrl">
            <div class="ccm-fileitem__icon">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /></svg>
            </div>
            <div class="ccm-fileitem__body">
              <span class="ccm-fileitem__name">{{ pf.file.name }}</span>
              <span class="ccm-fileitem__meta">{{ formatFileSize(pf.file.size) }}</span>
            </div>
            <button class="ccm-fileitem__remove" @click="removePendingFile(i)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" /></svg>
            </button>
          </div>
        </div>

        <p v-if="pendingFiles.length" class="ccm-upload-hint">Файлы загрузятся после создания карточки</p>
      </div>
    </div>

    <template #footer>
      <BaseButton variant="secondary" :disabled="loading" @click="handleClose">
        Отмена
      </BaseButton>
      <BaseButton :loading="loading" :disabled="!title.trim()" @click="handleCreate">
        Создать карточку
      </BaseButton>
    </template>
  </BaseModal>
</template>

<style scoped>
/* Header */
.ccm-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ccm-header__icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--color-primary-soft);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ccm-header__title {
  font-size: var(--font-size-lg);
  font-weight: 700;
  letter-spacing: -0.02em;
  margin: 0;
}

/* Body */
.ccm-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Sections with icon labels */
.ccm-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ccm-section__label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-tertiary);
}

.ccm-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: var(--color-primary);
  color: white;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: none;
}

/* Details row */
.ccm-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ccm-details__label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-tertiary);
}

.ccm-assignee {
  position: relative;
}

/* Upload */
.ccm-upload {
  display: block;
  cursor: pointer;
}

.ccm-upload__input {
  display: none;
}

.ccm-upload__content {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border: 1.5px dashed var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-tertiary);
  transition: all 0.2s;
}

.ccm-upload__icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: var(--color-surface-alt);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
}

.ccm-upload__text {
  font-size: 13px;
  font-weight: 500;
}

.ccm-upload:hover .ccm-upload__content,
.ccm-upload--drag .ccm-upload__content {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.ccm-upload:hover .ccm-upload__icon,
.ccm-upload--drag .ccm-upload__icon {
  background: var(--color-primary);
  color: white;
}

.ccm-upload-hint {
  font-size: 11px;
  color: var(--color-text-tertiary);
  margin: 0;
  font-style: italic;
}

/* Image previews */
.ccm-previews {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 8px;
}

.ccm-preview {
  position: relative;
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--color-border-light);
  background: var(--color-surface-alt);
  transition: transform 0.2s, box-shadow 0.2s;
}

.ccm-preview:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.ccm-preview__img {
  width: 100%;
  aspect-ratio: 4 / 3;
  object-fit: cover;
  display: block;
}

.ccm-preview__overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 20px 8px 6px;
  background: linear-gradient(to top, rgba(0, 0, 0, 0.6), transparent);
  display: flex;
  flex-direction: column;
  gap: 1px;
  opacity: 0;
  transition: opacity 0.2s;
}

.ccm-preview:hover .ccm-preview__overlay {
  opacity: 1;
}

.ccm-preview__name {
  font-size: 11px;
  font-weight: 500;
  color: white;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ccm-preview__size {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.7);
}

.ccm-preview__remove {
  position: absolute;
  top: 5px;
  right: 5px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(4px);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: all 0.15s;
}

.ccm-preview:hover .ccm-preview__remove {
  opacity: 1;
}

.ccm-preview__remove:hover {
  background: var(--color-danger);
  transform: scale(1.1);
}

/* Non-image files */
.ccm-filelist {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ccm-fileitem {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  background: var(--color-surface-alt);
  border: 1px solid var(--color-border-light);
}

.ccm-fileitem__icon {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
  display: flex;
}

.ccm-fileitem__body {
  flex: 1;
  min-width: 0;
}

.ccm-fileitem__name {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ccm-fileitem__meta {
  font-size: 11px;
  color: var(--color-text-tertiary);
}

.ccm-fileitem__remove {
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.ccm-fileitem__remove:hover {
  color: var(--color-danger);
  background: var(--color-danger-soft);
}

/* Metadata row */
.ccm-meta-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.ccm-meta-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ccm-meta-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-tertiary);
}

.ccm-priority-btns {
  display: flex;
  gap: 4px;
}

.ccm-priority-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border: 1.5px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: var(--color-text-secondary);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.ccm-priority-btn:hover {
  border-color: var(--btn-color);
}

.ccm-priority-btn--active {
  border-color: var(--btn-color);
  background: color-mix(in srgb, var(--btn-color) 8%, transparent);
  color: var(--color-text);
}

.ccm-priority-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.ccm-type-btns {
  display: flex;
  gap: 4px;
}

.ccm-type-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1.5px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: var(--color-text-tertiary);
  cursor: pointer;
  transition: all 0.15s;
}

.ccm-type-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.ccm-type-btn--active {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.ccm-date-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
}

.ccm-date-input {
  flex: 1;
  padding: 8px 12px;
  border: 1.5px solid var(--color-input-border);
  border-radius: var(--radius-sm);
  background: var(--color-input-bg);
  color: var(--color-text);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: all 0.15s;
  cursor: pointer;
}

.ccm-date-input:focus {
  border-color: var(--color-input-focus);
  box-shadow: var(--shadow-focus);
}

.ccm-date-input::-webkit-calendar-picker-indicator {
  cursor: pointer;
  opacity: 0.5;
  transition: opacity 0.15s;
}

.ccm-date-input::-webkit-calendar-picker-indicator:hover {
  opacity: 1;
}

.ccm-date-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: var(--color-surface-alt);
  border-radius: 50%;
  color: var(--color-text-tertiary);
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;
}

.ccm-date-clear:hover {
  background: var(--color-danger-soft);
  color: var(--color-danger);
}
</style>

<!-- Deep overrides for child components in create context -->
<style>
.ccm-body .rte__content {
  min-height: 80px;
  max-height: 200px;
}
</style>
