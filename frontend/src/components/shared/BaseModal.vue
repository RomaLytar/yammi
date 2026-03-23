<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

defineProps<{
  title?: string
  size?: 'default' | 'medium' | 'large' | 'fullscreen'
}>()
const emit = defineEmits<{ close: [] }>()

function onKeydown(e: KeyboardEvent): void {
  if (e.key === 'Escape') emit('close')
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div class="modal-overlay" @click.self="$emit('close')">
        <div
          class="modal"
          :class="[size ? `modal--${size}` : '']"
          role="dialog"
          aria-modal="true"
        >
          <div v-if="title || $slots.header" class="modal__header">
            <slot name="header">
              <h2 class="modal__title">{{ title }}</h2>
            </slot>
            <button class="modal__close" @click="$emit('close')" aria-label="Закрыть">
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <path d="M15 5L5 15M5 5l10 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </button>
          </div>
          <div class="modal__body">
            <slot />
          </div>
          <div v-if="$slots.footer" class="modal__footer">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--color-overlay);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10100;
  padding: 24px;
}

.modal {
  background: var(--color-surface);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  width: 100%;
  max-width: 560px;
  max-height: 90vh;
  overflow-y: auto;
  border: 1px solid var(--color-border-light);
  display: flex;
  flex-direction: column;
}

.modal--medium {
  max-width: 720px;
}

.modal--large {
  max-width: 1340px;
  width: 90vw;
  height: 88vh;
  max-height: 88vh;
  overflow: hidden;
}

.modal--fullscreen {
  max-width: calc(100vw - 48px);
  max-height: calc(100vh - 48px);
}

.modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg) var(--space-lg) var(--space-md);
  position: sticky;
  top: 0;
  background: var(--color-surface);
  z-index: 1;
  border-bottom: 1px solid var(--color-border-light);
}

.modal__title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  letter-spacing: var(--letter-spacing-tight);
}

.modal__close {
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  padding: var(--space-xs);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}
.modal__close:hover {
  color: var(--color-text);
  background: var(--color-primary-soft);
}

.modal__body { padding: var(--space-lg); flex: 1; overflow-y: auto; }

.modal__footer {
  padding: var(--space-md) var(--space-lg);
  border-top: 1px solid var(--color-border-light);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-sm);
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity var(--transition-normal);
}
.modal-enter-active .modal,
.modal-leave-active .modal {
  transition: transform var(--transition-normal), opacity var(--transition-normal);
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from .modal {
  transform: scale(0.95) translateY(8px);
  opacity: 0;
}
.modal-leave-to .modal {
  transform: scale(0.95) translateY(8px);
  opacity: 0;
}
</style>
