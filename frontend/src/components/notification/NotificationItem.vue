<script setup lang="ts">
import type { Notification } from '@/types/domain'
import { getTypeIcon, formatTime } from '@/composables/useNotificationUtils'

defineProps<{
  notification: Notification
  iconSize?: number
}>()

const emit = defineEmits<{
  markRead: [id: string]
  navigate: [boardId: string]
}>()
</script>

<template>
  <div
    class="notif-item"
    :class="{ 'notif-item--unread': !notification.isRead }"
    @click="emit('markRead', notification.id)"
  >
    <span
      class="notif-item__icon"
      v-html="`<svg width='${iconSize || 18}' height='${iconSize || 18}' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'>${getTypeIcon(notification.type)}</svg>`"
    />
    <div class="notif-item__content">
      <div class="notif-item__title">{{ notification.title }}</div>
      <div v-if="notification.message" class="notif-item__message">{{ notification.message }}</div>
      <div class="notif-item__meta">
        <span v-if="notification.metadata?.actor_name" class="notif-item__author">{{ notification.metadata.actor_name }}</span>
        <span class="notif-item__time">{{ formatTime(notification.createdAt) }}</span>
      </div>
    </div>
    <router-link
      v-if="notification.metadata?.board_id"
      :to="'/boards/' + notification.metadata.board_id"
      class="notif-item__board-link"
      title="Перейти в доску"
      @click.stop="emit('navigate', notification.metadata.board_id)"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
        <polyline points="15 3 21 3 21 9"/>
        <line x1="10" y1="14" x2="21" y2="3"/>
      </svg>
    </router-link>
    <div v-if="!notification.isRead" class="notif-item__dot" />
  </div>
</template>

<style scoped>
.notif-item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  cursor: pointer;
  transition: background var(--transition-fast);
  border-bottom: 1px solid var(--color-border-light);
}
.notif-item:last-child { border-bottom: none; }
.notif-item:hover { background: var(--color-surface-alt); }
.notif-item--unread { background: var(--color-primary-soft); }
.notif-item--unread:hover { background: rgba(99, 102, 241, 0.15); }
.notif-item__icon {
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
.notif-item--unread .notif-item__icon {
  color: var(--color-primary);
  background: rgba(99, 102, 241, 0.15);
}
.notif-item__content { flex: 1; min-width: 0; }
.notif-item__title {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text);
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.notif-item__message {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  margin-top: 2px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.notif-item__meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 2px;
}
.notif-item__author {
  font-size: 11px;
  color: var(--color-primary);
  font-weight: 500;
}
.notif-item__time {
  font-size: 11px;
  color: var(--color-text-tertiary);
}
.notif-item__board-link {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  color: var(--color-text-tertiary);
  transition: all var(--transition-fast);
  margin-top: 2px;
}
.notif-item__board-link:hover {
  color: var(--color-primary);
  background: var(--color-primary-soft);
}
.notif-item__dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-full);
  background: var(--color-primary);
  flex-shrink: 0;
  margin-top: 6px;
}
</style>
