<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Card } from '@/types/domain'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'

interface Props {
  card: Card
}

interface Emits {
  (e: 'close'): void
  (e: 'update', data: { title: string; description: string }): void
  (e: 'delete'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const title = ref(props.card.title)
const description = ref(props.card.description)
const loading = ref(false)

watch(() => props.card, (newCard) => {
  title.value = newCard.title
  description.value = newCard.description
})

function handleUpdate() {
  if (!title.value.trim()) return

  loading.value = true
  emit('update', {
    title: title.value.trim(),
    description: description.value.trim(),
  })
}

function handleDelete() {
  if (confirm('Удалить карточку?')) {
    loading.value = true
    emit('delete')
  }
}

function handleClose() {
  if (!loading.value) {
    emit('close')
  }
}
</script>

<template>
  <BaseModal title="Редактировать карточку" @close="handleClose">
    <form @submit.prevent="handleUpdate">
      <BaseInput
        v-model="title"
        label="Название"
        :disabled="loading"
        required
        autofocus
      />

      <BaseInput
        v-model="description"
        label="Описание"
        :disabled="loading"
        type="textarea"
      />

      <div class="card-info">
        <div class="card-info__item">
          <span class="card-info__label">ID:</span>
          <span class="card-info__value">{{ card.id.slice(0, 8) }}</span>
        </div>
        <div class="card-info__item">
          <span class="card-info__label">Создана:</span>
          <span class="card-info__value">{{ new Date(card.createdAt).toLocaleString('ru-RU') }}</span>
        </div>
      </div>

      <div class="modal-actions">
        <BaseButton
          type="button"
          variant="danger"
          :disabled="loading"
          @click="handleDelete"
        >
          Удалить
        </BaseButton>
        <div class="modal-actions__right">
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
            Сохранить
          </BaseButton>
        </div>
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

.card-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: var(--color-surface-alt, #f9fafb);
  border-radius: 8px;
  font-size: 13px;
}

.card-info__item {
  display: flex;
  gap: 8px;
}

.card-info__label {
  color: var(--color-text-tertiary, #9ca3af);
  font-weight: 500;
}

.card-info__value {
  color: var(--color-text-secondary, #6b7280);
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: space-between;
  margin-top: 8px;
}

.modal-actions__right {
  display: flex;
  gap: 12px;
}
</style>
