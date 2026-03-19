<script setup lang="ts">
import AppHeader from '@/components/layout/AppHeader.vue'
</script>

<template>
  <div class="default-layout">
    <AppHeader />
    <main class="default-layout__content">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.default-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.default-layout__content {
  flex: 1;
  padding: var(--space-xl) var(--space-lg);
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
}
</style>
