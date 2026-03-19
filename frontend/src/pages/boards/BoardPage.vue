<template>
  <div class="board-page">
    <div class="board-page__placeholder">
      <div class="board-page__icon">
        <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
          <rect x="6" y="10" width="14" height="28" rx="3" stroke="currentColor" stroke-width="2"/>
          <rect x="24" y="10" width="14" height="18" rx="3" stroke="currentColor" stroke-width="2"/>
        </svg>
      </div>
      <p>Board Service ещё не реализован на бэкенде.</p>
      <span>Здесь будет доска с колонками и карточками</span>
    </div>
  </div>
</template>

<style scoped>
.board-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.board-page__placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-2xl) var(--space-xl);
  text-align: center;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 2px dashed var(--color-border);
}

.board-page__icon {
  color: var(--color-text-tertiary);
  margin-bottom: var(--space-sm);
}

.board-page__placeholder p {
  color: var(--color-text-secondary);
  font-size: var(--font-size-md);
  font-weight: 500;
}

.board-page__placeholder span {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}
</style>
