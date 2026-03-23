<script setup lang="ts">
interface Props {
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  variant?: 'danger' | 'primary'
}

interface Emits {
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  confirmText: 'Подтвердить',
  cancelText: 'Отмена',
  variant: 'danger',
})

const emit = defineEmits<Emits>()

function handleConfirm() {
  emit('confirm')
}

function handleCancel() {
  emit('cancel')
}

function handleBackdropClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    handleCancel()
  }
}
</script>

<template>
  <div class="confirm-modal" @click="handleBackdropClick">
    <div class="confirm-modal__content">
      <div class="confirm-modal__header">
        <h3 class="confirm-modal__title">{{ title }}</h3>
      </div>

      <div class="confirm-modal__body">
        <p class="confirm-modal__message">{{ message }}</p>
      </div>

      <div class="confirm-modal__actions">
        <button
          class="confirm-modal__button confirm-modal__button--secondary"
          @click="handleCancel"
        >
          {{ cancelText }}
        </button>
        <button
          class="confirm-modal__button"
          :class="`confirm-modal__button--${variant}`"
          @click="handleConfirm"
        >
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.confirm-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.confirm-modal__content {
  background: white;
  border-radius: 12px;
  padding: 24px;
  max-width: 400px;
  width: 90%;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  animation: slideIn 0.2s ease-out;
}

@keyframes slideIn {
  from {
    transform: scale(0.95) translateY(-20px);
    opacity: 0;
  }
  to {
    transform: scale(1) translateY(0);
    opacity: 1;
  }
}

.confirm-modal__header {
  margin-bottom: 16px;
}

.confirm-modal__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-primary, #111827);
}

.confirm-modal__body {
  margin-bottom: 24px;
}

.confirm-modal__message {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: var(--color-text-secondary, #6b7280);
}

.confirm-modal__actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.confirm-modal__button {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.confirm-modal__button--secondary {
  background: var(--color-surface, #f3f4f6);
  color: var(--color-text-primary, #111827);
}

.confirm-modal__button--secondary:hover {
  background: var(--color-surface-alt, #e5e7eb);
}

.confirm-modal__button--danger {
  background: var(--color-danger, #dc2626);
  color: white;
}

.confirm-modal__button--danger:hover {
  background: var(--color-danger-dark, #b91c1c);
}

.confirm-modal__button--primary {
  background: var(--color-primary, #6b7c4e);
  color: white;
}

.confirm-modal__button--primary:hover {
  background: var(--color-primary-dark, #2563eb);
}
</style>
