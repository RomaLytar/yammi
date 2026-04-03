<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import type { MemberResponse } from '@/types/api'
import { useDebouncedSearch } from '@/composables/useDebouncedSearch'
import BaseSelect from '@/components/shared/BaseSelect.vue'
import type { SelectOption } from '@/components/shared/BaseSelect.vue'

interface Props {
  members: MemberResponse[]
  boardId: string
}

interface Emits {
  (e: 'update:filters', filters: BoardFilters): void
}

export interface BoardFilters {
  search: string
  assigneeIds: string[]
  priority: string
  taskType: string
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { searchInput, debouncedValue: debouncedSearch, clearSearch } = useDebouncedSearch(250)

const selectedAssignees = ref<Set<string>>(new Set())
const selectedPriority = ref('')
const selectedTaskType = ref('')

// --- localStorage persistence ---
const storageKey = computed(() => `board_filters_${props.boardId}`)

function saveFilters() {
  try {
    const data: Record<string, unknown> = {}
    if (selectedAssignees.value.size > 0) data.assigneeIds = [...selectedAssignees.value]
    if (selectedPriority.value) data.priority = selectedPriority.value
    if (selectedTaskType.value) data.taskType = selectedTaskType.value
    if (searchInput.value.trim()) data.search = searchInput.value.trim()

    if (Object.keys(data).length > 0) {
      localStorage.setItem(storageKey.value, JSON.stringify(data))
    } else {
      localStorage.removeItem(storageKey.value)
    }
  } catch { /* ignore */ }
}

function restoreFilters() {
  try {
    const raw = localStorage.getItem(storageKey.value)
    if (!raw) return
    const data = JSON.parse(raw)
    if (data.search) searchInput.value = data.search
    if (Array.isArray(data.assigneeIds)) {
      const valid = data.assigneeIds.filter((id: string) => props.members.some(m => m.user_id === id))
      selectedAssignees.value = new Set(valid)
    }
    // Legacy single assigneeId support
    if (data.assigneeId && props.members.some(m => m.user_id === data.assigneeId)) {
      selectedAssignees.value = new Set([data.assigneeId])
    }
    if (data.priority) selectedPriority.value = data.priority
    if (data.taskType) selectedTaskType.value = data.taskType
  } catch { /* ignore */ }
}

// Assignee avatars
const maxVisibleAvatars = 5
const showAllMembers = ref(false)
const memberDropdownRef = ref<HTMLElement | null>(null)

const assignees = computed(() => props.members || [])
const visibleAssignees = computed(() => assignees.value.slice(0, maxVisibleAvatars))
const overflowAssignees = computed(() => assignees.value.slice(maxVisibleAvatars))
const hasOverflow = computed(() => assignees.value.length > maxVisibleAvatars)

const hasActiveFilters = computed(() =>
  debouncedSearch.value !== '' ||
  selectedAssignees.value.size > 0 ||
  selectedPriority.value !== '' ||
  selectedTaskType.value !== ''
)

const priorityOptions: { value: string; label: string; color?: string }[] = [
  { value: '', label: 'Все приоритеты' },
  { value: 'critical', label: 'Критический', color: '#ef4444' },
  { value: 'high', label: 'Высокий', color: '#f59e0b' },
  { value: 'medium', label: 'Средний', color: '#7c5cfc' },
  { value: 'low', label: 'Низкий', color: '#10b981' },
]

const taskTypeOptions: { value: string; label: string; icon?: string }[] = [
  { value: '', label: 'Все типы' },
  { value: 'bug', label: 'Баг', icon: 'bug' },
  { value: 'feature', label: 'Фича', icon: 'star' },
  { value: 'task', label: 'Задача', icon: 'check' },
  { value: 'improvement', label: 'Улучшение', icon: 'arrow' },
]

// BaseSelect compatible options
const taskTypeSelectOptions: SelectOption[] = taskTypeOptions.map(o => ({ value: o.value, label: o.label }))
const prioritySelectOptions: SelectOption[] = priorityOptions.map(o => ({ value: o.value, label: o.label, color: o.color }))

function getInitial(member: MemberResponse): string {
  return (member.name || member.email || '?').charAt(0).toUpperCase()
}

function getMemberName(member: MemberResponse): string {
  return member.name || member.email || member.user_id.slice(0, 8)
}

function toggleAssignee(userId: string) {
  const next = new Set(selectedAssignees.value)
  if (next.has(userId)) next.delete(userId)
  else next.add(userId)
  selectedAssignees.value = next
}

function toggleOverflowDropdown() {
  showAllMembers.value = !showAllMembers.value
}

function resetFilters() {
  clearSearch()
  selectedAssignees.value = new Set()
  selectedPriority.value = ''
  selectedTaskType.value = ''
  try { localStorage.removeItem(storageKey.value) } catch { /* ignore */ }
}

// Если выбранные assignees больше не участники — убрать их
watch(() => props.members, (newMembers) => {
  const memberIds = new Set(newMembers.map(m => m.user_id))
  const filtered = new Set([...selectedAssignees.value].filter(id => memberIds.has(id)))
  if (filtered.size !== selectedAssignees.value.size) {
    selectedAssignees.value = filtered
  }
})

// Close dropdowns on outside click
function handleClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (memberDropdownRef.value && !memberDropdownRef.value.contains(target)) {
    showAllMembers.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  restoreFilters()
  if (searchInput.value || selectedAssignees.value.size > 0 || selectedPriority.value || selectedTaskType.value) {
    emit('update:filters', {
      search: searchInput.value.trim(),
      assigneeIds: [...selectedAssignees.value],
      priority: selectedPriority.value,
      taskType: selectedTaskType.value,
    })
  }
})
onUnmounted(() => document.removeEventListener('click', handleClickOutside))

// Emit + persist
watch([debouncedSearch, selectedAssignees, selectedPriority, selectedTaskType], () => {
  emit('update:filters', {
    search: debouncedSearch.value,
    assigneeIds: [...selectedAssignees.value],
    priority: selectedPriority.value,
    taskType: selectedTaskType.value,
  })
  saveFilters()
})
</script>

<template>
  <div class="filter-bar">
    <!-- Search input -->
    <div class="filter-bar__search">
      <svg class="filter-bar__search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
        <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
      </svg>
      <input
        v-model="searchInput"
        type="text"
        class="filter-bar__search-input"
        placeholder="Поиск по названию..."
        maxlength="200"
      />
      <button v-if="searchInput" class="filter-bar__search-clear" @click="clearSearch">&times;</button>
    </div>

    <!-- Assignee avatars (multi-select) -->
    <div class="filter-bar__assignees">
      <div
        v-for="member in visibleAssignees"
        :key="member.user_id"
        class="filter-bar__avatar"
        :class="{ 'filter-bar__avatar--active': selectedAssignees.has(member.user_id) }"
        :title="getMemberName(member)"
        @click="toggleAssignee(member.user_id)"
      >
        <img v-if="member.avatar_url" :src="member.avatar_url" :alt="getMemberName(member)" class="filter-bar__avatar-img" />
        <span v-else class="filter-bar__avatar-initial">{{ getInitial(member) }}</span>
        <span v-if="selectedAssignees.has(member.user_id)" class="filter-bar__avatar-check">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="3" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
        </span>
      </div>

      <!-- Overflow +N -->
      <div v-if="hasOverflow" ref="memberDropdownRef" class="filter-bar__overflow-wrap">
        <div
          class="filter-bar__avatar filter-bar__avatar--overflow"
          :class="{ 'filter-bar__avatar--active': overflowAssignees.some(m => selectedAssignees.has(m.user_id)) }"
          @click.stop="toggleOverflowDropdown"
        >
          +{{ overflowAssignees.length }}
        </div>

        <Transition name="dropdown">
          <div v-if="showAllMembers" class="filter-bar__dropdown">
            <div
              v-for="member in overflowAssignees"
              :key="member.user_id"
              class="filter-bar__dropdown-item"
              :class="{ 'filter-bar__dropdown-item--active': selectedAssignees.has(member.user_id) }"
              @click="toggleAssignee(member.user_id)"
            >
              <div class="filter-bar__dropdown-avatar">
                <img v-if="member.avatar_url" :src="member.avatar_url" :alt="getMemberName(member)" class="filter-bar__avatar-img" />
                <span v-else class="filter-bar__avatar-initial">{{ getInitial(member) }}</span>
              </div>
              <span class="filter-bar__dropdown-name">{{ getMemberName(member) }}</span>
              <svg v-if="selectedAssignees.has(member.user_id)" class="filter-bar__dropdown-check" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2.5" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- Task type -->
    <BaseSelect
      :model-value="selectedTaskType"
      :options="taskTypeSelectOptions"
      placeholder="Все типы"
      size="sm"
      @update:model-value="(v) => { selectedTaskType = String(v) }"
    />

    <!-- Priority -->
    <BaseSelect
      :model-value="selectedPriority"
      :options="prioritySelectOptions"
      placeholder="Все приоритеты"
      size="sm"
      @update:model-value="(v) => { selectedPriority = String(v) }"
    />

    <!-- Reset button -->
    <button v-if="hasActiveFilters" class="filter-bar__reset" @click="resetFilters">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
        <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
      </svg>
      Сбросить
    </button>
  </div>
</template>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 24px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border-light);
  flex-wrap: wrap;
}

/* Search */
.filter-bar__search {
  position: relative;
  display: flex;
  align-items: center;
  min-width: 200px;
  max-width: 280px;
  flex: 1;
}
.filter-bar__search-icon {
  position: absolute;
  left: 10px;
  color: var(--color-text-tertiary);
  pointer-events: none;
}
.filter-bar__search-input {
  width: 100%;
  padding: 7px 30px 7px 34px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  color: var(--color-text-primary);
  background: var(--color-input-bg, var(--color-surface-alt));
  outline: none;
  transition: border-color 0.15s;
}
.filter-bar__search-input::placeholder { color: var(--color-text-tertiary); }
.filter-bar__search-input:focus { border-color: var(--color-primary); }
.filter-bar__search-clear {
  position: absolute;
  right: 6px;
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  font-size: 18px;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}
.filter-bar__search-clear:hover { color: var(--color-text-primary); }

/* Assignee avatars */
.filter-bar__assignees {
  display: flex;
  align-items: center;
  gap: 4px;
}
.filter-bar__avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  border: 2.5px solid transparent;
  transition: all 0.15s;
  background: var(--gradient-primary);
  color: white;
  flex-shrink: 0;
  position: relative;
}
.filter-bar__avatar:hover {
  transform: scale(1.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
.filter-bar__avatar--active {
  border-color: #10b981;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.3);
}
.filter-bar__avatar-check {
  position: absolute;
  bottom: -3px;
  right: -3px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #10b981;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--color-surface);
}
.filter-bar__avatar--overflow {
  background: var(--color-surface-alt);
  color: var(--color-text-secondary);
  font-size: 11px;
  font-weight: 600;
  border: 2px solid var(--color-border);
}
.filter-bar__avatar--overflow:hover {
  background: var(--color-surface);
  border-color: var(--color-text-tertiary);
}
.filter-bar__avatar--overflow.filter-bar__avatar--active {
  border-color: #10b981;
}
.filter-bar__avatar-img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}
.filter-bar__avatar-initial { line-height: 1; }

/* Shared dropdown */
.filter-bar__overflow-wrap,
.filter-bar__dropdown-wrap {
  position: relative;
}
.filter-bar__dropdown {
  position: absolute;
  top: calc(100% + 6px);
  left: 50%;
  transform: translateX(-50%);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12), 0 2px 8px rgba(0, 0, 0, 0.06);
  padding: 4px;
  z-index: 50;
  min-width: 200px;
  max-height: 260px;
  overflow-y: auto;
}
.filter-bar__dropdown--select {
  min-width: 180px;
}
.filter-bar__dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.1s;
  font-size: 13px;
  color: var(--color-text-primary);
}
.filter-bar__dropdown-item:hover {
  background: var(--color-surface-alt);
}
.filter-bar__dropdown-item--active {
  background: rgba(16, 185, 129, 0.08);
}
.filter-bar__dropdown-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--gradient-primary);
  color: white;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
}
.filter-bar__dropdown-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.filter-bar__dropdown-check {
  flex-shrink: 0;
  margin-left: auto;
}

/* Custom dropdown trigger */
.filter-bar__dropdown-trigger {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 10px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
  background: var(--color-input-bg, var(--color-surface-alt));
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}
.filter-bar__dropdown-trigger:hover {
  border-color: var(--color-text-tertiary);
  color: var(--color-text-primary);
}
.filter-bar__dropdown-trigger--active {
  border-color: #10b981;
  color: var(--color-text-primary);
  background: rgba(16, 185, 129, 0.06);
}
.filter-bar__chevron {
  transition: transform 0.2s;
  color: var(--color-text-tertiary);
}
.filter-bar__chevron--open {
  transform: rotate(180deg);
}
.filter-bar__priority-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.filter-bar__opt-icon {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}
.filter-bar__opt-icon-placeholder {
  width: 14px;
  flex-shrink: 0;
}

/* Dropdown transition */
.dropdown-enter-active { transition: all 0.15s ease-out; }
.dropdown-leave-active { transition: all 0.1s ease-in; }
.dropdown-enter-from { opacity: 0; transform: translateX(-50%) translateY(-4px); }
.dropdown-leave-to { opacity: 0; transform: translateX(-50%) translateY(-4px); }

/* Reset button */
.filter-bar__reset {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 7px 12px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
  background: var(--color-surface-alt);
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}
.filter-bar__reset:hover {
  border-color: var(--color-danger, #dc2626);
  color: var(--color-danger, #dc2626);
  background: var(--color-danger-soft, rgba(239, 68, 68, 0.06));
}

/* Scrollbar for dropdown */
.filter-bar__dropdown::-webkit-scrollbar { width: 6px; }
.filter-bar__dropdown::-webkit-scrollbar-track { background: transparent; }
.filter-bar__dropdown::-webkit-scrollbar-thumb { background: var(--color-text-tertiary); border-radius: 3px; }
</style>
