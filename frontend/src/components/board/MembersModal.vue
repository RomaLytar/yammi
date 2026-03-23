<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import * as boardsApi from '@/api/boards'
import * as usersApi from '@/api/users'
import type { MemberResponse } from '@/types/api'
import BaseModal from '@/components/shared/BaseModal.vue'

interface Props {
  boardId: string
  isOwner: boolean
}

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

const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'updated'): void }>()

const loading = ref(true)
const membersList = ref<MemberInfo[]>([])
const searchQuery = ref('')
const searchResults = ref<SearchUser[]>([])
const searching = ref(false)
const adding = ref<string | null>(null)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  await loadMembers()
})

// Debounced auto-search after 3 characters
watch(searchQuery, (val) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  searchResults.value = []

  if (val.trim().length < 3) return

  debounceTimer = setTimeout(async () => {
    searching.value = true
    try {
      const results = await usersApi.searchByEmail(val.trim())
      // Filter out already existing members
      const memberIds = new Set(membersList.value.map(m => m.userId))
      searchResults.value = results.filter(u => !memberIds.has(u.id))
    } catch {
      searchResults.value = []
    } finally {
      searching.value = false
    }
  }, 300)
})

async function loadMembers() {
  try {
    loading.value = true
    const members = await boardsApi.getMembers(props.boardId)

    membersList.value = members.map((m: MemberResponse) => ({
      userId: m.user_id,
      role: m.role,
      name: m.name || 'Неизвестный',
      email: m.email || '',
      avatarUrl: m.avatar_url || '',
    }))
  } finally {
    loading.value = false
  }
}

async function handleAddMember(user: SearchUser) {
  adding.value = user.id
  try {
    await boardsApi.addMember(props.boardId, {
      user_id: user.id,
      role: 'member',
    })
    searchQuery.value = ''
    searchResults.value = []
    await loadMembers()
    emit('updated')
  } catch (err) {
    console.error('Failed to add member:', err)
  } finally {
    adding.value = null
  }
}

async function handleRemoveMember(userId: string) {
  try {
    await boardsApi.removeMember(props.boardId, userId)
    await loadMembers()
    emit('updated')
  } catch (err) {
    console.error('Failed to remove member:', err)
  }
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map(w => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2) || '?'
}

</script>

<template>
  <BaseModal title="Участники" @close="emit('close')">
    <!-- Add member (owner only) -->
    <div v-if="isOwner" class="add-section">
      <div class="search-wrapper">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Введите email для поиска..."
          class="search-input"
        />

        <div v-if="searching || searchResults.length > 0 || (!searching && searchQuery.trim().length >= 3 && searchResults.length === 0)" class="search-dropdown">
          <div v-if="searching" class="search-hint">Поиск...</div>

          <div
            v-for="user in searchResults"
            :key="user.id"
            class="search-result-item"
            :class="{ 'search-result-item--adding': adding === user.id }"
            @click="handleAddMember(user)"
          >
            <div class="member-info">
              <div class="avatar avatar--sm">
                <img v-if="user.avatarUrl" :src="user.avatarUrl" :alt="user.name" />
                <span v-else>{{ getInitials(user.name) }}</span>
              </div>
              <div>
                <div class="member-name">{{ user.name }}</div>
                <div class="member-email">{{ user.email }}</div>
              </div>
            </div>
          </div>

          <div v-if="!searching && searchQuery.trim().length >= 3 && searchResults.length === 0" class="search-hint">
            Никого не найдено
          </div>
        </div>
      </div>
    </div>

    <!-- Members list -->
    <div v-if="loading" class="modal-loading">Загрузка...</div>

    <div v-else class="members-list">
      <div
        v-for="member in membersList"
        :key="member.userId"
        class="member-row"
      >
        <div class="member-info">
          <div class="avatar">
            <img v-if="member.avatarUrl" :src="member.avatarUrl" :alt="member.name" />
            <span v-else>{{ getInitials(member.name) }}</span>
          </div>
          <div>
            <div class="member-name">{{ member.name }}</div>
            <div class="member-email">{{ member.email }}</div>
          </div>
        </div>
        <div class="member-actions">
          <span class="role-badge" :class="{ 'role-badge--owner': member.role === 'owner' }">
            {{ member.role === 'owner' ? 'Владелец' : 'Участник' }}
          </span>
          <button
            v-if="isOwner && member.role !== 'owner'"
            class="remove-btn"
            @click="handleRemoveMember(member.userId)"
            title="Удалить участника"
          >
            &times;
          </button>
        </div>
      </div>
    </div>
  </BaseModal>
</template>

<style scoped>
.modal-loading {
  text-align: center;
  padding: 24px;
  color: #6b7280;
}

/* Add section */
.add-section {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.search-wrapper {
  position: relative;
}

.search-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.search-input:focus {
  border-color: var(--color-primary);
}

.search-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  z-index: 10;
  padding: 4px;
  max-height: 200px;
  overflow-y: auto;
}

.search-hint {
  padding: 8px 10px;
  font-size: 13px;
  color: #9ca3af;
}

.search-result-item {
  display: flex;
  align-items: center;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
}

.search-result-item:hover {
  background: var(--color-primary-light);
}

.search-result-item--adding {
  opacity: 0.5;
  pointer-events: none;
}

/* Members list */
.members-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.member-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #f9fafb;
  border-radius: 8px;
}

.member-info {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.avatar {
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

.avatar--sm {
  width: 32px;
  height: 32px;
  font-size: 12px;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.member-name {
  font-size: 14px;
  font-weight: 500;
  color: #111827;
}

.member-email {
  font-size: 12px;
  color: #9ca3af;
}

.member-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.role-badge {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 12px;
  background: #e5e7eb;
  color: #6b7280;
}

.role-badge--owner {
  background: var(--color-primary-light);
  color: var(--color-primary-hover);
}

.remove-btn {
  background: none;
  border: none;
  color: #dc2626;
  font-size: 20px;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
  opacity: 0.5;
}

.remove-btn:hover {
  opacity: 1;
}
</style>
