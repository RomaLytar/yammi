<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import AppHeader from '@/components/layout/AppHeader.vue'

const userStore = useUserStore()

onMounted(async () => {
  // Загружаем профиль если еще не загружен
  if (!userStore.profile) {
    try {
      await userStore.fetchProfile()
    } catch (error) {
      console.error('[DefaultLayout] Failed to fetch profile:', error)
    }
  }
})
</script>

<template>
  <div class="default-layout">
    <AppHeader />
    <main class="default-layout__content">
      <slot />
    </main>
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
  padding: var(--space-xl) var(--space-lg);
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
}
</style>
