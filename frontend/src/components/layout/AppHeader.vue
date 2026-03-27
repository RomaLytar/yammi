<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'
import BaseAvatar from '@/components/shared/BaseAvatar.vue'
import NotificationBell from '@/components/notification/NotificationBell.vue'
import NotificationPanel from '@/components/notification/NotificationPanel.vue'
import ThemeSwitcher from '@/components/layout/ThemeSwitcher.vue'
import GlobalLabelsModal from '@/components/board/GlobalLabelsModal.vue'

const userStore = useUserStore()

const showNotifications = ref(false)
const notificationsRef = ref<HTMLElement | null>(null)
const showGlobalLabels = ref(false)

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
      <button class="app-header__icon-btn" title="Глобальные метки" @click="showGlobalLabels = true">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
          <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" /><line x1="7" y1="7" x2="7.01" y2="7" />
        </svg>
      </button>
      <ThemeSwitcher />
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

    <GlobalLabelsModal v-if="showGlobalLabels" @close="showGlobalLabels = false" />
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

.app-header__icon-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-full);
  width: 36px;
  height: 36px;
  cursor: pointer;
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.7);
  padding: 0;
}
.app-header__icon-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.app-header__notifications { position: relative; }

.app-header__profile {
  text-decoration: none;
  display: flex;
  align-items: center;
}
</style>
