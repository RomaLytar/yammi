<script setup lang="ts">
defineProps<{
  label?: string
  type?: string
  placeholder?: string
  error?: string
  disabled?: boolean
}>()

const model = defineModel<string>({ required: true })
</script>

<template>
  <div class="base-input" :class="{ 'base-input--error': error }">
    <label v-if="label" class="base-input__label">{{ label }}</label>
    <textarea
      v-if="type === 'textarea'"
      v-model="model"
      class="base-input__field base-input__textarea"
      :placeholder="placeholder"
      :disabled="disabled"
      rows="4"
    />
    <input
      v-else
      v-model="model"
      class="base-input__field"
      :type="type || 'text'"
      :placeholder="placeholder"
      :disabled="disabled"
    />
    <span v-if="error" class="base-input__error">{{ error }}</span>
  </div>
</template>

<style scoped>
.base-input { display: flex; flex-direction: column; gap: 6px; }

.base-input__label {
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-secondary);
  letter-spacing: 0.01em;
}

.base-input__field {
  padding: 10px 14px;
  border: 1.5px solid var(--color-input-border);
  border-radius: var(--radius-md);
  background: var(--color-input-bg);
  color: var(--color-text);
  outline: none;
  transition: all var(--transition-fast);
}

.base-input__field::placeholder {
  color: var(--color-text-tertiary);
}

.base-input__field:focus {
  border-color: var(--color-input-focus);
  background: var(--color-surface);
  box-shadow: var(--shadow-focus);
}

.base-input--error .base-input__field {
  border-color: var(--color-danger);
  box-shadow: 0 0 0 3px var(--color-danger-soft);
}

.base-input__error {
  font-size: var(--font-size-xs);
  color: var(--color-danger);
}

.base-input__textarea {
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
  line-height: 1.5;
}
</style>
