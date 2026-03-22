<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useNotificationsStore } from '@/stores/notifications'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'
import { useDebouncedSearch } from '@/composables/useDebouncedSearch'

const store = useNotificationsStore()

const NOTIFICATION_TYPES = [
  { value: '', label: 'Все' },
  { value: 'welcome', label: 'Приветствие' },
  { value: 'board', label: 'Доски' },
  { value: 'column', label: 'Колонки' },
  { value: 'card', label: 'Карточки' },
  { value: 'member', label: 'Участники' },
]

const { searchInput, debouncedValue } = useDebouncedSearch()

watch(debouncedValue, () => {
  store.search = debouncedValue.value
  store.fetchNotifications(true)
})

function onTypeChange(type: string) {
  store.typeFilter = type
  store.fetchNotifications(true)
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return 'только что'
  if (minutes < 60) return `${minutes} мин назад`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} ч назад`
  const days = Math.floor(hours / 24)
  if (days < 7) return `${days} д назад`
  return date.toLocaleDateString('ru-RU')
}

const TYPE_ICONS: Record<string, string> = {
  welcome: '<path d="M18 8h1a4 4 0 0 1 0 8h-1"/><path d="M2 8h16v9a4 4 0 0 1-4 4H6a4 4 0 0 1-4-4V8z"/><line x1="6" y1="1" x2="6" y2="4"/><line x1="10" y1="1" x2="10" y2="4"/><line x1="14" y1="1" x2="14" y2="4"/>',
  board_created: '<rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/>',
  board_updated: '<rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/>',
  board_deleted: '<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
  column_created: '<line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>',
  column_updated: '<line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>',
  column_deleted: '<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
  card_created: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>',
  card_updated: '<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>',
  card_moved: '<polyline points="5 9 2 12 5 15"/><polyline points="19 9 22 12 19 15"/><line x1="2" y1="12" x2="22" y2="12"/>',
  card_deleted: '<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
  member_added: '<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="20" y1="8" x2="20" y2="14"/><line x1="23" y1="11" x2="17" y2="11"/>',
  member_removed: '<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="23" y1="11" x2="17" y2="11"/>',
}
const DEFAULT_ICON = '<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/>'

function getTypeIcon(type: string): string {
  return TYPE_ICONS[type] || DEFAULT_ICON
}

async function handleMarkRead(id: string) {
  await store.markAsRead([id])
}

async function handleMarkAllRead() {
  await store.markAllAsRead()
}

function loadMore() {
  if (store.hasMore && !store.loading) {
    store.fetchNotifications(false)
  }
}

onMounted(() => {
  store.fetchNotifications(true)
})
</script>

<template>
  <div class="nlist">
    <div class="nlist__toolbar">
      <input
        v-model="searchInput"
        type="text"
        placeholder="Поиск по уведомлениям..."
        class="nlist__search"
      />
      <div class="nlist__types">
        <button
          v-for="t in NOTIFICATION_TYPES"
          :key="t.value"
          class="nlist__type-btn"
          :class="{ 'nlist__type-btn--active': store.typeFilter === t.value }"
          @click="onTypeChange(t.value)"
        >
          {{ t.label }}
        </button>
      </div>
      <button v-if="store.unreadCount > 0" class="nlist__mark-all" @click="handleMarkAllRead">
        Прочитать все
      </button>
    </div>

    <div class="nlist__items">
      <div v-if="store.loading && store.notifications.length === 0" class="nlist__empty">
        <BaseSpinner size="md" />
      </div>

      <div v-else-if="store.notifications.length === 0" class="nlist__empty">
        Нет уведомлений
      </div>

      <div
        v-for="n in store.notifications"
        :key="n.id"
        class="nlist__item"
        :class="{ 'nlist__item--unread': !n.isRead }"
        @click="handleMarkRead(n.id)"
      >
        <span class="nlist__icon" v-html="`<svg width='20' height='20' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'>${getTypeIcon(n.type)}</svg>`"></span>
        <div class="nlist__content">
          <div class="nlist__title">{{ n.title }}</div>
          <div v-if="n.message" class="nlist__message">{{ n.message }}</div>
          <div class="nlist__meta">
            <span v-if="n.metadata?.actor_name" class="nlist__author">{{ n.metadata.actor_name }}</span>
            <span class="nlist__time">{{ formatTime(n.createdAt) }}</span>
          </div>
        </div>
        <div v-if="!n.isRead" class="nlist__dot" />
      </div>

      <div v-if="store.hasMore && !store.loading" class="nlist__load-more">
        <button class="nlist__load-btn" @click="loadMore">Загрузить ещё</button>
      </div>

      <div v-if="store.loading && store.notifications.length > 0" class="nlist__loading">
        <BaseSpinner size="sm" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.nlist__toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  flex-wrap: wrap;
  margin-bottom: var(--space-lg);
}
.nlist__search {
  flex: 1;
  min-width: 200px;
  padding: 10px 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  background: var(--color-surface);
  box-sizing: border-box;
}
.nlist__search:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: var(--shadow-focus);
}
.nlist__types {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
.nlist__type-btn {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  padding: 6px 14px;
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.nlist__type-btn:hover {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}
.nlist__type-btn--active {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}
.nlist__mark-all {
  background: none;
  border: 1px solid var(--color-primary);
  color: var(--color-primary);
  padding: 6px 16px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}
.nlist__mark-all:hover {
  background: var(--color-primary);
  color: white;
}
.nlist__items {
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
}
.nlist__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-2xl);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}
.nlist__item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-md);
  padding: var(--space-md) var(--space-lg);
  cursor: pointer;
  transition: background var(--transition-fast);
  border-bottom: 1px solid var(--color-border-light);
}
.nlist__item:last-child { border-bottom: none; }
.nlist__item:hover { background: var(--color-surface-alt); }
.nlist__item--unread { background: var(--color-primary-soft); }
.nlist__item--unread:hover { background: rgba(99, 102, 241, 0.15); }
.nlist__icon {
  flex-shrink: 0;
  margin-top: 2px;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  background: var(--color-surface-alt);
}
.nlist__item--unread .nlist__icon {
  color: var(--color-primary);
  background: rgba(99, 102, 241, 0.15);
}
.nlist__content { flex: 1; min-width: 0; }
.nlist__title {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text);
  word-break: break-word;
}
.nlist__message {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  margin-top: 4px;
}
.nlist__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}
.nlist__author {
  font-size: var(--font-size-xs);
  color: var(--color-primary);
  font-weight: 500;
}
.nlist__time {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}
.nlist__dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-full);
  background: var(--color-primary);
  flex-shrink: 0;
  margin-top: 8px;
}
.nlist__load-more {
  display: flex;
  justify-content: center;
  padding: var(--space-md);
}
.nlist__load-btn {
  background: none;
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  padding: 8px 24px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.nlist__load-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}
.nlist__loading {
  display: flex;
  justify-content: center;
  padding: var(--space-md);
}
</style>
