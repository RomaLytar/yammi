<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { UserLabel } from '@/types/domain'
import * as boardsApi from '@/api/boards'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'

const emit = defineEmits<{ (e: 'close'): void }>()

const userLabels = ref<UserLabel[]>([])
const loading = ref(true)
const saving = ref(false)

// Create
const newName = ref('')
const newColor = ref('#7c5cfc')

// Edit
const editingId = ref<string | null>(null)
const editName = ref('')
const editColor = ref('')

// Delete
const deleteTarget = ref<UserLabel | null>(null)

const COLORS = [
  '#ef4444', '#f97316', '#f59e0b', '#eab308',
  '#84cc16', '#22c55e', '#10b981', '#14b8a6',
  '#06b6d4', '#3b82f6', '#6366f1', '#7c5cfc',
  '#8b5cf6', '#a855f7', '#d946ef', '#ec4899',
]

onMounted(async () => {
  try {
    userLabels.value = await boardsApi.listUserLabels()
  } catch (err) {
    console.error('Failed to load global labels:', err)
  } finally {
    loading.value = false
  }
})

async function handleCreate() {
  if (!newName.value.trim()) return
  saving.value = true
  try {
    const label = await boardsApi.createUserLabel(newName.value.trim(), newColor.value)
    userLabels.value.push(label)
    newName.value = ''
    newColor.value = '#7c5cfc'
  } catch (err) {
    console.error('Failed to create global label:', err)
  } finally {
    saving.value = false
  }
}

function startEdit(label: UserLabel) {
  editingId.value = label.id
  editName.value = label.name
  editColor.value = label.color
}

function cancelEdit() {
  editingId.value = null
}

async function saveEdit() {
  if (!editingId.value || !editName.value.trim()) return
  saving.value = true
  try {
    const updated = await boardsApi.updateUserLabel(editingId.value, editName.value.trim(), editColor.value)
    const idx = userLabels.value.findIndex(l => l.id === editingId.value)
    if (idx !== -1) userLabels.value[idx] = updated
    editingId.value = null
  } catch (err) {
    console.error('Failed to update global label:', err)
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  try {
    await boardsApi.deleteUserLabel(deleteTarget.value.id)
    userLabels.value = userLabels.value.filter(l => l.id !== deleteTarget.value!.id)
  } catch (err) {
    console.error('Failed to delete global label:', err)
  } finally {
    deleteTarget.value = null
  }
}
</script>

<template>
  <BaseModal size="medium" @close="emit('close')">
    <template #header>
      <div class="glm-header">
        <div class="glm-header__icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" /><line x1="7" y1="7" x2="7.01" y2="7" />
          </svg>
        </div>
        <div>
          <h2 class="glm-header__title">Глобальные метки</h2>
          <p class="glm-header__hint">Доступны на всех ваших досках</p>
        </div>
      </div>
    </template>

    <div class="glm-body">
      <div v-if="loading" class="glm-loading">
        <span>Загрузка...</span>
      </div>

      <template v-else>
        <!-- Label list -->
        <div class="glm-list">
          <div v-for="label in userLabels" :key="label.id" class="glm-row">
            <template v-if="editingId === label.id">
              <div class="glm-edit">
                <span class="glm-dot" :style="{ background: editColor }" />
                <input
                  v-model="editName"
                  class="glm-input"
                  placeholder="Название..."
                  @keyup.enter="saveEdit"
                  @keyup.escape="cancelEdit"
                />
                <div class="glm-colors">
                  <button
                    v-for="c in COLORS" :key="c"
                    class="glm-swatch"
                    :class="{ 'glm-swatch--active': editColor === c }"
                    :style="{ background: c }"
                    @click="editColor = c"
                  />
                </div>
                <div class="glm-edit__btns">
                  <BaseButton size="sm" :loading="saving" @click="saveEdit">Сохранить</BaseButton>
                  <BaseButton size="sm" variant="ghost" @click="cancelEdit">Отмена</BaseButton>
                </div>
              </div>
            </template>
            <template v-else>
              <span class="glm-dot" :style="{ background: label.color }" />
              <span class="glm-name">{{ label.name }}</span>
              <div class="glm-actions">
                <button class="glm-icon-btn" title="Редактировать" @click="startEdit(label)">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </button>
                <button class="glm-icon-btn glm-icon-btn--danger" title="Удалить" @click="deleteTarget = label">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                </button>
              </div>
            </template>
          </div>

          <div v-if="userLabels.length === 0" class="glm-empty">
            Глобальные метки ещё не созданы
          </div>
        </div>

        <!-- Add form -->
        <div class="glm-add">
          <div class="glm-add__row">
            <span class="glm-dot" :style="{ background: newColor }" />
            <input
              v-model="newName"
              class="glm-input"
              placeholder="Новая глобальная метка..."
              @keyup.enter="handleCreate"
            />
            <BaseButton size="sm" :loading="saving" :disabled="!newName.trim()" @click="handleCreate">
              Добавить
            </BaseButton>
          </div>
          <div class="glm-colors">
            <button
              v-for="c in COLORS" :key="c"
              class="glm-swatch"
              :class="{ 'glm-swatch--active': newColor === c }"
              :style="{ background: c }"
              @click="newColor = c"
            />
          </div>
        </div>
      </template>
    </div>

    <template #footer>
      <BaseButton variant="secondary" @click="emit('close')">Закрыть</BaseButton>
    </template>
  </BaseModal>

  <ConfirmModal
    v-if="deleteTarget"
    title="Удалить глобальную метку"
    :message="`Удалить метку «${deleteTarget.name}»? Она будет убрана со всех карточек на всех досках.`"
    confirm-text="Удалить"
    variant="danger"
    @confirm="confirmDelete"
    @cancel="deleteTarget = null"
  />
</template>

<style scoped>
.glm-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.glm-header__icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: var(--color-primary-soft);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.glm-header__title {
  font-size: var(--font-size-lg);
  font-weight: 700;
  margin: 0;
}
.glm-header__hint {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin: 2px 0 0;
}
.glm-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.glm-loading {
  text-align: center;
  padding: 32px;
  color: var(--color-text-tertiary);
}
.glm-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.glm-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  transition: background 0.15s;
}
.glm-row:hover {
  background: var(--color-surface-alt);
}
.glm-dot {
  width: 20px;
  height: 20px;
  border-radius: 6px;
  flex-shrink: 0;
}
.glm-name {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
}
.glm-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.15s;
}
.glm-row:hover .glm-actions {
  opacity: 1;
}
.glm-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  transition: all 0.15s;
}
.glm-icon-btn:hover {
  background: var(--color-surface-alt);
  color: var(--color-text-primary);
}
.glm-icon-btn--danger:hover {
  background: var(--color-danger-soft);
  color: var(--color-danger);
}
.glm-edit {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.glm-edit__btns {
  display: flex;
  gap: 8px;
}
.glm-input {
  flex: 1;
  padding: 8px 12px;
  border: 1.5px solid var(--color-input-border);
  border-radius: var(--radius-sm);
  background: var(--color-input-bg);
  color: var(--color-text);
  font-size: 14px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.15s;
}
.glm-input:focus {
  border-color: var(--color-input-focus);
}
.glm-colors {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.glm-swatch {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s;
}
.glm-swatch:hover {
  transform: scale(1.15);
}
.glm-swatch--active {
  border-color: var(--color-text);
  box-shadow: 0 0 0 2px var(--color-surface);
}
.glm-add {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--color-border-light);
}
.glm-add__row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.glm-empty {
  text-align: center;
  padding: 24px;
  color: var(--color-text-tertiary);
  font-size: 14px;
}
</style>
