<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useNotificationsStore } from '@/stores/notifications'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'
import NotificationItem from './NotificationItem.vue'
import { NOTIFICATION_TYPES } from '@/composables/useNotificationUtils'
import { useDebouncedSearch } from '@/composables/useDebouncedSearch'

const store = useNotificationsStore()

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

onMounted(() => {
  store.fetchNotifications(true)
})
</script>

<template>
  <div class="nlist">
    <div class="nlist__toolbar">
      <input v-model="searchInput" type="text" placeholder="Поиск по уведомлениям..." class="nlist__search" />
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

      <NotificationItem
        v-for="n in store.notifications"
        :key="n.id"
        :notification="n"
        :icon-size="20"
        @mark-read="handleMarkRead"
      />

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
.nlist__toolbar { display: flex; align-items: center; gap: var(--space-md); flex-wrap: wrap; margin-bottom: var(--space-lg); }
.nlist__search {
  flex: 1; min-width: 200px; padding: 10px 16px; border: 1px solid var(--color-border);
  border-radius: var(--radius-sm); font-size: var(--font-size-sm); background: var(--color-surface); box-sizing: border-box;
}
.nlist__search:focus { outline: none; border-color: var(--color-primary); box-shadow: var(--shadow-focus); }
.nlist__types { display: flex; gap: 4px; flex-wrap: wrap; }
.nlist__type-btn {
  background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius-full);
  padding: 6px 14px; font-size: var(--font-size-xs); color: var(--color-text-secondary); cursor: pointer; transition: all var(--transition-fast);
}
.nlist__type-btn:hover { background: var(--color-primary-soft); color: var(--color-primary); }
.nlist__type-btn--active { background: var(--color-primary); color: white; border-color: var(--color-primary); }
.nlist__mark-all {
  background: none; border: 1px solid var(--color-primary); color: var(--color-primary); padding: 6px 16px;
  border-radius: var(--radius-full); font-size: var(--font-size-xs); font-weight: 500; cursor: pointer;
  transition: all var(--transition-fast); white-space: nowrap;
}
.nlist__mark-all:hover { background: var(--color-primary); color: white; }
.nlist__items {
  display: flex; flex-direction: column; background: var(--color-surface);
  border: 1px solid var(--color-border); border-radius: var(--radius-md); overflow: hidden;
}
.nlist__empty { display: flex; align-items: center; justify-content: center; padding: var(--space-2xl); color: var(--color-text-tertiary); font-size: var(--font-size-sm); }
.nlist__load-more { display: flex; justify-content: center; padding: var(--space-md); }
.nlist__load-btn {
  background: none; border: 1px solid var(--color-border); color: var(--color-text-secondary); padding: 8px 24px;
  border-radius: var(--radius-full); font-size: var(--font-size-xs); cursor: pointer; transition: all var(--transition-fast);
}
.nlist__load-btn:hover { border-color: var(--color-primary); color: var(--color-primary); }
.nlist__loading { display: flex; justify-content: center; padding: var(--space-md); }
</style>
