<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useNotificationsStore } from '@/stores/notifications'
import { useRealtimeConnection } from '@/composables/useRealtimeBoard'
import AppHeader from '@/components/layout/AppHeader.vue'
import NotificationToastContainer from '@/components/notification/NotificationToastContainer.vue'

const userStore = useUserStore()
const notificationsStore = useNotificationsStore()
const { connect } = useRealtimeConnection()

onMounted(async () => {
  // Загружаем профиль если еще не загружен
  if (!userStore.profile) {
    try {
      await userStore.fetchProfile()
    } catch (error) {
      console.error('[DefaultLayout] Failed to fetch profile:', error)
    }
  }

  // Инициализируем WebSocket соединение
  connect()

  // Загружаем начальный счётчик непрочитанных уведомлений
  notificationsStore.fetchUnreadCount()
})
</script>

<template>
  <div class="default-layout">
    <AppHeader />
    <main class="default-layout__content">
      <slot />
    </main>
    <NotificationToastContainer />
  </div>
</template>

<style scoped>
.default-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.default-layout__content {
  flex: 1;
}
</style>
