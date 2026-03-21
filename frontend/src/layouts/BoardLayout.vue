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
  if (!userStore.profile) {
    try {
      await userStore.fetchProfile()
    } catch (error) {
      console.error('[BoardLayout] Failed to fetch profile:', error)
    }
  }

  connect()
  notificationsStore.fetchUnreadCount()
})
</script>

<template>
  <div class="board-layout">
    <AppHeader />
    <main class="board-layout__content">
      <slot />
    </main>
    <NotificationToastContainer />
  </div>
</template>

<style scoped>
.board-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.board-layout__content {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
}
</style>
