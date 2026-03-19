<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  name: string
  src?: string
  size?: 'sm' | 'md' | 'lg'
}>()

const initials = computed(() => {
  return props.name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
})

const sizeClass = computed(() => `base-avatar--${props.size || 'md'}`)
</script>

<template>
  <div class="base-avatar" :class="sizeClass">
    <img v-if="src" :src="src" :alt="name" class="base-avatar__img" />
    <span v-else class="base-avatar__initials">{{ initials }}</span>
  </div>
</template>

<style scoped>
.base-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-full);
  background: var(--gradient-primary);
  color: white;
  font-weight: 600;
  flex-shrink: 0;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.25);
}

.base-avatar--sm { width: 28px; height: 28px; font-size: 11px; }
.base-avatar--md { width: 36px; height: 36px; font-size: 13px; }
.base-avatar--lg { width: 52px; height: 52px; font-size: 18px; }

.base-avatar__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
