<script setup lang="ts">
import { useNotificationsStore } from '@/stores/notifications'
import { computed } from 'vue'

const store = useNotificationsStore()
const emit = defineEmits<{ toggle: [] }>()
const hasUnread = computed(() => store.unreadCount > 0)
const displayCount = computed(() => store.unreadCount > 99 ? '99+' : String(store.unreadCount))
</script>

<template>
  <button class="notification-bell" @click="emit('toggle')" :class="{ 'notification-bell--active': hasUnread }">
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
      <path d="M13.73 21a2 2 0 0 1-3.46 0" />
    </svg>
    <span v-if="hasUnread" class="notification-bell__badge">{{ displayCount }}</span>
  </button>
</template>

<style scoped>
.notification-bell {
  position: relative;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: rgba(255, 255, 255, 0.7);
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  cursor: pointer;
}
.notification-bell:hover {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}
.notification-bell--active {
  color: var(--color-success);
}
.notification-bell--active:hover {
  color: var(--color-success);
}
.notification-bell__badge {
  position: absolute;
  top: -4px;
  right: -4px;
  background: var(--color-danger);
  color: white;
  font-size: 10px;
  font-weight: 700;
  min-width: 18px;
  height: 18px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
  line-height: 1;
}
</style>
