<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import type { BoardTemplate } from '@/types/domain'
import * as boardsApi from '@/api/boards'

interface Emits {
  (e: 'close'): void
  (e: 'create', data: { title: string; description: string }): void
  (e: 'createFromTemplate', data: { templateId: string; title: string }): void
}

const emit = defineEmits<Emits>()

const title = ref('')
const description = ref('')
const loading = ref(false)

// --- Board templates ---
const boardTemplates = ref<BoardTemplate[]>([])
const selectedBoardTemplateId = ref('')

onMounted(async () => {
  try {
    boardTemplates.value = await boardsApi.listBoardTemplates()
  } catch (err) {
    console.error('Failed to load board templates:', err)
  }
})

function handleCreate() {
  if (!title.value.trim()) return

  loading.value = true

  if (selectedBoardTemplateId.value) {
    emit('createFromTemplate', {
      templateId: selectedBoardTemplateId.value,
      title: title.value.trim(),
    })
  } else {
    emit('create', {
      title: title.value.trim(),
      description: description.value.trim(),
    })
  }
}

function handleClose() {
  if (!loading.value) {
    emit('close')
  }
}
</script>

<template>
  <BaseModal title="Создать доску" @close="handleClose">
    <form @submit.prevent="handleCreate">
      <!-- Template selection -->
      <div v-if="boardTemplates.length > 0" class="cbm-template-section">
        <label class="cbm-template-label">Шаблон</label>
        <select class="cbm-template-select" v-model="selectedBoardTemplateId" :disabled="loading">
          <option value="">Пустая доска</option>
          <option v-for="t in boardTemplates" :key="t.id" :value="t.id">{{ t.name }}</option>
        </select>
      </div>

      <BaseInput
        v-model="title"
        label="Название доски"
        placeholder="Моя доска"
        :disabled="loading"
        required
        autofocus
      />

      <BaseInput
        v-if="!selectedBoardTemplateId"
        v-model="description"
        label="Описание"
        placeholder="Описание доски (необязательно)"
        :disabled="loading"
        type="textarea"
      />

      <div class="modal-actions">
        <BaseButton
          type="button"
          variant="secondary"
          :disabled="loading"
          @click="handleClose"
        >
          Отмена
        </BaseButton>
        <BaseButton
          type="submit"
          :loading="loading"
          :disabled="!title.trim()"
        >
          {{ selectedBoardTemplateId ? 'Создать из шаблона' : 'Создать' }}
        </BaseButton>
      </div>
    </form>
  </BaseModal>
</template>

<style scoped>
form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 8px;
}

.cbm-template-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.cbm-template-label {
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-secondary);
  letter-spacing: 0.01em;
}

.cbm-template-select {
  width: 100%;
  padding: 10px 14px;
  border: 1.5px solid var(--color-input-border);
  border-radius: var(--radius-md);
  background: var(--color-input-bg);
  color: var(--color-text);
  font-size: 14px;
  font-family: inherit;
  outline: none;
  transition: all 0.15s;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1l4 4 4-4' stroke='%239ca3af' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
}

.cbm-template-select:focus {
  border-color: var(--color-input-focus);
  box-shadow: var(--shadow-focus);
}
</style>
