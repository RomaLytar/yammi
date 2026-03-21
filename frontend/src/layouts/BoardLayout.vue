<script setup lang="ts">
import AppHeader from '@/components/layout/AppHeader.vue'
</script>

<template>
  <div class="board-layout">
    <AppHeader />
    <main class="board-layout__content">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.board-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.board-layout__content {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
}
</style>
