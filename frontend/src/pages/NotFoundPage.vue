<template>
  <div class="not-found">
    <span class="not-found__code">404</span>
    <h1>Страница не найдена</h1>
    <p>Запрашиваемая страница не существует или была перемещена</p>
    <RouterLink to="/" class="not-found__link">На главную</RouterLink>
  </div>
</template>

<style scoped>
.not-found {
  text-align: center;
  background: var(--color-surface);
  padding: var(--space-2xl) var(--space-xl);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  border: 1px solid var(--color-border-light);
  max-width: 480px;
  margin: var(--space-2xl) auto;
}

.not-found__code {
  font-size: 80px;
  font-weight: 700;
  letter-spacing: var(--letter-spacing-tight);
  background: var(--gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  line-height: 1;
}

.not-found h1 {
  font-size: var(--font-size-lg);
  font-weight: 600;
  margin-top: var(--space-md);
  color: var(--color-text);
}

.not-found p {
  margin: var(--space-sm) 0 var(--space-lg);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.not-found__link {
  display: inline-flex;
  align-items: center;
  padding: 10px 24px;
  background: var(--gradient-primary);
  color: white;
  border-radius: var(--radius-md);
  font-weight: 600;
  font-size: var(--font-size-sm);
  box-shadow: var(--shadow-primary);
  transition: all var(--transition-fast);
}

.not-found__link:hover {
  box-shadow: 0 6px 20px -2px rgba(99, 102, 241, 0.4);
  transform: translateY(-1px);
  text-decoration: none;
  color: white;
}
</style>
