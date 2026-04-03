<script setup lang="ts">
export type BoardTab = 'board' | 'releases' | 'backlog'

interface Props {
  modelValue: BoardTab
  activeReleaseName?: string
  releasesEnabled?: boolean
}

interface Emits {
  (e: 'update:modelValue', tab: BoardTab): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()
</script>

<template>
  <nav class="board-subnav">
    <button
      class="board-subnav__tab"
      :class="{ 'board-subnav__tab--active': modelValue === 'board' }"
      @click="emit('update:modelValue', 'board')"
    >
      Доска
      <span v-if="activeReleaseName && releasesEnabled" class="board-subnav__release">{{ activeReleaseName }}</span>
    </button>
    <template v-if="releasesEnabled">
      <button
        class="board-subnav__tab"
        :class="{ 'board-subnav__tab--active': modelValue === 'releases' }"
        @click="emit('update:modelValue', 'releases')"
      >
        Релизы
      </button>
      <button
        class="board-subnav__tab"
        :class="{ 'board-subnav__tab--active': modelValue === 'backlog' }"
        @click="emit('update:modelValue', 'backlog')"
      >
        Бэклог
      </button>
    </template>
  </nav>
</template>

<style scoped>
.board-subnav {
  display: flex;
  gap: 2px;
  padding: 0 24px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border-light);
}
.board-subnav__tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}
.board-subnav__tab:hover {
  color: var(--color-text-primary);
  background: var(--color-surface-alt);
}
.board-subnav__tab--active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}
.board-subnav__release {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
  letter-spacing: 0.01em;
}
</style>
