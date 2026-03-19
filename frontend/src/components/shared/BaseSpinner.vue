<script setup lang="ts">
defineProps<{ size?: 'sm' | 'md' | 'lg' }>()
</script>

<template>
  <div class="base-spinner" :class="`base-spinner--${size || 'md'}`" />
</template>

<style scoped>
.base-spinner {
  border-radius: 50%;
  border: 3px solid var(--color-border-light);
  border-top-color: var(--color-primary);
  animation: spin 0.7s linear infinite;
}

.base-spinner--sm { width: 18px; height: 18px; border-width: 2px; }
.base-spinner--md { width: 32px; height: 32px; }
.base-spinner--lg { width: 48px; height: 48px; border-width: 4px; }

@keyframes spin { to { transform: rotate(360deg); } }
</style>
