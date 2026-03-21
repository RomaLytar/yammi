<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useBoardsStore } from '@/stores/boards'
import { useAuthStore } from '@/stores/auth'
import * as usersApi from '@/api/users'
import type { Board } from '@/types/domain'
import CreateBoardModal from '@/components/board/CreateBoardModal.vue'
import BoardDetailsModal from '@/components/board/BoardDetailsModal.vue'
import MembersModal from '@/components/board/MembersModal.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'
import { useBulkSelect } from '@/composables/useBulkSelect'
import { useDebouncedSearch } from '@/composables/useDebouncedSearch'

const router = useRouter()
const boardsStore = useBoardsStore()
const authStore = useAuthStore()

const showCreateModal = ref(false)
const creatingBoard = ref(false)

// Action menu
const activeMenu = ref<string | null>(null)

// Modals
const deleteTarget = ref<Board | null>(null)
const detailsTarget = ref<Board | null>(null)
const membersTarget = ref<Board | null>(null)
const showBulkDeleteConfirm = ref(false)

// Debounced search
const { searchInput, debouncedValue, clearSearch: clearSearchInput } = useDebouncedSearch(300)

watch(debouncedValue, (val) => {
  boardsStore.search = val
  reload()
})

function clearSearch() {
  clearSearchInput()
  boardsStore.search = ''
  reload()
}

// Bulk select
const { selectMode, selectedIds, selectedCount, toggleSelectMode, toggleSelect, toggleSelectAll: toggleSelectAllRaw, clearSelection } = useBulkSelect(
  (id: string) => {
    const board = boardsStore.boards.find(b => b.id === id)
    return !!board && board.ownerId === authStore.userId
  }
)

const selectableBoards = computed(() =>
  boardsStore.boards.filter(b => b.ownerId === authStore.userId)
)
const allSelectableSelected = computed(() =>
  selectableBoards.value.length > 0 && selectableBoards.value.every(b => selectedIds.value.has(b.id))
)

function toggleSelectAll() {
  toggleSelectAllRaw(boardsStore.boards.map(b => b.id))
}

// Owner profiles cache
const ownerProfiles = reactive<Record<string, { name: string; avatarUrl: string }>>({})

function isOwner(board: Board): boolean {
  return board.ownerId === authStore.userId
}

async function reload() {
  await boardsStore.fetchBoards(true)
  fetchOwnerProfiles()
}

function toggleOwnerOnly() {
  boardsStore.ownerOnly = !boardsStore.ownerOnly
  reload()
}

function setSortBy(val: 'updated_at' | 'created_at' | 'title') {
  boardsStore.sortBy = val
  reload()
}

async function handleBulkDelete() {
  if (selectedIds.value.size === 0) return
  try {
    await boardsStore.deleteBoards([...selectedIds.value])
    clearSelection()
  } catch (err) {
    console.error('Failed to batch delete:', err)
  } finally {
    showBulkDeleteConfirm.value = false
  }
}

onMounted(async () => {
  await boardsStore.fetchBoards(true)
  fetchOwnerProfiles()
  document.addEventListener('click', closeMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeMenu)
})

function closeMenu() {
  activeMenu.value = null
}

async function fetchOwnerProfiles() {
  const ownerIds = [...new Set(boardsStore.boards.map(b => b.ownerId))]
  for (const id of ownerIds) {
    if (ownerProfiles[id]) continue
    try {
      const profile = await usersApi.getProfile(id)
      ownerProfiles[id] = { name: profile.name, avatarUrl: profile.avatarUrl }
    } catch {
      ownerProfiles[id] = { name: 'Неизвестный', avatarUrl: '' }
    }
  }
}

function getInitials(name: string): string {
  return name.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) || '?'
}

async function handleCreateBoard(data: { title: string; description: string }) {
  try {
    creatingBoard.value = true
    const board = await boardsStore.createBoard(data.title, data.description)
    showCreateModal.value = false
    router.push(`/boards/${board.id}`)
  } catch (error) {
    console.error('Failed to create board:', error)
  } finally {
    creatingBoard.value = false
  }
}

function openBoard(boardId: string) {
  if (selectMode.value) return
  router.push(`/boards/${boardId}`)
}

async function loadMore() {
  if (!boardsStore.loading && boardsStore.hasMore) {
    await boardsStore.fetchBoards(false)
    fetchOwnerProfiles()
  }
}

function toggleMenu(boardId: string, e: Event) {
  e.stopPropagation()
  activeMenu.value = activeMenu.value === boardId ? null : boardId
}

function openDetails(board: Board, e: Event) {
  e.stopPropagation()
  activeMenu.value = null
  detailsTarget.value = board
}

function openMembers(board: Board, e: Event) {
  e.stopPropagation()
  activeMenu.value = null
  membersTarget.value = board
}

function confirmDelete(board: Board, e: Event) {
  e.stopPropagation()
  activeMenu.value = null
  deleteTarget.value = board
}

async function handleDeleteBoard() {
  if (!deleteTarget.value) return
  try {
    await boardsStore.deleteBoards([deleteTarget.value.id])
  } catch (err) {
    console.error('Failed to delete board:', err)
  } finally {
    deleteTarget.value = null
  }
}
</script>

<template>
  <div class="board-list-page">
    <div class="board-list-header">
      <div class="header-left">
        <h1>Мои доски</h1>
        <label class="toggle" @click.prevent="toggleOwnerOnly">
          <span class="toggle-track" :class="{ 'toggle-track--active': boardsStore.ownerOnly }">
            <span class="toggle-thumb" />
          </span>
          <span class="toggle-label">Только мои</span>
        </label>
        <div class="sort-group">
          <button class="sort-btn" :class="{ 'sort-btn--active': boardsStore.sortBy === 'updated_at' }" @click="setSortBy('updated_at')">По активности</button>
          <button class="sort-btn" :class="{ 'sort-btn--active': boardsStore.sortBy === 'created_at' }" @click="setSortBy('created_at')">По дате</button>
          <button class="sort-btn" :class="{ 'sort-btn--active': boardsStore.sortBy === 'title' }" @click="setSortBy('title')">По алфавиту</button>
        </div>
      </div>
      <BaseButton @click="showCreateModal = true">+ Создать доску</BaseButton>
    </div>

    <div class="board-toolbar">
      <div class="search-box">
        <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        <input v-model="searchInput" type="text" placeholder="Поиск по названию..." class="search-field" />
        <button v-if="searchInput" class="search-clear" @click="clearSearch">&times;</button>
      </div>

      <button class="bulk-toggle" :class="{ 'bulk-toggle--active': selectMode }" @click="toggleSelectMode">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 11 12 14 22 4" />
          <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />
        </svg>
        Выбрать
      </button>

      <template v-if="selectMode">
        <button class="bulk-btn bulk-btn--select-all" @click="toggleSelectAll">
          {{ allSelectableSelected ? 'Снять все' : 'Выбрать все' }}
        </button>
        <button
          v-if="selectedCount > 0"
          class="bulk-btn bulk-btn--delete"
          @click="showBulkDeleteConfirm = true"
        >
          Удалить ({{ selectedCount }})
        </button>
      </template>
    </div>

    <div v-if="boardsStore.loading && boardsStore.boards.length === 0" class="board-list-empty">
      <BaseSpinner />
    </div>

    <div v-else-if="boardsStore.error" class="board-list-error">
      <p>{{ boardsStore.error }}</p>
      <BaseButton variant="secondary" @click="boardsStore.fetchBoards(true)">Повторить</BaseButton>
    </div>

    <div v-else-if="boardsStore.boards.length === 0" class="board-list-empty">
      <div class="empty-state">
        <div class="empty-state__icon">
          <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="3" width="7" height="7" rx="1" /><rect x="14" y="3" width="7" height="7" rx="1" />
            <rect x="3" y="14" width="7" height="7" rx="1" /><rect x="14" y="14" width="7" height="7" rx="1" />
          </svg>
        </div>
        <h2>У вас пока нет досок</h2>
        <p>Создайте первую доску для управления задачами</p>
        <BaseButton @click="showCreateModal = true">Создать доску</BaseButton>
      </div>
    </div>

    <div v-else class="board-list">
      <div
        v-for="board in boardsStore.boards"
        :key="board.id"
        class="board-item"
        :class="{
          'board-item--selected': selectMode && selectedIds.has(board.id),
          'board-item--select-mode': selectMode,
        }"
        @click="selectMode && isOwner(board) ? toggleSelect(board.id, $event) : openBoard(board.id)"
      >
        <!-- Checkbox for bulk select -->
        <div v-if="selectMode" class="board-checkbox" @click.stop>
          <input
            v-if="isOwner(board)"
            type="checkbox"
            :checked="selectedIds.has(board.id)"
            @change="toggleSelect(board.id, $event)"
          />
          <span v-else class="checkbox-disabled" title="Только владелец может удалить" />
        </div>

        <div class="board-item__body">
          <div class="board-item__header">
            <h3>{{ board.title }}</h3>
            <button v-if="!selectMode" class="dots-button" @click="toggleMenu(board.id, $event)" title="Действия">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <circle cx="8" cy="3" r="1.5" /><circle cx="8" cy="8" r="1.5" /><circle cx="8" cy="13" r="1.5" />
              </svg>
            </button>

            <Transition name="dropdown">
              <div v-if="activeMenu === board.id" class="dropdown-menu" @click.stop>
                <button class="dropdown-item" @click="openDetails(board, $event)">
                  <svg class="dropdown-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="3" y="3" width="18" height="18" rx="2" /><line x1="9" y1="3" x2="9" y2="21" /><line x1="3" y1="9" x2="21" y2="9" />
                  </svg>
                  Детали
                </button>
                <button class="dropdown-item" @click="openMembers(board, $event)">
                  <svg class="dropdown-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" />
                    <path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
                  </svg>
                  Участники
                </button>
                <template v-if="isOwner(board)">
                  <div class="dropdown-divider"></div>
                  <button class="dropdown-item dropdown-item--danger" @click="confirmDelete(board, $event)">
                    <svg class="dropdown-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                      <polyline points="3 6 5 6 21 6" /><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                    </svg>
                    Удалить
                  </button>
                </template>
              </div>
            </Transition>
          </div>

          <p v-if="board.description" class="board-item__description">{{ board.description }}</p>

          <div class="board-item__footer">
            <span class="board-item__date">{{ new Date(board.createdAt).toLocaleDateString('ru-RU') }}</span>
            <div v-if="ownerProfiles[board.ownerId]" class="owner-avatar" :title="ownerProfiles[board.ownerId].name">
              <img v-if="ownerProfiles[board.ownerId].avatarUrl" :src="ownerProfiles[board.ownerId].avatarUrl" :alt="ownerProfiles[board.ownerId].name" />
              <span v-else class="avatar-initials">{{ getInitials(ownerProfiles[board.ownerId].name) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="boardsStore.hasMore" class="board-list-footer">
        <BaseButton variant="secondary" :loading="boardsStore.loading" @click="loadMore">Загрузить ещё</BaseButton>
      </div>
    </div>

    <!-- Modals -->
    <CreateBoardModal v-if="showCreateModal" @close="showCreateModal = false" @create="handleCreateBoard" />

    <ConfirmModal
      v-if="deleteTarget"
      title="Удалить доску"
      :message="`Вы уверены, что хотите удалить доску «${deleteTarget.title}»? Все колонки и карточки будут удалены безвозвратно.`"
      confirm-text="Удалить"
      variant="danger"
      @confirm="handleDeleteBoard"
      @cancel="deleteTarget = null"
    />

    <ConfirmModal
      v-if="showBulkDeleteConfirm"
      title="Удалить выбранные доски"
      :message="`Удалить ${selectedCount} ${selectedCount === 1 ? 'доску' : 'досок'}? Все колонки и карточки будут удалены безвозвратно.`"
      confirm-text="Удалить все"
      variant="danger"
      @confirm="handleBulkDelete"
      @cancel="showBulkDeleteConfirm = false"
    />

    <BoardDetailsModal v-if="detailsTarget" :board-id="detailsTarget.id" :board-title="detailsTarget.title" @close="detailsTarget = null" />
    <MembersModal v-if="membersTarget" :board-id="membersTarget.id" :is-owner="membersTarget.ownerId === authStore.userId" @close="membersTarget = null" @updated="fetchOwnerProfiles()" />
  </div>
</template>

<style scoped>
.board-list-page {
  padding: 24px var(--space-lg);
  min-height: 100vh;
  background: var(--gradient-auth-bg);
  position: relative;
}
.board-list-page::before {
  content: '';
  position: absolute;
  top: 20%; left: 10%;
  width: 500px; height: 500px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.06) 0%, transparent 70%);
  pointer-events: none;
}
.board-list-page::after {
  content: '';
  position: absolute;
  bottom: 10%; right: 10%;
  width: 400px; height: 400px;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.05) 0%, transparent 70%);
  pointer-events: none;
}

.board-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}
.board-list-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
  color: var(--color-text-primary, #111827);
}

/* Toggle */
.toggle { display: flex; align-items: center; gap: 8px; cursor: pointer; user-select: none; }
.toggle-track { width: 36px; height: 20px; background: #d1d5db; border-radius: 10px; position: relative; transition: background 0.2s; }
.toggle-track--active { background: #3b82f6; }
.toggle-thumb { position: absolute; top: 2px; left: 2px; width: 16px; height: 16px; background: white; border-radius: 50%; transition: transform 0.2s; }
.toggle-track--active .toggle-thumb { transform: translateX(16px); }
.toggle-label { font-size: 13px; color: #6b7280; white-space: nowrap; }

/* Sort */
.sort-group { display: flex; gap: 4px; background: rgba(255,255,255,0.5); border-radius: 8px; padding: 2px; border: 1px solid #e5e7eb; }
.sort-btn { padding: 6px 12px; border: none; border-radius: 6px; font-size: 13px; color: #6b7280; background: transparent; cursor: pointer; white-space: nowrap; transition: all 0.15s; }
.sort-btn:hover { color: #374151; background: rgba(255,255,255,0.8); }
.sort-btn--active { background: white; color: #111827; font-weight: 500; box-shadow: 0 1px 2px rgba(0,0,0,0.06); }

/* Toolbar */
.board-toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }

.search-box { position: relative; flex: 1; max-width: 320px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #9ca3af; pointer-events: none; }
.search-field { width: 100%; padding: 8px 32px 8px 36px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 14px; outline: none; background: rgba(255,255,255,0.7); transition: border-color 0.2s; box-sizing: border-box; }
.search-field:focus { border-color: #3b82f6; background: white; }
.search-clear { position: absolute; right: 8px; top: 50%; transform: translateY(-50%); background: none; border: none; color: #9ca3af; font-size: 18px; cursor: pointer; padding: 0 4px; line-height: 1; }
.search-clear:hover { color: #374151; }

/* Bulk */
.bulk-toggle { display: flex; align-items: center; gap: 6px; padding: 7px 14px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 13px; color: #6b7280; background: rgba(255,255,255,0.7); cursor: pointer; white-space: nowrap; transition: all 0.15s; }
.bulk-toggle:hover { border-color: #9ca3af; color: #374151; }
.bulk-toggle--active { border-color: #3b82f6; color: #3b82f6; background: #eff6ff; }
.bulk-btn { padding: 7px 14px; border: none; border-radius: 8px; font-size: 13px; cursor: pointer; white-space: nowrap; transition: all 0.15s; }
.bulk-btn--select-all { background: rgba(255,255,255,0.7); color: #374151; border: 1px solid #d1d5db; }
.bulk-btn--select-all:hover { background: white; }
.bulk-btn--delete { background: #dc2626; color: white; }
.bulk-btn--delete:hover { background: #b91c1c; }

/* Board grid */
.board-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 20px; }
.board-item { background: var(--color-surface, #fff); border: 1px solid var(--color-border, #e5e7eb); border-radius: 12px; padding: 20px; cursor: pointer; transition: all 0.2s; position: relative; display: flex; gap: 12px; }
.board-item:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.1); border-color: var(--color-primary, #3b82f6); }
.board-item--select-mode { cursor: default; }
.board-item--selected { border-color: #3b82f6; background: #eff6ff; }
.board-item__body { flex: 1; min-width: 0; }

.board-checkbox { display: flex; align-items: flex-start; padding-top: 2px; }
.board-checkbox input[type="checkbox"] { width: 18px; height: 18px; accent-color: #3b82f6; cursor: pointer; }
.checkbox-disabled { width: 18px; height: 18px; border: 1.5px solid #d1d5db; border-radius: 4px; opacity: 0.3; }

.board-item__header { display: flex; justify-content: space-between; align-items: flex-start; position: relative; }
.board-item__header h3 { margin: 0; font-size: 18px; font-weight: 600; color: var(--color-text-primary, #111827); flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dots-button { background: none; border: none; color: #9ca3af; cursor: pointer; padding: 4px; border-radius: 6px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; transition: all 0.15s; }
.dots-button:hover { background: #f3f4f6; color: #374151; }

/* Dropdown */
.dropdown-menu { position: absolute; top: 100%; right: 0; margin-top: 4px; background: white; border: 1px solid #e5e7eb; border-radius: 8px; box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1); z-index: 50; min-width: 160px; padding: 4px; }
.dropdown-enter-active, .dropdown-leave-active { transition: all 0.15s ease; }
.dropdown-enter-from, .dropdown-leave-to { opacity: 0; transform: scale(0.95) translateY(-4px); }
.dropdown-item { display: flex; align-items: center; gap: 8px; width: 100%; padding: 8px 12px; background: none; border: none; border-radius: 6px; font-size: 14px; color: #374151; cursor: pointer; text-align: left; transition: background 0.1s; }
.dropdown-item:hover { background: #f3f4f6; }
.dropdown-item--danger { color: #dc2626; }
.dropdown-item--danger:hover { background: #fef2f2; }
.dropdown-icon { width: 16px; height: 16px; flex-shrink: 0; }
.dropdown-divider { height: 1px; background: #e5e7eb; margin: 4px 0; }

.board-item__description { margin: 12px 0 0 0; font-size: 14px; color: var(--color-text-secondary, #6b7280); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.board-item__footer { margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--color-border, #e5e7eb); display: flex; justify-content: space-between; align-items: center; }
.board-item__date { font-size: 12px; color: var(--color-text-tertiary, #9ca3af); }

/* Owner avatar */
.owner-avatar { width: 28px; height: 28px; border-radius: 50%; background: #dbeafe; color: #3b82f6; display: flex; align-items: center; justify-content: center; overflow: hidden; cursor: default; flex-shrink: 0; transition: transform 0.15s; }
.owner-avatar:hover { transform: scale(1.15); }
.owner-avatar img { width: 100%; height: 100%; object-fit: cover; }
.avatar-initials { font-size: 11px; font-weight: 600; line-height: 1; }

.board-list-empty, .board-list-error { display: flex; align-items: center; justify-content: center; min-height: 400px; }
.board-list-error { flex-direction: column; gap: 16px; }
.empty-state { text-align: center; max-width: 400px; }
.empty-state__icon { margin-bottom: 16px; display: flex; justify-content: center; }
.empty-state h2 { margin: 0 0 8px 0; font-size: 24px; font-weight: 600; color: var(--color-text-primary, #111827); }
.empty-state p { margin: 0 0 24px 0; font-size: 16px; color: var(--color-text-secondary, #6b7280); }
.board-list-footer { grid-column: 1 / -1; display: flex; justify-content: center; margin-top: 16px; }
</style>
