<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBoardStore } from '@/stores/board'
import { useAuthStore } from '@/stores/auth'
import * as boardsApi from '@/api/boards'
import * as usersApi from '@/api/users'
import type { MemberResponse } from '@/types/api'
import type { Label, UserLabel, AutomationRule } from '@/types/domain'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import TemplateManager from '@/components/board/TemplateManager.vue'

const route = useRoute()
const router = useRouter()
const boardStore = useBoardStore()
const authStore = useAuthStore()

const boardId = route.params.boardId as string

// --- Tabs ---
type Tab = 'labels' | 'members' | 'automations' | 'templates' | 'settings'
const activeTab = ref<Tab>('labels')

// --- Loading ---
const pageLoading = ref(true)
const pageError = ref<string | null>(null)

// --- Board labels ---
const newLabelName = ref('')
const newLabelColor = ref('#7c5cfc')
const editingLabelId = ref<string | null>(null)
const editLabelName = ref('')
const editLabelColor = ref('')
const labelSaving = ref(false)
const deleteLabelTarget = ref<Label | null>(null)

// --- Global labels (managed separately in GlobalLabelsModal) ---

// --- Members ---
interface MemberInfo {
  userId: string
  role: 'owner' | 'member'
  name: string
  email: string
  avatarUrl: string
}

interface SearchUser {
  id: string
  email: string
  name: string
  avatarUrl: string
}

const membersList = ref<MemberInfo[]>([])
const membersLoading = ref(false)
const memberSearchQuery = ref('')
const memberSearchResults = ref<SearchUser[]>([])
const memberSearching = ref(false)
const addingMemberId = ref<string | null>(null)
let memberSearchTimer: ReturnType<typeof setTimeout> | null = null

// --- Automations ---
const automationRules = ref<AutomationRule[]>([])
const automationsLoading = ref(false)
const automationSaving = ref(false)
const deleteRuleTarget = ref<AutomationRule | null>(null)

const showNewRuleForm = ref(false)
const newRuleName = ref('')
const newRuleTriggerType = ref('card_moved_to_column')
const newRuleTriggerColumnId = ref('')
const newRuleActionType = ref('assign_member')
const newRuleActionUserId = ref('')
const newRuleActionLabelId = ref('')
const newRuleActionPriority = ref('high')
const newRuleActionColumnId = ref('')

const editingRuleId = ref<string | null>(null)
const editRuleName = ref('')
const editRuleTriggerType = ref('')
const editRuleTriggerColumnId = ref('')
const editRuleActionType = ref('')
const editRuleActionUserId = ref('')
const editRuleActionLabelId = ref('')
const editRuleActionPriority = ref('')
const editRuleActionColumnId = ref('')

const TRIGGER_TYPES = [
  { value: 'card_moved_to_column', label: 'Карточка перемещена в колонку' },
  { value: 'card_created', label: 'Карточка создана' },
]

const ACTION_TYPES = [
  { value: 'assign_member', label: 'Назначить исполнителя' },
  { value: 'add_label', label: 'Добавить метку' },
  { value: 'set_priority', label: 'Установить приоритет' },
  { value: 'move_card', label: 'Переместить в колонку' },
]

const PRIORITIES = [
  { value: 'low', label: 'Низкий' },
  { value: 'medium', label: 'Средний' },
  { value: 'high', label: 'Высокий' },
  { value: 'critical', label: 'Критический' },
]

// --- Settings ---
const useBoardLabelsOnly = ref(false)
const settingsSaving = ref(false)
const settingsSaved = ref(false)

const isOwner = computed(() => boardStore.board?.ownerId === authStore.userId)

const PRESET_COLORS = [
  '#ef4444', '#f97316', '#f59e0b', '#eab308',
  '#84cc16', '#22c55e', '#10b981', '#14b8a6',
  '#06b6d4', '#3b82f6', '#6366f1', '#7c5cfc',
  '#8b5cf6', '#a855f7', '#d946ef', '#ec4899',
]

onMounted(async () => {
  try {
    pageLoading.value = true
    pageError.value = null

    // Load board if not already loaded
    if (!boardStore.board || boardStore.boardId !== boardId) {
      await boardStore.fetchBoard(boardId)
    }

    // Load members
    const members = await boardsApi.getMembers(boardId)

    membersList.value = members.map((m: MemberResponse) => ({
      userId: m.user_id,
      role: m.role,
      name: m.name || 'Неизвестный',
      email: m.email || '',
      avatarUrl: m.avatar_url || '',
    }))

    useBoardLabelsOnly.value = boardStore.boardSettings?.useBoardLabelsOnly ?? false

    // Load automations
    try {
      automationRules.value = await boardsApi.listAutomationRules(boardId)
    } catch {
      automationRules.value = []
    }
  } catch (err) {
    pageError.value = 'Ошибка загрузки настроек доски'
    console.error('Failed to load board settings:', err)
  } finally {
    pageLoading.value = false
  }
})

// --- Board label actions ---

async function handleCreateBoardLabel() {
  if (!newLabelName.value.trim() || !newLabelColor.value) return
  labelSaving.value = true
  try {
    await boardStore.createBoardLabel(newLabelName.value.trim(), newLabelColor.value)
    newLabelName.value = ''
    newLabelColor.value = '#7c5cfc'
  } catch (err) {
    console.error('Failed to create label:', err)
  } finally {
    labelSaving.value = false
  }
}

function startEditLabel(label: Label) {
  editingLabelId.value = label.id
  editLabelName.value = label.name
  editLabelColor.value = label.color
}

function cancelEditLabel() {
  editingLabelId.value = null
  editLabelName.value = ''
  editLabelColor.value = ''
}

async function saveEditLabel() {
  if (!editingLabelId.value || !editLabelName.value.trim()) return
  labelSaving.value = true
  try {
    await boardStore.updateBoardLabel(editingLabelId.value, editLabelName.value.trim(), editLabelColor.value)
    editingLabelId.value = null
  } catch (err) {
    console.error('Failed to update label:', err)
  } finally {
    labelSaving.value = false
  }
}

async function confirmDeleteBoardLabel() {
  if (!deleteLabelTarget.value) return
  try {
    await boardStore.deleteBoardLabel(deleteLabelTarget.value.id)
  } catch (err) {
    console.error('Failed to delete label:', err)
  } finally {
    deleteLabelTarget.value = null
  }
}

// Global labels managed in GlobalLabelsModal (accessed from BoardListPage)

// --- Automations ---

function buildTriggerConfig(triggerType: string, columnId: string): Record<string, string> {
  if ((triggerType === 'card_moved_to_column' || triggerType === 'card_created') && columnId) {
    return { column_id: columnId }
  }
  return {}
}

function buildActionConfig(actionType: string, userId: string, labelId: string, priority: string, columnId: string): Record<string, string> {
  if (actionType === 'assign_member' && userId) return { user_id: userId }
  if (actionType === 'add_label' && labelId) return { label_id: labelId }
  if (actionType === 'set_priority' && priority) return { priority }
  if (actionType === 'move_card' && columnId) return { column_id: columnId }
  return {}
}

function resetNewRuleForm() {
  showNewRuleForm.value = false
  newRuleName.value = ''
  newRuleTriggerType.value = 'card_moved_to_column'
  newRuleTriggerColumnId.value = ''
  newRuleActionType.value = 'assign_member'
  newRuleActionUserId.value = ''
  newRuleActionLabelId.value = ''
  newRuleActionPriority.value = 'high'
  newRuleActionColumnId.value = ''
}

async function handleCreateRule() {
  if (!newRuleName.value.trim()) return
  automationSaving.value = true
  try {
    const rule = await boardsApi.createAutomationRule(boardId, {
      name: newRuleName.value.trim(),
      trigger_type: newRuleTriggerType.value,
      trigger_config: buildTriggerConfig(newRuleTriggerType.value, newRuleTriggerColumnId.value),
      action_type: newRuleActionType.value,
      action_config: buildActionConfig(newRuleActionType.value, newRuleActionUserId.value, newRuleActionLabelId.value, newRuleActionPriority.value, newRuleActionColumnId.value),
    })
    automationRules.value.push(rule)
    resetNewRuleForm()
  } catch (err) {
    console.error('Failed to create automation rule:', err)
  } finally {
    automationSaving.value = false
  }
}

function startEditRule(rule: AutomationRule) {
  editingRuleId.value = rule.id
  editRuleName.value = rule.name
  editRuleTriggerType.value = rule.triggerType
  editRuleTriggerColumnId.value = rule.triggerConfig?.column_id || ''
  editRuleActionType.value = rule.actionType
  editRuleActionUserId.value = rule.actionConfig?.user_id || ''
  editRuleActionLabelId.value = rule.actionConfig?.label_id || ''
  editRuleActionPriority.value = rule.actionConfig?.priority || 'high'
  editRuleActionColumnId.value = rule.actionConfig?.column_id || ''
}

function cancelEditRule() {
  editingRuleId.value = null
}

async function saveEditRule() {
  if (!editingRuleId.value || !editRuleName.value.trim()) return
  automationSaving.value = true
  try {
    const updated = await boardsApi.updateAutomationRule(boardId, editingRuleId.value, {
      name: editRuleName.value.trim(),
      trigger_type: editRuleTriggerType.value,
      trigger_config: buildTriggerConfig(editRuleTriggerType.value, editRuleTriggerColumnId.value),
      action_type: editRuleActionType.value,
      action_config: buildActionConfig(editRuleActionType.value, editRuleActionUserId.value, editRuleActionLabelId.value, editRuleActionPriority.value, editRuleActionColumnId.value),
    })
    const idx = automationRules.value.findIndex(r => r.id === editingRuleId.value)
    if (idx !== -1) automationRules.value[idx] = updated
    editingRuleId.value = null
  } catch (err) {
    console.error('Failed to update automation rule:', err)
  } finally {
    automationSaving.value = false
  }
}

async function toggleRuleEnabled(rule: AutomationRule) {
  try {
    const updated = await boardsApi.updateAutomationRule(boardId, rule.id, { name: rule.name, enabled: !rule.enabled })
    const idx = automationRules.value.findIndex(r => r.id === rule.id)
    if (idx !== -1) automationRules.value[idx] = updated
  } catch (err) {
    console.error('Failed to toggle rule:', err)
  }
}

async function confirmDeleteRule() {
  if (!deleteRuleTarget.value) return
  try {
    await boardsApi.deleteAutomationRule(boardId, deleteRuleTarget.value.id)
    automationRules.value = automationRules.value.filter(r => r.id !== deleteRuleTarget.value!.id)
  } catch (err) {
    console.error('Failed to delete automation rule:', err)
  } finally {
    deleteRuleTarget.value = null
  }
}

function getTriggerLabel(type: string): string {
  return TRIGGER_TYPES.find(t => t.value === type)?.label || type
}

function getActionLabel(type: string): string {
  return ACTION_TYPES.find(a => a.value === type)?.label || type
}

function getColumnName(columnId: string): string {
  const col = boardStore.columns.find(c => c.id === columnId)
  return col?.title || columnId.slice(0, 8)
}

function getMemberNameById(userId: string): string {
  return boardStore.getMemberName(userId)
}

function getLabelName(labelId: string): string {
  const label = boardStore.allAvailableLabels.find(l => l.id === labelId)
  return label?.name || labelId.slice(0, 8)
}

function getPriorityLabel(p: string): string {
  return PRIORITIES.find(pr => pr.value === p)?.label || p
}

function describeRule(rule: AutomationRule): string {
  let trigger = getTriggerLabel(rule.triggerType)
  if (rule.triggerConfig?.column_id) {
    trigger += ` "${getColumnName(rule.triggerConfig.column_id)}"`
  }

  let action = getActionLabel(rule.actionType)
  if (rule.actionType === 'assign_member' && rule.actionConfig?.user_id) {
    action += ` → ${getMemberNameById(rule.actionConfig.user_id)}`
  } else if (rule.actionType === 'add_label' && rule.actionConfig?.label_id) {
    action += ` → ${getLabelName(rule.actionConfig.label_id)}`
  } else if (rule.actionType === 'set_priority' && rule.actionConfig?.priority) {
    action += ` → ${getPriorityLabel(rule.actionConfig.priority)}`
  } else if (rule.actionType === 'move_card' && rule.actionConfig?.column_id) {
    action += ` → ${getColumnName(rule.actionConfig.column_id)}`
  }

  return `${trigger} → ${action}`
}

// --- Members ---

function handleMemberSearch(val: string) {
  memberSearchQuery.value = val
  if (memberSearchTimer) clearTimeout(memberSearchTimer)
  memberSearchResults.value = []

  if (val.trim().length < 3) return

  memberSearchTimer = setTimeout(async () => {
    memberSearching.value = true
    try {
      const results = await usersApi.searchByEmail(val.trim())
      const memberIds = new Set(membersList.value.map(m => m.userId))
      memberSearchResults.value = results.filter(u => !memberIds.has(u.id))
    } catch {
      memberSearchResults.value = []
    } finally {
      memberSearching.value = false
    }
  }, 300)
}

async function handleAddMember(user: SearchUser) {
  addingMemberId.value = user.id
  try {
    await boardsApi.addMember(boardId, { user_id: user.id, role: 'member' })
    memberSearchQuery.value = ''
    memberSearchResults.value = []
    // Reload members
    const members = await boardsApi.getMembers(boardId)
    membersList.value = members.map((m: MemberResponse) => ({
      userId: m.user_id,
      role: m.role,
      name: m.name || 'Неизвестный',
      email: m.email || '',
      avatarUrl: m.avatar_url || '',
    }))
  } catch (err) {
    console.error('Failed to add member:', err)
  } finally {
    addingMemberId.value = null
  }
}

async function handleRemoveMember(userId: string) {
  try {
    await boardsApi.removeMember(boardId, userId)
    membersList.value = membersList.value.filter(m => m.userId !== userId)
  } catch (err) {
    console.error('Failed to remove member:', err)
  }
}

function getInitials(name: string): string {
  return name.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) || '?'
}

// --- Settings ---

async function handleSaveSettings() {
  settingsSaving.value = true
  settingsSaved.value = false
  try {
    await boardStore.saveBoardSettings(useBoardLabelsOnly.value)
    settingsSaved.value = true
    setTimeout(() => { settingsSaved.value = false }, 2000)
  } catch (err) {
    console.error('Failed to save settings:', err)
  } finally {
    settingsSaving.value = false
  }
}

function goBack() {
  router.push(`/boards/${boardId}`)
}
</script>

<template>
  <div class="bsp-page">
    <!-- Header -->
    <div class="bsp-header">
      <div class="bsp-header__left">
        <button class="bsp-back-btn" @click="goBack" title="Назад к доске">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="19" y1="12" x2="5" y2="12" /><polyline points="12 19 5 12 12 5" />
          </svg>
        </button>
        <div>
          <h1 class="bsp-header__title">Настройки доски</h1>
          <p v-if="boardStore.board" class="bsp-header__board-name">{{ boardStore.board.title }}</p>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="pageLoading" class="bsp-loading">
      <BaseSpinner />
    </div>

    <!-- Error -->
    <div v-else-if="pageError" class="bsp-error">
      <p>{{ pageError }}</p>
      <BaseButton variant="secondary" @click="goBack">Вернуться к доске</BaseButton>
    </div>

    <!-- Content -->
    <div v-else class="bsp-content">
      <!-- Tabs -->
      <div class="bsp-tabs">
        <button
          class="bsp-tab"
          :class="{ 'bsp-tab--active': activeTab === 'labels' }"
          @click="activeTab = 'labels'"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
            <line x1="7" y1="7" x2="7.01" y2="7" />
          </svg>
          Метки
        </button>
        <button
          class="bsp-tab"
          :class="{ 'bsp-tab--active': activeTab === 'members' }"
          @click="activeTab = 'members'"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" />
            <path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
          </svg>
          Участники
        </button>
        <button
          class="bsp-tab"
          :class="{ 'bsp-tab--active': activeTab === 'automations' }"
          @click="activeTab = 'automations'"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="16 3 21 3 21 8" /><line x1="4" y1="20" x2="21" y2="3" />
            <polyline points="21 16 21 21 16 21" /><line x1="15" y1="15" x2="21" y2="21" />
            <line x1="4" y1="4" x2="9" y2="9" />
          </svg>
          Автоматизация
        </button>
        <button
          class="bsp-tab"
          :class="{ 'bsp-tab--active': activeTab === 'templates' }"
          @click="activeTab = 'templates'"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="3" width="18" height="18" rx="2" /><path d="M7 7h10M7 12h10M7 17h6" />
          </svg>
          Шаблоны
        </button>
        <button
          class="bsp-tab"
          :class="{ 'bsp-tab--active': activeTab === 'settings' }"
          @click="activeTab = 'settings'"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
          Настройки
        </button>
      </div>

      <!-- Tab: Labels -->
      <div v-if="activeTab === 'labels'" class="bsp-panel">
        <!-- Board labels -->
        <div class="bsp-section">
          <h2 class="bsp-section__title">Метки доски</h2>
          <p class="bsp-section__hint">Метки, доступные только на этой доске</p>

          <div class="bsp-label-list">
            <div v-for="label in boardStore.labels" :key="label.id" class="bsp-label-row">
              <template v-if="editingLabelId === label.id">
                <div class="bsp-label-edit">
                  <span class="bsp-label-dot" :style="{ background: editLabelColor }" />
                  <input
                    v-model="editLabelName"
                    class="bsp-label-input"
                    placeholder="Название метки..."
                    @keyup.enter="saveEditLabel"
                    @keyup.escape="cancelEditLabel"
                  />
                  <div class="bsp-color-picker-inline">
                    <button
                      v-for="color in PRESET_COLORS"
                      :key="color"
                      class="bsp-color-swatch"
                      :class="{ 'bsp-color-swatch--active': editLabelColor === color }"
                      :style="{ background: color }"
                      @click="editLabelColor = color"
                    />
                  </div>
                  <div class="bsp-label-edit__actions">
                    <BaseButton size="sm" :loading="labelSaving" @click="saveEditLabel">Сохранить</BaseButton>
                    <BaseButton size="sm" variant="ghost" @click="cancelEditLabel">Отмена</BaseButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <span class="bsp-label-dot" :style="{ background: label.color }" />
                <span class="bsp-label-name">{{ label.name }}</span>
                <div class="bsp-label-actions">
                  <button class="bsp-icon-btn" title="Редактировать" @click="startEditLabel(label)">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </button>
                  <button v-if="isOwner" class="bsp-icon-btn bsp-icon-btn--danger" title="Удалить" @click="deleteLabelTarget = label">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  </button>
                </div>
              </template>
            </div>

            <div v-if="boardStore.labels.length === 0" class="bsp-empty">
              Метки доски ещё не созданы
            </div>
          </div>

          <!-- Add board label -->
          <div class="bsp-add-form">
            <div class="bsp-add-form__row">
              <span class="bsp-label-dot" :style="{ background: newLabelColor }" />
              <input
                v-model="newLabelName"
                class="bsp-label-input"
                placeholder="Новая метка..."
                @keyup.enter="handleCreateBoardLabel"
              />
              <BaseButton size="sm" :loading="labelSaving" :disabled="!newLabelName.trim()" @click="handleCreateBoardLabel">
                Добавить
              </BaseButton>
            </div>
            <div class="bsp-color-picker-inline">
              <button
                v-for="color in PRESET_COLORS"
                :key="color"
                class="bsp-color-swatch"
                :class="{ 'bsp-color-swatch--active': newLabelColor === color }"
                :style="{ background: color }"
                @click="newLabelColor = color"
              />
            </div>
          </div>
        </div>

        <!-- Global labels note -->
        <div class="bsp-section">
          <p class="bsp-section__hint">Глобальные метки управляются отдельно — через иконку меток на странице списка досок.</p>
        </div>
      </div>

      <!-- Tab: Members -->
      <div v-if="activeTab === 'members'" class="bsp-panel">
        <div class="bsp-section">
          <h2 class="bsp-section__title">Участники доски</h2>

          <!-- Add member (owner only) -->
          <div v-if="isOwner" class="bsp-member-search">
            <div class="bsp-member-search__wrapper">
              <input
                :value="memberSearchQuery"
                type="text"
                class="bsp-member-search__input"
                placeholder="Введите email для поиска..."
                @input="handleMemberSearch(($event.target as HTMLInputElement).value)"
              />
              <div
                v-if="memberSearching || memberSearchResults.length > 0 || (!memberSearching && memberSearchQuery.trim().length >= 3 && memberSearchResults.length === 0)"
                class="bsp-member-search__dropdown"
              >
                <div v-if="memberSearching" class="bsp-member-search__hint">Поиск...</div>
                <div
                  v-for="user in memberSearchResults"
                  :key="user.id"
                  class="bsp-member-search__result"
                  :class="{ 'bsp-member-search__result--adding': addingMemberId === user.id }"
                  @click="handleAddMember(user)"
                >
                  <div class="bsp-member-avatar bsp-member-avatar--sm">
                    <img v-if="user.avatarUrl" :src="user.avatarUrl" :alt="user.name" />
                    <span v-else>{{ getInitials(user.name) }}</span>
                  </div>
                  <div>
                    <div class="bsp-member-name">{{ user.name }}</div>
                    <div class="bsp-member-email">{{ user.email }}</div>
                  </div>
                </div>
                <div v-if="!memberSearching && memberSearchQuery.trim().length >= 3 && memberSearchResults.length === 0" class="bsp-member-search__hint">
                  Никого не найдено
                </div>
              </div>
            </div>
          </div>

          <!-- Members list -->
          <div v-if="membersLoading" class="bsp-loading-inline">Загрузка...</div>
          <div v-else class="bsp-members-list">
            <div v-for="member in membersList" :key="member.userId" class="bsp-member-row">
              <div class="bsp-member-info">
                <div class="bsp-member-avatar">
                  <img v-if="member.avatarUrl" :src="member.avatarUrl" :alt="member.name" />
                  <span v-else>{{ getInitials(member.name) }}</span>
                </div>
                <div>
                  <div class="bsp-member-name">{{ member.name }}</div>
                  <div class="bsp-member-email">{{ member.email }}</div>
                </div>
              </div>
              <div class="bsp-member-row__actions">
                <span class="bsp-role-badge" :class="{ 'bsp-role-badge--owner': member.role === 'owner' }">
                  {{ member.role === 'owner' ? 'Владелец' : 'Участник' }}
                </span>
                <button
                  v-if="isOwner && member.role !== 'owner'"
                  class="bsp-remove-member-btn"
                  title="Удалить участника"
                  @click="handleRemoveMember(member.userId)"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab: Automations -->
      <div v-if="activeTab === 'automations'" class="bsp-panel">
        <div class="bsp-section">
          <h2 class="bsp-section__title">Правила автоматизации</h2>
          <p class="bsp-section__hint">
            Автоматические действия при событиях на доске. Например, при перемещении карточки в колонку — назначить исполнителя.
          </p>

          <!-- Rules list -->
          <div class="bsp-auto-list">
            <div v-for="rule in automationRules" :key="rule.id" class="bsp-auto-row">
              <!-- Edit mode -->
              <template v-if="editingRuleId === rule.id">
                <div class="bsp-auto-form">
                  <input v-model="editRuleName" class="bsp-label-input" placeholder="Название правила..." />

                  <div class="bsp-auto-form__row">
                    <span class="bsp-auto-form__label">Триггер</span>
                    <select v-model="editRuleTriggerType" class="bsp-auto-select">
                      <option v-for="t in TRIGGER_TYPES" :key="t.value" :value="t.value">{{ t.label }}</option>
                    </select>
                  </div>

                  <div v-if="editRuleTriggerType === 'card_moved_to_column' || editRuleTriggerType === 'card_created'" class="bsp-auto-form__row">
                    <span class="bsp-auto-form__label">Колонка</span>
                    <select v-model="editRuleTriggerColumnId" class="bsp-auto-select">
                      <option value="">Любая</option>
                      <option v-for="col in boardStore.columns" :key="col.id" :value="col.id">{{ col.title }}</option>
                    </select>
                  </div>

                  <div class="bsp-auto-form__row">
                    <span class="bsp-auto-form__label">Действие</span>
                    <select v-model="editRuleActionType" class="bsp-auto-select">
                      <option v-for="a in ACTION_TYPES" :key="a.value" :value="a.value">{{ a.label }}</option>
                    </select>
                  </div>

                  <div v-if="editRuleActionType === 'assign_member'" class="bsp-auto-form__row">
                    <span class="bsp-auto-form__label">Участник</span>
                    <select v-model="editRuleActionUserId" class="bsp-auto-select">
                      <option value="">Выберите...</option>
                      <option v-for="m in membersList" :key="m.userId" :value="m.userId">{{ m.name }}</option>
                    </select>
                  </div>
                  <div v-if="editRuleActionType === 'add_label'" class="bsp-auto-form__row">
                    <span class="bsp-auto-form__label">Метка</span>
                    <select v-model="editRuleActionLabelId" class="bsp-auto-select">
                      <option value="">Выберите...</option>
                      <option v-for="l in boardStore.allAvailableLabels" :key="l.id" :value="l.id">{{ l.name }}{{ l.isGlobal ? ' (глобальная)' : '' }}</option>
                    </select>
                  </div>
                  <div v-if="editRuleActionType === 'set_priority'" class="bsp-auto-form__row">
                    <span class="bsp-auto-form__label">Приоритет</span>
                    <select v-model="editRuleActionPriority" class="bsp-auto-select">
                      <option v-for="p in PRIORITIES" :key="p.value" :value="p.value">{{ p.label }}</option>
                    </select>
                  </div>
                  <div v-if="editRuleActionType === 'move_card'" class="bsp-auto-form__row">
                    <span class="bsp-auto-form__label">Колонка</span>
                    <select v-model="editRuleActionColumnId" class="bsp-auto-select">
                      <option value="">Выберите...</option>
                      <option v-for="col in boardStore.columns" :key="col.id" :value="col.id">{{ col.title }}</option>
                    </select>
                  </div>

                  <div class="bsp-label-edit__actions">
                    <BaseButton size="sm" :loading="automationSaving" @click="saveEditRule">Сохранить</BaseButton>
                    <BaseButton size="sm" variant="ghost" @click="cancelEditRule">Отмена</BaseButton>
                  </div>
                </div>
              </template>
              <!-- View mode -->
              <template v-else>
                <div class="bsp-auto-row__info">
                  <label class="bsp-toggle bsp-toggle--sm" @click.prevent="toggleRuleEnabled(rule)">
                    <span class="bsp-toggle__track bsp-toggle__track--sm" :class="{ 'bsp-toggle__track--active': rule.enabled }">
                      <span class="bsp-toggle__thumb" />
                    </span>
                  </label>
                  <div class="bsp-auto-row__text">
                    <span class="bsp-auto-row__name" :class="{ 'bsp-auto-row__name--disabled': !rule.enabled }">{{ rule.name }}</span>
                    <span class="bsp-auto-row__desc">{{ describeRule(rule) }}</span>
                  </div>
                </div>
                <div class="bsp-label-actions">
                  <button v-if="isOwner" class="bsp-icon-btn" title="Редактировать" @click="startEditRule(rule)">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </button>
                  <button v-if="isOwner" class="bsp-icon-btn bsp-icon-btn--danger" title="Удалить" @click="deleteRuleTarget = rule">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  </button>
                </div>
              </template>
            </div>

            <div v-if="automationRules.length === 0 && !showNewRuleForm" class="bsp-empty">
              Правила автоматизации ещё не созданы
            </div>
          </div>

          <!-- New rule form -->
          <div v-if="showNewRuleForm" class="bsp-add-form">
            <input v-model="newRuleName" class="bsp-label-input" placeholder="Название правила..." @keyup.escape="resetNewRuleForm" />

            <div class="bsp-auto-form__row">
              <span class="bsp-auto-form__label">Триггер</span>
              <select v-model="newRuleTriggerType" class="bsp-auto-select">
                <option v-for="t in TRIGGER_TYPES" :key="t.value" :value="t.value">{{ t.label }}</option>
              </select>
            </div>

            <div v-if="newRuleTriggerType === 'card_moved_to_column' || newRuleTriggerType === 'card_created'" class="bsp-auto-form__row">
              <span class="bsp-auto-form__label">Колонка</span>
              <select v-model="newRuleTriggerColumnId" class="bsp-auto-select">
                <option value="">Любая</option>
                <option v-for="col in boardStore.columns" :key="col.id" :value="col.id">{{ col.title }}</option>
              </select>
            </div>

            <div class="bsp-auto-form__row">
              <span class="bsp-auto-form__label">Действие</span>
              <select v-model="newRuleActionType" class="bsp-auto-select">
                <option v-for="a in ACTION_TYPES" :key="a.value" :value="a.value">{{ a.label }}</option>
              </select>
            </div>

            <div v-if="newRuleActionType === 'assign_member'" class="bsp-auto-form__row">
              <span class="bsp-auto-form__label">Участник</span>
              <select v-model="newRuleActionUserId" class="bsp-auto-select">
                <option value="">Выберите...</option>
                <option v-for="m in membersList" :key="m.userId" :value="m.userId">{{ m.name }}</option>
              </select>
            </div>
            <div v-if="newRuleActionType === 'add_label'" class="bsp-auto-form__row">
              <span class="bsp-auto-form__label">Метка</span>
              <select v-model="newRuleActionLabelId" class="bsp-auto-select">
                <option value="">Выберите...</option>
                <option v-for="l in boardStore.allAvailableLabels" :key="l.id" :value="l.id">{{ l.name }}{{ l.isGlobal ? ' (глобальная)' : '' }}</option>
              </select>
            </div>
            <div v-if="newRuleActionType === 'set_priority'" class="bsp-auto-form__row">
              <span class="bsp-auto-form__label">Приоритет</span>
              <select v-model="newRuleActionPriority" class="bsp-auto-select">
                <option v-for="p in PRIORITIES" :key="p.value" :value="p.value">{{ p.label }}</option>
              </select>
            </div>
            <div v-if="newRuleActionType === 'move_card'" class="bsp-auto-form__row">
              <span class="bsp-auto-form__label">Колонка</span>
              <select v-model="newRuleActionColumnId" class="bsp-auto-select">
                <option value="">Выберите...</option>
                <option v-for="col in boardStore.columns" :key="col.id" :value="col.id">{{ col.title }}</option>
              </select>
            </div>

            <div class="bsp-label-edit__actions">
              <BaseButton size="sm" :loading="automationSaving" :disabled="!newRuleName.trim()" @click="handleCreateRule">Создать</BaseButton>
              <BaseButton size="sm" variant="ghost" @click="resetNewRuleForm">Отмена</BaseButton>
            </div>
          </div>

          <!-- Add button -->
          <div v-if="isOwner && !showNewRuleForm" style="margin-top: 16px;">
            <BaseButton variant="secondary" size="sm" @click="showNewRuleForm = true">
              Добавить правило
            </BaseButton>
          </div>
        </div>
      </div>

      <!-- Tab: Templates -->
      <div v-if="activeTab === 'templates'" class="bsp-panel">
        <TemplateManager />
      </div>

      <!-- Tab: Settings -->
      <div v-if="activeTab === 'settings'" class="bsp-panel">
        <div class="bsp-section">
          <h2 class="bsp-section__title">Общие настройки</h2>

          <div class="bsp-setting-row">
            <div class="bsp-setting-info">
              <span class="bsp-setting-label">Использовать только метки этой доски</span>
              <span class="bsp-setting-desc">
                Когда включено, при выборе меток для карточек будут показаны только метки этой доски, без глобальных меток.
              </span>
            </div>
            <label class="bsp-toggle" @click.prevent="useBoardLabelsOnly = !useBoardLabelsOnly">
              <span class="bsp-toggle__track" :class="{ 'bsp-toggle__track--active': useBoardLabelsOnly }">
                <span class="bsp-toggle__thumb" />
              </span>
            </label>
          </div>

          <div class="bsp-setting-actions">
            <BaseButton :loading="settingsSaving" @click="handleSaveSettings">
              Сохранить настройки
            </BaseButton>
            <Transition name="bsp-fade">
              <span v-if="settingsSaved" class="bsp-saved-msg">Сохранено</span>
            </Transition>
          </div>
        </div>
      </div>
    </div>

    <!-- Confirm modals -->
    <ConfirmModal
      v-if="deleteLabelTarget"
      title="Удалить метку"
      :message="`Удалить метку «${deleteLabelTarget.name}»? Она будет убрана со всех карточек.`"
      confirm-text="Удалить"
      variant="danger"
      @confirm="confirmDeleteBoardLabel"
      @cancel="deleteLabelTarget = null"
    />

    <ConfirmModal
      v-if="deleteRuleTarget"
      title="Удалить правило"
      :message="`Удалить правило «${deleteRuleTarget.name}»?`"
      confirm-text="Удалить"
      variant="danger"
      @confirm="confirmDeleteRule"
      @cancel="deleteRuleTarget = null"
    />

  </div>
</template>

<style scoped>
/* ===== Page layout ===== */
.bsp-page {
  min-height: 100vh;
  background: var(--gradient-auth-bg);
  position: relative;
}

.bsp-page::before {
  content: '';
  position: absolute;
  top: 20%; left: 10%;
  width: 500px; height: 500px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.06) 0%, transparent 70%);
  pointer-events: none;
}

/* ===== Header ===== */
.bsp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  position: relative;
  z-index: 1;
}

.bsp-header__left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.bsp-back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface-alt);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}
.bsp-back-btn:hover {
  border-color: var(--color-text-tertiary);
  color: var(--color-text-primary);
  background: var(--color-surface);
}

.bsp-header__title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  background: var(--gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.bsp-header__board-name {
  margin: 2px 0 0 0;
  font-size: 14px;
  color: var(--color-text-secondary);
}

/* ===== Loading / Error ===== */
.bsp-loading,
.bsp-error {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  color: var(--color-text-secondary);
}

.bsp-error {
  flex-direction: column;
  gap: 16px;
}

/* ===== Content ===== */
.bsp-content {
  max-width: 800px;
  margin: 0 auto;
  padding: 24px;
}

/* ===== Tabs ===== */
.bsp-tabs {
  display: flex;
  gap: 4px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 4px;
  margin-bottom: 24px;
}

.bsp-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  justify-content: center;
  padding: 10px 16px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}
.bsp-tab:hover {
  background: var(--color-surface-alt);
  color: var(--color-text-primary);
}
.bsp-tab--active {
  background: var(--color-primary-soft);
  color: var(--color-primary);
  font-weight: 600;
}

/* ===== Panel ===== */
.bsp-panel {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

/* ===== Section ===== */
.bsp-section {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 16px;
  padding: 24px;
}

.bsp-section__title {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--color-text-primary);
}

.bsp-section__hint {
  margin: 0 0 20px 0;
  font-size: 13px;
  color: var(--color-text-tertiary);
}

/* ===== Label list ===== */
.bsp-label-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 20px;
}

.bsp-label-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--color-surface-alt);
  transition: background 0.15s;
}
.bsp-label-row:hover {
  background: var(--color-bg-subtle);
}

.bsp-label-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.bsp-label-name {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bsp-label-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.15s;
}
.bsp-label-row:hover .bsp-label-actions {
  opacity: 1;
}

.bsp-icon-btn {
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
.bsp-icon-btn:hover {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}
.bsp-icon-btn--danger:hover {
  background: var(--color-danger-soft);
  color: var(--color-danger);
}

/* ===== Label edit ===== */
.bsp-label-edit {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.bsp-label-edit__actions {
  display: flex;
  gap: 8px;
}

.bsp-label-input {
  flex: 1;
  padding: 8px 12px;
  border: 1.5px solid var(--color-input-border);
  border-radius: 8px;
  background: var(--color-input-bg);
  color: var(--color-text);
  font-size: 14px;
  outline: none;
  transition: all 0.15s;
}
.bsp-label-input:focus {
  border-color: var(--color-input-focus);
  box-shadow: var(--shadow-focus);
}

/* ===== Color picker ===== */
.bsp-color-picker-inline {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.bsp-color-swatch {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s;
  padding: 0;
}
.bsp-color-swatch:hover {
  transform: scale(1.15);
}
.bsp-color-swatch--active {
  border-color: var(--color-text-primary);
  box-shadow: 0 0 0 2px var(--color-surface), 0 0 0 4px var(--color-text-tertiary);
  transform: scale(1.1);
}

/* ===== Add form ===== */
.bsp-add-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  background: var(--color-surface-alt);
  border: 1px dashed var(--color-border);
  border-radius: 12px;
}

.bsp-add-form__row {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* ===== Empty ===== */
.bsp-empty {
  padding: 16px;
  text-align: center;
  color: var(--color-text-tertiary);
  font-size: 14px;
}

/* ===== Members ===== */
.bsp-member-search {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--color-border);
}

.bsp-member-search__wrapper {
  position: relative;
}

.bsp-member-search__input {
  width: 100%;
  padding: 10px 14px;
  border: 1.5px solid var(--color-input-border);
  border-radius: 8px;
  background: var(--color-input-bg);
  color: var(--color-text);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}
.bsp-member-search__input:focus {
  border-color: var(--color-input-focus);
  box-shadow: var(--shadow-focus);
}

.bsp-member-search__dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  z-index: 10;
  padding: 4px;
  max-height: 200px;
  overflow-y: auto;
}

.bsp-member-search__hint {
  padding: 8px 10px;
  font-size: 13px;
  color: var(--color-text-tertiary);
}

.bsp-member-search__result {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
}
.bsp-member-search__result:hover {
  background: var(--color-primary-soft);
}
.bsp-member-search__result--adding {
  opacity: 0.5;
  pointer-events: none;
}

.bsp-members-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.bsp-member-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 14px;
  background: var(--color-surface-alt);
  border-radius: 10px;
  transition: background 0.15s;
}
.bsp-member-row:hover {
  background: var(--color-bg-subtle);
}

.bsp-member-info {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.bsp-member-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-primary-light);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  overflow: hidden;
  flex-shrink: 0;
}
.bsp-member-avatar--sm {
  width: 32px;
  height: 32px;
  font-size: 12px;
}
.bsp-member-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.bsp-member-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-primary);
}

.bsp-member-email {
  font-size: 12px;
  color: var(--color-text-tertiary);
}

.bsp-member-row__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.bsp-role-badge {
  font-size: 12px;
  padding: 3px 10px;
  border-radius: 12px;
  background: var(--color-surface);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}
.bsp-role-badge--owner {
  background: var(--color-primary-soft);
  color: var(--color-primary);
  border-color: transparent;
}

.bsp-remove-member-btn {
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
  opacity: 0;
}
.bsp-member-row:hover .bsp-remove-member-btn {
  opacity: 1;
}
.bsp-remove-member-btn:hover {
  background: var(--color-danger-soft);
  color: var(--color-danger);
}

.bsp-loading-inline {
  text-align: center;
  padding: 24px;
  color: var(--color-text-tertiary);
}

/* ===== Settings tab ===== */
.bsp-setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 16px 0;
  border-bottom: 1px solid var(--color-border-light);
}

.bsp-setting-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.bsp-setting-label {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.bsp-setting-desc {
  font-size: 13px;
  color: var(--color-text-tertiary);
  line-height: 1.5;
}

.bsp-toggle {
  cursor: pointer;
  user-select: none;
  flex-shrink: 0;
}

.bsp-toggle__track {
  display: block;
  width: 44px;
  height: 24px;
  background: var(--color-border);
  border-radius: 12px;
  position: relative;
  transition: background 0.2s;
}
.bsp-toggle__track--active {
  background: var(--color-primary);
}

.bsp-toggle__thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}
.bsp-toggle__track--active .bsp-toggle__thumb {
  transform: translateX(20px);
}

.bsp-setting-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 24px;
}

.bsp-saved-msg {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-success);
}

/* ===== Automation rules ===== */
.bsp-auto-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 16px;
}

.bsp-auto-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 10px;
  background: var(--color-surface-alt);
  transition: background 0.15s;
}
.bsp-auto-row:hover {
  background: var(--color-bg-subtle);
}
.bsp-auto-row:hover .bsp-label-actions {
  opacity: 1;
}

.bsp-auto-row__info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.bsp-auto-row__text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.bsp-auto-row__name {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.bsp-auto-row__name--disabled {
  opacity: 0.5;
}

.bsp-auto-row__desc {
  font-size: 12px;
  color: var(--color-text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bsp-auto-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.bsp-auto-form__row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bsp-auto-form__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
  min-width: 80px;
  flex-shrink: 0;
}

.bsp-auto-select {
  flex: 1;
  padding: 8px 12px;
  border: 1.5px solid var(--color-input-border);
  border-radius: 8px;
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
.bsp-auto-select:focus {
  border-color: var(--color-input-focus);
  box-shadow: var(--shadow-focus);
}

.bsp-toggle--sm {
  flex-shrink: 0;
}

.bsp-toggle__track--sm {
  width: 36px;
  height: 20px;
  border-radius: 10px;
}
.bsp-toggle__track--sm .bsp-toggle__thumb {
  width: 16px;
  height: 16px;
}
.bsp-toggle__track--sm.bsp-toggle__track--active .bsp-toggle__thumb {
  transform: translateX(16px);
}

.bsp-fade-enter-active,
.bsp-fade-leave-active {
  transition: opacity 0.3s;
}
.bsp-fade-enter-from,
.bsp-fade-leave-to {
  opacity: 0;
}
</style>
