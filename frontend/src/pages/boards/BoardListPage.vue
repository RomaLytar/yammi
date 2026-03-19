<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

onMounted(async () => {
  if (!userStore.profile) {
    await userStore.fetchProfile()
  }
})
</script>

<template>
  <div class="board-list-page">
    <h1>Мои доски</h1>
    <div class="board-list-page__placeholder">
      <div class="board-list-page__icon">
        <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
          <rect x="6" y="10" width="14" height="28" rx="3" stroke="currentColor" stroke-width="2"/>
          <rect x="24" y="10" width="14" height="18" rx="3" stroke="currentColor" stroke-width="2"/>
        </svg>
      </div>
      <p>Board Service ещё не реализован на бэкенде.</p>
      <span>Здесь будет список досок</span>
    </div>
  </div>
</template>

<style scoped>
.board-list-page h1 {
  font-size: var(--font-size-xl);
  font-weight: 700;
  letter-spacing: var(--letter-spacing-tight);
  margin-bottom: var(--space-lg);
}

.board-list-page__placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-2xl) var(--space-xl);
  text-align: center;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 2px dashed var(--color-border);
}

.board-list-page__icon {
  color: var(--color-text-tertiary);
  margin-bottom: var(--space-sm);
}

.board-list-page__placeholder p {
  color: var(--color-text-secondary);
  font-size: var(--font-size-md);
  font-weight: 500;
}

.board-list-page__placeholder span {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}
</style>
