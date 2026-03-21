<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'
import BaseAvatar from '@/components/shared/BaseAvatar.vue'
import NotificationBell from '@/components/notification/NotificationBell.vue'
import NotificationPanel from '@/components/notification/NotificationPanel.vue'

const userStore = useUserStore()

const showNotifications = ref(false)
const notificationsRef = ref<HTMLElement | null>(null)

function toggleNotifications() {
  showNotifications.value = !showNotifications.value
}
</script>

<template>
  <header class="app-header">
    <RouterLink to="/boards" class="app-header__logo">Yammi</RouterLink>

    <nav class="app-header__nav">
      <RouterLink to="/boards" class="app-header__link">Доски</RouterLink>
    </nav>

    <div class="app-header__right">
      <div ref="notificationsRef" class="app-header__notifications">
        <NotificationBell @toggle="toggleNotifications" />
        <Teleport to="body">
          <NotificationPanel v-if="showNotifications" :anchor-el="notificationsRef" @close="showNotifications = false" />
        </Teleport>
      </div>
      <RouterLink v-if="userStore.profile" to="/profile" class="app-header__profile">
        <BaseAvatar :name="userStore.profile.name" :src="userStore.profile.avatarUrl || undefined" size="md" />
      </RouterLink>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: 0 var(--space-lg);
  height: 56px;
  background: var(--gradient-header);
  color: white;
  flex-shrink: 0;
  backdrop-filter: saturate(180%) blur(8px);
}

.app-header__logo {
  font-size: var(--font-size-lg);
  font-weight: 700;
  color: white;
  text-decoration: none;
  letter-spacing: var(--letter-spacing-tight);
}

.app-header__nav { flex: 1; display: flex; gap: var(--space-md); }

.app-header__link {
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  font-size: var(--font-size-sm);
  font-weight: 500;
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.app-header__link:hover {
  color: white;
  background: rgba(255, 255, 255, 0.1);
}

.app-header__right { display: flex; align-items: center; gap: var(--space-sm); }

.app-header__notifications { position: relative; }

.app-header__profile {
  text-decoration: none;
  display: flex;
  align-items: center;
}
</style>
