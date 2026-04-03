<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'

export interface SelectOption {
  value: string | number
  label: string
  color?: string
  sublabel?: string
}

interface Props {
  modelValue: string | number
  options: SelectOption[]
  placeholder?: string
  disabled?: boolean
  size?: 'sm' | 'md'
  label?: string
}

interface Emits {
  (e: 'update:modelValue', value: string | number): void
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: 'Выбрать...',
  disabled: false,
  size: 'md',
})

const emit = defineEmits<Emits>()

const isOpen = ref(false)
const triggerRef = ref<HTMLElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const focusedIndex = ref(-1)

const selectedOption = computed(() =>
  props.options.find(o => o.value === props.modelValue)
)

const displayLabel = computed(() =>
  selectedOption.value?.label || props.placeholder
)

function toggle() {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    focusedIndex.value = props.options.findIndex(o => o.value === props.modelValue)
    nextTick(() => scrollToFocused())
  }
}

function select(option: SelectOption) {
  emit('update:modelValue', option.value)
  isOpen.value = false
}

function handleClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (triggerRef.value?.contains(target) || dropdownRef.value?.contains(target)) return
  isOpen.value = false
}

function handleKeydown(e: KeyboardEvent) {
  if (!isOpen.value) {
    if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
      e.preventDefault()
      toggle()
    }
    return
  }

  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      focusedIndex.value = Math.min(focusedIndex.value + 1, props.options.length - 1)
      scrollToFocused()
      break
    case 'ArrowUp':
      e.preventDefault()
      focusedIndex.value = Math.max(focusedIndex.value - 1, 0)
      scrollToFocused()
      break
    case 'Enter':
      e.preventDefault()
      if (focusedIndex.value >= 0) select(props.options[focusedIndex.value])
      break
    case 'Escape':
      isOpen.value = false
      triggerRef.value?.focus()
      break
  }
}

function scrollToFocused() {
  nextTick(() => {
    const el = dropdownRef.value?.querySelector(`[data-index="${focusedIndex.value}"]`)
    el?.scrollIntoView({ block: 'nearest' })
  })
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<template>
  <div class="bs" :class="[`bs--${size}`, { 'bs--disabled': disabled, 'bs--open': isOpen }]">
    <label v-if="label" class="bs__label">{{ label }}</label>
    <button
      ref="triggerRef"
      type="button"
      class="bs__trigger"
      :disabled="disabled"
      @click="toggle"
      @keydown="handleKeydown"
    >
      <span v-if="selectedOption?.color" class="bs__dot" :style="{ background: selectedOption.color }" />
      <span class="bs__value" :class="{ 'bs__value--placeholder': !selectedOption }">
        {{ displayLabel }}
      </span>
      <svg class="bs__chevron" width="12" height="12" viewBox="0 0 12 8" fill="none">
        <path d="M1 1.5L6 6.5L11 1.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </button>

    <Transition name="bs-drop">
      <div v-if="isOpen" ref="dropdownRef" class="bs__dropdown" @keydown="handleKeydown">
        <div
          v-for="(option, i) in options"
          :key="option.value"
          :data-index="i"
          class="bs__option"
          :class="{
            'bs__option--selected': option.value === modelValue,
            'bs__option--focused': i === focusedIndex,
          }"
          @click.stop="select(option)"
          @mouseenter="focusedIndex = i"
        >
          <span v-if="option.color" class="bs__dot" :style="{ background: option.color }" />
          <span class="bs__option-label">
            {{ option.label }}
            <span v-if="option.sublabel" class="bs__option-sub">{{ option.sublabel }}</span>
          </span>
          <svg v-if="option.value === modelValue" class="bs__check" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.bs {
  position: relative;
  display: inline-flex;
  flex-direction: column;
  gap: 6px;
  min-width: 140px;
}
.bs--sm { min-width: 120px; }

.bs__label {
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-secondary);
  letter-spacing: 0.01em;
}

/* Trigger */
.bs__trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  width: 100%;
  border: 1.5px solid var(--color-input-border, var(--color-border));
  border-radius: var(--radius-sm);
  background: var(--color-input-bg, var(--color-surface-alt));
  color: var(--color-text-primary, var(--color-text));
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: inherit;
  font-size: var(--font-size-sm);
  text-align: left;
  white-space: nowrap;
  outline: none;
}

.bs--md .bs__trigger { height: 38px; }
.bs--sm .bs__trigger { height: 32px; font-size: var(--font-size-xs); padding: 0 10px; gap: 6px; }

.bs__trigger:hover:not(:disabled) {
  border-color: var(--color-text-tertiary);
}
.bs__trigger:focus-visible {
  border-color: var(--color-input-focus, var(--color-primary));
  box-shadow: var(--shadow-focus);
}
.bs--open .bs__trigger {
  border-color: var(--color-input-focus, var(--color-primary));
  box-shadow: var(--shadow-focus);
}
.bs--disabled .bs__trigger {
  opacity: 0.5;
  cursor: not-allowed;
}

.bs__value {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}
.bs__value--placeholder {
  color: var(--color-text-tertiary);
}

.bs__chevron {
  color: var(--color-text-tertiary);
  transition: transform var(--transition-fast);
  flex-shrink: 0;
}
.bs--open .bs__chevron {
  transform: rotate(180deg);
}

/* Dot (color indicator) */
.bs__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.bs--sm .bs__dot { width: 6px; height: 6px; }

/* Dropdown */
.bs__dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  z-index: 60;
  margin-top: 4px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-lg);
  padding: 4px;
  max-height: 240px;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.bs__dropdown::-webkit-scrollbar { width: 5px; }
.bs__dropdown::-webkit-scrollbar-track { background: transparent; }
.bs__dropdown::-webkit-scrollbar-thumb { background: var(--color-text-tertiary); border-radius: 3px; }

/* Option */
.bs__option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  font-size: var(--font-size-sm);
  color: var(--color-text-primary, var(--color-text));
  transition: background 80ms ease;
}
.bs--sm .bs__option { padding: 6px 8px; font-size: var(--font-size-xs); }

.bs__option--focused,
.bs__option:hover {
  background: var(--color-surface-alt, var(--color-input-bg));
}
.bs__option--selected {
  font-weight: 600;
}
.bs__option--selected.bs__option--focused,
.bs__option--selected:hover {
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
}

.bs__option-label {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}
.bs__option-sub {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
  font-weight: 400;
}

.bs__check {
  color: var(--color-primary);
  flex-shrink: 0;
}

/* Transition */
.bs-drop-enter-active { transition: opacity 120ms ease-out, transform 120ms ease-out; }
.bs-drop-leave-active { transition: opacity 80ms ease-in, transform 80ms ease-in; }
.bs-drop-enter-from { opacity: 0; transform: translateY(-4px); }
.bs-drop-leave-to { opacity: 0; transform: translateY(-4px); }
</style>
