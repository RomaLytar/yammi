<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useNotificationsStore } from '@/stores/notifications'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'
import NotificationItem from './NotificationItem.vue'
import NotificationSettingsModal from './NotificationSettingsModal.vue'
import { NOTIFICATION_TYPES } from '@/composables/useNotificationUtils'
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
  store.fetchNotifications(true)
})

onUnmounted(() => {
  window.removeEventListener('resize', updatePosition)
  document.removeEventListener('mousedown', onClickOutside)
})

const { searchInput, debouncedValue } = useDebouncedSearch()

watch(debouncedValue, () => {
  store.search = debouncedValue.value
  store.fetchNotifications(true)
})

function onTypeChange(type: string) {
  store.typeFilter = type
  store.fetchNotifications(true)
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
      <input v-model="searchInput" type="text" placeholder="Поиск..." class="notification-panel__search" />
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

      <NotificationItem
        v-for="n in store.notifications"
        :key="n.id"
        :notification="n"
        @mark-read="handleMarkRead"
        @navigate="emit('close')"
      />

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
.notification-panel__title { font-size: var(--font-size-md); font-weight: 600; color: var(--color-text); }
.notification-panel__mark-all { background: none; border: none; color: var(--color-primary); font-size: var(--font-size-xs); font-weight: 500; cursor: pointer; }
.notification-panel__mark-all:hover { color: var(--color-primary-hover); }
.notification-panel__header-actions { display: flex; align-items: center; gap: var(--space-sm); }
.notification-panel__settings-btn {
  background: none; border: none; color: var(--color-text-tertiary); cursor: pointer;
  display: flex; align-items: center; padding: 4px; border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.notification-panel__settings-btn:hover { color: var(--color-text-secondary); background: var(--color-surface-alt); }
.notification-panel__filters { padding: var(--space-sm) var(--space-md); border-bottom: 1px solid var(--color-border); }
.notification-panel__search {
  width: 100%; padding: 6px 12px; border: 1px solid var(--color-border); border-radius: var(--radius-sm);
  font-size: var(--font-size-xs); background: var(--color-input-bg); margin-bottom: var(--space-xs); box-sizing: border-box;
}
.notification-panel__search:focus { outline: none; border-color: var(--color-primary); box-shadow: var(--shadow-focus); }
.notification-panel__types { display: flex; gap: 4px; flex-wrap: wrap; }
.notification-panel__type-btn {
  background: var(--color-surface-alt); border: 1px solid var(--color-border); border-radius: var(--radius-full);
  padding: 2px 10px; font-size: 11px; color: var(--color-text-secondary); cursor: pointer; transition: all var(--transition-fast);
}
.notification-panel__type-btn:hover { background: var(--color-primary-soft); color: var(--color-primary); }
.notification-panel__type-btn--active { background: var(--color-primary); color: white; border-color: var(--color-primary); }
.notification-panel__list { flex: 1; overflow-y: auto; max-height: 400px; }
.notification-panel__empty { display: flex; align-items: center; justify-content: center; padding: var(--space-xl); color: var(--color-text-tertiary); font-size: var(--font-size-sm); }
.notification-panel__loading { display: flex; justify-content: center; padding: var(--space-md); }
.notification-panel__view-all {
  display: block; text-align: center; padding: var(--space-sm); font-size: var(--font-size-xs); font-weight: 500;
  color: var(--color-primary); text-decoration: none; border-top: 1px solid var(--color-border); transition: background var(--transition-fast);
}
.notification-panel__view-all:hover { background: var(--color-surface-alt); }
</style>
