<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useNotificationsStore } from '@/stores/notifications'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseButton from '@/components/shared/BaseButton.vue'

const store = useNotificationsStore()
const emit = defineEmits<{ close: [] }>()

const enabled = ref(true)
const realtimeEnabled = ref(true)
const saving = ref(false)

onMounted(async () => {
  await store.fetchSettings()
  enabled.value = store.settings.enabled
  realtimeEnabled.value = store.settings.realtimeEnabled
})

async function save() {
  saving.value = true
  try {
    await store.updateSettings(enabled.value, realtimeEnabled.value)
    emit('close')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <BaseModal title="Настройки уведомлений" @close="emit('close')">
    <div class="settings-form">
      <label class="settings-toggle">
        <input type="checkbox" v-model="enabled" />
        <span>Уведомления включены</span>
      </label>
      <p class="settings-hint">Когда выключено, новые уведомления не сохраняются. Счётчик и колокольчик не обновляются</p>

      <label class="settings-toggle">
        <input type="checkbox" v-model="realtimeEnabled" />
        <span>Всплывающие уведомления</span>
      </label>
      <p class="settings-hint">Показывать всплывающие уведомления в правом верхнем углу при новых событиях</p>
    </div>

    <template #footer>
      <BaseButton variant="secondary" @click="emit('close')">Отмена</BaseButton>
      <BaseButton @click="save" :loading="saving">Сохранить</BaseButton>
    </template>
  </BaseModal>
</template>

<style scoped>
.settings-form { display: flex; flex-direction: column; gap: var(--space-md); }
.settings-toggle {
  display: flex; align-items: center; gap: var(--space-sm);
  font-size: var(--font-size-sm); font-weight: 500; cursor: pointer;
}
.settings-toggle input { width: 18px; height: 18px; accent-color: var(--color-primary); }
.settings-hint {
  font-size: var(--font-size-xs); color: var(--color-text-tertiary);
  margin-top: -8px; margin-left: 26px;
}
</style>
