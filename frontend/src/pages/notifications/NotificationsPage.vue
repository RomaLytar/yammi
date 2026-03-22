<script setup lang="ts">
import { useNotificationsStore } from '@/stores/notifications'
import NotificationList from '@/components/notification/NotificationList.vue'
import NotificationSettingsModal from '@/components/notification/NotificationSettingsModal.vue'
import { ref } from 'vue'

const store = useNotificationsStore()
const showSettings = ref(false)
</script>

<template>
  <div class="notifications-page">
    <div class="notifications-page__header">
      <h1 class="notifications-page__title">
        Уведомления
        <span v-if="store.unreadCount > 0" class="notifications-page__badge">{{ store.unreadCount }}</span>
      </h1>
      <button class="notifications-page__settings" @click="showSettings = true">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>
          <circle cx="12" cy="12" r="3"/>
        </svg>
        Настройки
      </button>
    </div>

    <NotificationList />
    <NotificationSettingsModal v-if="showSettings" @close="showSettings = false" />
  </div>
</template>

<style scoped>
.notifications-page {
  max-width: 800px;
  margin: 0 auto;
  padding: var(--space-lg);
}
.notifications-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-lg);
}
.notifications-page__title {
  font-size: var(--font-size-xl);
  font-weight: 700;
  letter-spacing: var(--letter-spacing-tight);
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}
.notifications-page__badge {
  font-size: var(--font-size-xs);
  font-weight: 600;
  background: var(--color-primary);
  color: white;
  padding: 2px 10px;
  border-radius: var(--radius-full);
}
.notifications-page__settings {
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  padding: 8px 16px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.notifications-page__settings:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}
</style>
