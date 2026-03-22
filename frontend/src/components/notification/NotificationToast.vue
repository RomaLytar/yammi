<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Notification } from '@/types/domain'
import { getTypeIcon } from '@/composables/useNotificationUtils'

const props = defineProps<{
  notification: Notification
}>()

const emit = defineEmits<{ close: [] }>()
const visible = ref(false)

onMounted(() => {
  requestAnimationFrame(() => { visible.value = true })
  setTimeout(() => {
    visible.value = false
    setTimeout(() => emit('close'), 300)
  }, 5000)
})
</script>

<template>
  <div class="toast" :class="{ 'toast--visible': visible }" @click="emit('close')">
    <span class="toast__icon" v-html="`<svg width='18' height='18' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'>${getTypeIcon(notification.type)}</svg>`"></span>
    <div class="toast__body">
      <div class="toast__title">{{ notification.title }}</div>
      <div v-if="notification.message" class="toast__message">{{ notification.message }}</div>
    </div>
    <button class="toast__close" @click.stop="emit('close')">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
      </svg>
    </button>
  </div>
</template>

<style scoped>
.toast {
  display: flex;
  align-items: flex-start;
  gap: var(--space-sm);
  padding: 12px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  max-width: 380px;
  cursor: pointer;
  transform: translateX(110%);
  opacity: 0;
  transition: all 300ms cubic-bezier(0.4, 0, 0.2, 1);
}
.toast--visible {
  transform: translateX(0);
  opacity: 1;
}
.toast__icon {
  flex-shrink: 0;
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--color-primary-soft);
}
.toast__body { flex: 1; min-width: 0; }
.toast__title {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.toast__message {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  margin-top: 2px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.toast__close {
  flex-shrink: 0;
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  cursor: pointer;
  padding: 2px;
  display: flex;
  transition: color var(--transition-fast);
}
.toast__close:hover { color: var(--color-text); }
</style>
