<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

export interface SelectOption {
  value: string
  label: string
  sublabel?: string
}

interface Props {
  options: SelectOption[]
  modelValue: string
  label?: string
  placeholder?: string
  disabled?: boolean
  clearable?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: 'Выберите...',
  clearable: true,
})
const emit = defineEmits<Emits>()

const isOpen = ref(false)
const search = ref('')
const containerRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)

const selectedOption = computed(() =>
  props.options.find(o => o.value === props.modelValue)
)

const filteredOptions = computed(() => {
  if (!search.value) return props.options
  const q = search.value.toLowerCase()
  return props.options.filter(
    o => o.label.toLowerCase().includes(q) || o.sublabel?.toLowerCase().includes(q)
  )
})

function open() {
  if (props.disabled) return
  isOpen.value = true
  search.value = ''
  requestAnimationFrame(() => inputRef.value?.focus())
}

function close() {
  isOpen.value = false
  search.value = ''
}

function select(option: SelectOption) {
  emit('update:modelValue', option.value)
  close()
}

function clear(e: Event) {
  e.stopPropagation()
  emit('update:modelValue', '')
  close()
}

function handleClickOutside(e: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    close()
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') close()
}

onMounted(() => {
  document.addEventListener('mousedown', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div class="search-select" :class="{ 'search-select--disabled': disabled }" ref="containerRef">
    <label v-if="label" class="search-select__label">{{ label }}</label>

    <!-- Trigger -->
    <div class="search-select__trigger" :class="{ 'search-select__trigger--open': isOpen }" @click="open">
      <template v-if="!isOpen">
        <span v-if="selectedOption" class="search-select__value">
          {{ selectedOption.label }}
          <span v-if="selectedOption.sublabel" class="search-select__sublabel">{{ selectedOption.sublabel }}</span>
        </span>
        <span v-else class="search-select__placeholder">{{ placeholder }}</span>
      </template>
      <input
        v-else
        ref="inputRef"
        v-model="search"
        class="search-select__input"
        :placeholder="'Поиск...'"
        @click.stop
      />
      <div class="search-select__actions">
        <button
          v-if="clearable && modelValue && !isOpen"
          class="search-select__clear"
          type="button"
          @click="clear"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
        <svg class="search-select__chevron" :class="{ 'search-select__chevron--open': isOpen }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="6 9 12 15 18 9"/></svg>
      </div>
    </div>

    <!-- Dropdown -->
    <Transition name="dropdown">
      <div v-if="isOpen" class="search-select__dropdown">
        <div v-if="filteredOptions.length === 0" class="search-select__empty">
          Ничего не найдено
        </div>
        <div
          v-for="option in filteredOptions"
          :key="option.value"
          class="search-select__option"
          :class="{ 'search-select__option--selected': option.value === modelValue }"
          @click="select(option)"
        >
          <span class="search-select__option-label">{{ option.label }}</span>
          <span v-if="option.sublabel" class="search-select__option-sublabel">{{ option.sublabel }}</span>
          <svg v-if="option.value === modelValue" class="search-select__check" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.search-select {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.search-select--disabled {
  opacity: 0.5;
  pointer-events: none;
}

.search-select__label {
  font-size: var(--font-size-xs, 12px);
  font-weight: 600;
  color: var(--color-text-secondary, #6b7280);
  letter-spacing: 0.01em;
}

.search-select__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 9px 12px;
  border: 1.5px solid var(--color-input-border, #d1d5db);
  border-radius: var(--radius-md, 10px);
  background: var(--color-input-bg, #f9fafb);
  cursor: pointer;
  transition: all 0.15s;
  min-height: 42px;
}

.search-select__trigger:hover {
  border-color: var(--color-text-tertiary, #9ca3af);
}

.search-select__trigger--open {
  border-color: var(--color-input-focus, #6b7c4e);
  background: var(--color-surface, #fff);
  box-shadow: var(--shadow-focus, 0 0 0 3px rgba(99, 102, 241, 0.15));
}

.search-select__value {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--color-text, #111827);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-select__sublabel {
  font-size: 12px;
  color: var(--color-text-tertiary, #9ca3af);
}

.search-select__placeholder {
  font-size: 14px;
  color: var(--color-text-tertiary, #9ca3af);
}

.search-select__input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  color: var(--color-text, #111827);
  padding: 0;
  min-width: 0;
}

.search-select__input::placeholder {
  color: var(--color-text-tertiary, #9ca3af);
}

.search-select__actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.search-select__clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border: none;
  background: var(--color-input-bg, #f3f4f6);
  border-radius: 50%;
  color: var(--color-text-tertiary, #9ca3af);
  cursor: pointer;
  transition: all 0.15s;
}

.search-select__clear:hover {
  background: var(--color-danger-soft, #fef2f2);
  color: var(--color-danger, #ef4444);
}

.search-select__chevron {
  color: var(--color-text-tertiary, #9ca3af);
  transition: transform 0.2s;
  flex-shrink: 0;
}

.search-select__chevron--open {
  transform: rotate(180deg);
}

.search-select__dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 240px;
  overflow-y: auto;
  background: var(--color-surface, #fff);
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: var(--radius-md, 10px);
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 4px 10px -5px rgba(0, 0, 0, 0.04);
  z-index: 50;
  padding: 4px;
}

.search-select__empty {
  padding: 12px;
  text-align: center;
  font-size: 13px;
  color: var(--color-text-tertiary, #9ca3af);
}

.search-select__option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.1s;
}

.search-select__option:hover {
  background: var(--color-primary-soft, #eef2ff);
}

.search-select__option--selected {
  background: var(--color-primary-soft, #eef2ff);
}

.search-select__option-label {
  flex: 1;
  font-size: 14px;
  color: var(--color-text, #111827);
}

.search-select__option-sublabel {
  font-size: 12px;
  color: var(--color-text-tertiary, #9ca3af);
}

.search-select__check {
  color: var(--color-primary, #6b7c4e);
  flex-shrink: 0;
}

/* Dropdown scrollbar */
.search-select__dropdown::-webkit-scrollbar { width: 6px; }
.search-select__dropdown::-webkit-scrollbar-track { background: transparent; }
.search-select__dropdown::-webkit-scrollbar-thumb { background: var(--color-border, #d1d5db); border-radius: 3px; }

/* Transition */
.dropdown-enter-active, .dropdown-leave-active { transition: all 0.15s ease; }
.dropdown-enter-from, .dropdown-leave-to { opacity: 0; transform: translateY(-4px); }
</style>
