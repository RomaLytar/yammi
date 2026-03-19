<template>
  <div class="auth-layout">
    <div class="auth-layout__container">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.auth-layout {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--gradient-auth-bg);
  position: relative;
  overflow: hidden;
}

.auth-layout::before {
  content: '';
  position: absolute;
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.08) 0%, transparent 70%);
  top: -200px;
  right: -100px;
  pointer-events: none;
}

.auth-layout::after {
  content: '';
  position: absolute;
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.06) 0%, transparent 70%);
  bottom: -150px;
  left: -100px;
  pointer-events: none;
}

.auth-layout__container {
  width: 100%;
  max-width: 420px;
  padding: var(--space-lg);
  position: relative;
  z-index: 1;
}
</style>
