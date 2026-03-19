<script setup lang="ts">
defineProps<{
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
  loading?: boolean
  block?: boolean
}>()
</script>

<template>
  <button
    class="base-button"
    :class="[
      `base-button--${variant || 'primary'}`,
      `base-button--${size || 'md'}`,
      { 'base-button--block': block, 'base-button--loading': loading },
    ]"
    :disabled="disabled || loading"
  >
    <span v-if="loading" class="base-button__spinner" />
    <slot />
  </button>
</template>

<style scoped>
.base-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  border: none;
  border-radius: var(--radius-md);
  font-weight: 600;
  letter-spacing: -0.01em;
  transition: all var(--transition-fast);
  white-space: nowrap;
  position: relative;
  overflow: hidden;
}

.base-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.base-button--block { width: 100%; }

/* Sizes */
.base-button--sm { padding: 6px 14px; font-size: var(--font-size-xs); border-radius: var(--radius-sm); }
.base-button--md { padding: 10px 20px; font-size: var(--font-size-sm); }
.base-button--lg { padding: 14px 28px; font-size: var(--font-size-md); }

/* Variants */
.base-button--primary {
  background: var(--gradient-primary);
  color: white;
  box-shadow: var(--shadow-primary);
}
.base-button--primary:hover:not(:disabled) {
  box-shadow: 0 6px 20px -2px rgba(99, 102, 241, 0.4);
  transform: translateY(-1px);
}
.base-button--primary:active:not(:disabled) {
  transform: translateY(0);
  box-shadow: var(--shadow-sm);
}

.base-button--secondary {
  background: var(--color-surface);
  color: var(--color-text);
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}
.base-button--secondary:hover:not(:disabled) {
  background: var(--color-input-bg);
  border-color: var(--color-text-tertiary);
}

.base-button--danger {
  background: var(--color-danger);
  color: white;
  box-shadow: 0 4px 16px -2px rgba(239, 68, 68, 0.3);
}
.base-button--danger:hover:not(:disabled) {
  background: var(--color-danger-hover);
  transform: translateY(-1px);
}

.base-button--ghost {
  background: transparent;
  color: var(--color-text-secondary);
}
.base-button--ghost:hover:not(:disabled) {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

/* Spinner */
.base-button__spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
