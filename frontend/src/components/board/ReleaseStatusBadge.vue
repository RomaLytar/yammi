<script setup lang="ts">
import type { ReleaseStatus } from '@/types/domain'

interface Props {
  status: ReleaseStatus
}

const props = defineProps<Props>()

const statusConfig: Record<ReleaseStatus, { label: string; class: string }> = {
  draft: { label: 'Черновик', class: 'badge--draft' },
  active: { label: 'Активный', class: 'badge--active' },
  completed: { label: 'Завершён', class: 'badge--completed' },
}
</script>

<template>
  <span class="badge" :class="statusConfig[status].class">
    {{ statusConfig[status].label }}
  </span>
</template>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}
.badge--draft {
  background: var(--color-surface-alt);
  color: var(--color-text-tertiary);
  border: 1px solid var(--color-border);
}
.badge--active {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}
.badge--completed {
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
  border: 1px solid rgba(107, 114, 128, 0.2);
}
</style>
