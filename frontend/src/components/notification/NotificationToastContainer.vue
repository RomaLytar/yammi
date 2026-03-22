<script setup lang="ts">
import { useNotificationsStore } from '@/stores/notifications'
import { storeToRefs } from 'pinia'
import NotificationToast from './NotificationToast.vue'

const store = useNotificationsStore()
const { toasts } = storeToRefs(store)

function dismiss(id: string) {
  store.removeToast(id)
}
</script>

<template>
  <Teleport to="body">
    <div class="toast-container">
      <TransitionGroup name="toast-list">
        <NotificationToast
          v-for="n in toasts"
          :key="n.id"
          :notification="n"
          @close="dismiss(n.id)"
        />
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 72px;
  right: 16px;
  z-index: 10001;
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  pointer-events: none;
}
.toast-container > * {
  pointer-events: auto;
}

.toast-list-enter-active,
.toast-list-leave-active {
  transition: all 300ms cubic-bezier(0.4, 0, 0.2, 1);
}
.toast-list-enter-from {
  transform: translateX(110%);
  opacity: 0;
}
.toast-list-leave-to {
  transform: translateX(110%);
  opacity: 0;
}
</style>
