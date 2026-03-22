<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useNotificationsStore } from '@/stores/notifications'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'
import NotificationSettingsModal from './NotificationSettingsModal.vue'
import { useDebouncedSearch } from '@/composables/useDebouncedSearch'

const props = defineProps<{
  anchorEl?: HTMLElement | null
}>()

const store = useNotificationsStore()
const emit = defineEmits<{ close: [] }>()
const showSettings = ref(false)

const panelRef = ref<HTMLElement | null>(null)
const panelStyle = ref<Record<string, string>>({})

function updatePosition() {
  if (!props.anchorEl) return
  const rect = props.anchorEl.getBoundingClientRect()
  panelStyle.value = {
    position: 'fixed',
    top: `${rect.bottom + 8}px`,
    right: `${window.innerWidth - rect.right}px`,
  }
}

function onClickOutside(e: MouseEvent) {
  if (showSettings.value) return
  const target = e.target as Node
  if (
    panelRef.value && !panelRef.value.contains(target) &&
    props.anchorEl && !props.anchorEl.contains(target)
  ) {
    emit('close')
  }
}

onMounted(() => {
  nextTick(updatePosition)
  window.addEventListener('resize', updatePosition)
  document.addEventListener('mousedown', onClickOutside)
})

onUnmounted(() => {
  window.removeEventListener('resize', updatePosition)
  document.removeEventListener('mousedown', onClickOutside)
})

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

// SVG paths — современные outline иконки (stroke-only, без fill)
const TYPE_ICONS: Record<string, string> = {
  // welcome — рука
  welcome: '<path d="M18 8h1a4 4 0 0 1 0 8h-1"/><path d="M2 8h16v9a4 4 0 0 1-4 4H6a4 4 0 0 1-4-4V8z"/><line x1="6" y1="1" x2="6" y2="4"/><line x1="10" y1="1" x2="10" y2="4"/><line x1="14" y1="1" x2="14" y2="4"/>',
  // board — доска (layout)
  board_created: '<rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/>',
  board_updated: '<rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/>',
  board_deleted: '<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
  // column — колонки
  column_created: '<line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>',
  column_updated: '<line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>',
  column_deleted: '<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
  // card — карточка (file-text)
  card_created: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>',
  card_updated: '<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>',
  card_moved: '<polyline points="5 9 2 12 5 15"/><polyline points="19 9 22 12 19 15"/><line x1="2" y1="12" x2="22" y2="12"/>',
  card_deleted: '<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
  // member — пользователь
  member_added: '<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="20" y1="8" x2="20" y2="14"/><line x1="23" y1="11" x2="17" y2="11"/>',
  member_removed: '<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="23" y1="11" x2="17" y2="11"/>',
}

// fallback — колокольчик
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
  <div ref="panelRef" class="notification-panel" :style="panelStyle">
    <div class="notification-panel__header">
      <h3 class="notification-panel__title">Уведомления</h3>
      <div class="notification-panel__header-actions">
        <button v-if="store.unreadCount > 0" class="notification-panel__mark-all" @click="handleMarkAllRead">
          Прочитать все
        </button>
        <button class="notification-panel__settings-btn" @click="showSettings = true" title="Настройки">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>
            <circle cx="12" cy="12" r="3"/>
          </svg>
        </button>
      </div>
    </div>
    <NotificationSettingsModal v-if="showSettings" @close="showSettings = false" />

    <div class="notification-panel__filters">
      <input
        v-model="searchInput"
        type="text"
        placeholder="Поиск..."
        class="notification-panel__search"
      />
      <div class="notification-panel__types">
        <button
          v-for="t in NOTIFICATION_TYPES"
          :key="t.value"
          class="notification-panel__type-btn"
          :class="{ 'notification-panel__type-btn--active': store.typeFilter === t.value }"
          @click="onTypeChange(t.value)"
        >
          {{ t.label }}
        </button>
      </div>
    </div>

    <div class="notification-panel__list" @scroll="($event: Event) => { const el = $event.target as HTMLElement; if (el.scrollHeight - el.scrollTop - el.clientHeight < 100) loadMore() }">
      <div v-if="store.loading && store.notifications.length === 0" class="notification-panel__empty">
        <BaseSpinner size="md" />
      </div>

      <div v-else-if="store.notifications.length === 0" class="notification-panel__empty">
        Нет уведомлений
      </div>

      <div
        v-for="n in store.notifications"
        :key="n.id"
        class="notification-item"
        :class="{ 'notification-item--unread': !n.isRead }"
        @click="handleMarkRead(n.id)"
      >
        <span class="notification-item__icon" v-html="`<svg width='18' height='18' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'>${getTypeIcon(n.type)}</svg>`"></span>
        <div class="notification-item__content">
          <div class="notification-item__title">{{ n.title }}</div>
          <div v-if="n.message" class="notification-item__message">{{ n.message }}</div>
          <div class="notification-item__meta">
            <span v-if="n.metadata?.actor_name" class="notification-item__author">{{ n.metadata.actor_name }}</span>
            <span class="notification-item__time">{{ formatTime(n.createdAt) }}</span>
          </div>
        </div>
        <div v-if="!n.isRead" class="notification-item__dot" />
      </div>

      <div v-if="store.loading && store.notifications.length > 0" class="notification-panel__loading">
        <BaseSpinner size="sm" />
      </div>
    </div>
    <RouterLink to="/notifications" class="notification-panel__view-all" @click="emit('close')">
      Все уведомления
    </RouterLink>
  </div>
</template>

<style scoped>
.notification-panel {
  width: 400px;
  max-height: 560px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  z-index: 10000;
  overflow: hidden;
}
.notification-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md);
  border-bottom: 1px solid var(--color-border);
}
.notification-panel__title {
  font-size: var(--font-size-md);
  font-weight: 600;
  color: var(--color-text);
}
.notification-panel__mark-all {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: var(--font-size-xs);
  font-weight: 500;
  cursor: pointer;
}
.notification-panel__mark-all:hover { color: var(--color-primary-hover); }
.notification-panel__header-actions { display: flex; align-items: center; gap: var(--space-sm); }
.notification-panel__settings-btn {
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: 4px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.notification-panel__settings-btn:hover {
  color: var(--color-text-secondary);
  background: var(--color-surface-alt);
}
.notification-panel__filters {
  padding: var(--space-sm) var(--space-md);
  border-bottom: 1px solid var(--color-border);
}
.notification-panel__search {
  width: 100%;
  padding: 6px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  background: var(--color-input-bg);
  margin-bottom: var(--space-xs);
  box-sizing: border-box;
}
.notification-panel__search:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: var(--shadow-focus);
}
.notification-panel__types {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
.notification-panel__type-btn {
  background: var(--color-surface-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  padding: 2px 10px;
  font-size: 11px;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.notification-panel__type-btn:hover {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}
.notification-panel__type-btn--active {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}
.notification-panel__list {
  flex: 1;
  overflow-y: auto;
  max-height: 400px;
}
.notification-panel__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}
.notification-panel__loading {
  display: flex;
  justify-content: center;
  padding: var(--space-md);
}
.notification-item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  cursor: pointer;
  transition: background var(--transition-fast);
  border-bottom: 1px solid var(--color-border-light);
}
.notification-item:hover { background: var(--color-surface-alt); }
.notification-item--unread { background: var(--color-primary-soft); }
.notification-item--unread:hover { background: rgba(99, 102, 241, 0.15); }
.notification-item__icon {
  flex-shrink: 0;
  margin-top: 2px;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--color-surface-alt);
}
.notification-item--unread .notification-item__icon {
  color: var(--color-primary);
  background: var(--color-primary-soft);
}
.notification-item__content { flex: 1; min-width: 0; }
.notification-item__title {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text);
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.notification-item__message {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  margin-top: 2px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.notification-item__meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 2px;
}
.notification-item__author {
  font-size: 11px;
  color: var(--color-primary);
  font-weight: 500;
}
.notification-item__time {
  font-size: 11px;
  color: var(--color-text-tertiary);
}
.notification-item__dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-full);
  background: var(--color-primary);
  flex-shrink: 0;
  margin-top: 6px;
}
.notification-panel__view-all {
  display: block;
  text-align: center;
  padding: var(--space-sm);
  font-size: var(--font-size-xs);
  font-weight: 500;
  color: var(--color-primary);
  text-decoration: none;
  border-top: 1px solid var(--color-border);
  transition: background var(--transition-fast);
}
.notification-panel__view-all:hover {
  background: var(--color-surface-alt);
}
</style>
