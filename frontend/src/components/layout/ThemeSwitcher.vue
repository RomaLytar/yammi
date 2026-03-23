<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useTheme } from '@/composables/useTheme'

const { currentTheme, setTheme, themes } = useTheme()
const open = ref(false)
const btnRef = ref<HTMLElement | null>(null)
const dropdownStyle = ref({ top: '0px', right: '0px' })

function toggle() {
  if (!open.value && btnRef.value) {
    const rect = btnRef.value.getBoundingClientRect()
    dropdownStyle.value = {
      top: `${rect.bottom + 8}px`,
      right: `${window.innerWidth - rect.right}px`,
    }
  }
  open.value = !open.value
}

function pick(id: typeof currentTheme.value) {
  setTheme(id)
  open.value = false
}

function closeOnOutsideClick() {
  open.value = false
}

onMounted(() => document.addEventListener('click', closeOnOutsideClick))
onUnmounted(() => document.removeEventListener('click', closeOnOutsideClick))
</script>

<template>
  <div class="theme-switcher">
    <button ref="btnRef" class="theme-btn" @click.stop="toggle" title="Сменить тему">
      <!-- Palette SVG icon -->
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="13.5" cy="6.5" r="0.5" fill="currentColor" stroke="none" />
        <circle cx="17.5" cy="10.5" r="0.5" fill="currentColor" stroke="none" />
        <circle cx="8.5" cy="7.5" r="0.5" fill="currentColor" stroke="none" />
        <circle cx="6.5" cy="12.5" r="0.5" fill="currentColor" stroke="none" />
        <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z" />
      </svg>
    </button>

    <Teleport to="body">
      <Transition name="dropdown">
        <div v-if="open" class="theme-dropdown" :style="dropdownStyle" @click.stop>
          <button
            v-for="t in themes"
            :key="t.id"
            class="theme-option"
            :class="{ 'theme-option--active': currentTheme === t.id }"
            @click="pick(t.id)"
          >
            <span class="theme-swatch" :style="{ background: t.color }" />
            <span class="theme-option__label">{{ t.label }}</span>
            <svg v-if="currentTheme === t.id" class="theme-check" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </button>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.theme-switcher {
  position: relative;
}

.theme-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-full);
  width: 36px;
  height: 36px;
  cursor: pointer;
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.7);
  padding: 0;
}

.theme-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.theme-dropdown {
  position: fixed;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  padding: 4px;
  min-width: 170px;
  z-index: 9999;
}

.theme-option {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 12px;
  border: none;
  border-radius: var(--radius-sm);
  background: none;
  font-size: 14px;
  color: var(--color-text);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.theme-option:hover {
  background: var(--color-bg-subtle);
}

.theme-option--active {
  background: var(--color-primary-soft);
  font-weight: 500;
}

.theme-swatch {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 0 0 1.5px rgba(0, 0, 0, 0.1) inset;
}

.theme-option__label {
  flex: 1;
  text-align: left;
}

.theme-check {
  color: var(--color-primary);
  flex-shrink: 0;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
