<script setup lang="ts">
import { ref } from 'vue'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'

interface Emits {
  (e: 'close'): void
  (e: 'create', data: { title: string; description: string }): void
}

const emit = defineEmits<Emits>()

const title = ref('')
const description = ref('')
const loading = ref(false)

function handleCreate() {
  if (!title.value.trim()) return

  loading.value = true
  emit('create', {
    title: title.value.trim(),
    description: description.value.trim(),
  })
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
      <BaseInput
        v-model="title"
        label="Название доски"
        placeholder="Моя доска"
        :disabled="loading"
        required
        autofocus
      />

      <BaseInput
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
          Создать
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
</style>
