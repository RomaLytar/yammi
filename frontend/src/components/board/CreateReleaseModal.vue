<script setup lang="ts">
import { ref } from 'vue'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'

interface Emits {
  (e: 'close'): void
  (e: 'create', data: { name: string; description: string }): void
}

const emit = defineEmits<Emits>()

const name = ref('')
const description = ref('')
const loading = ref(false)

function handleSubmit() {
  if (!name.value.trim()) return
  loading.value = true
  emit('create', { name: name.value.trim(), description: description.value.trim() })
}
</script>

<template>
  <BaseModal title="Создать релиз" @close="$emit('close')">
    <form @submit.prevent="handleSubmit" class="form">
      <BaseInput
        v-model="name"
        label="Название"
        placeholder="Sprint 1, Release v2.0..."
        :maxlength="255"
        autofocus
        required
      />
      <div class="form__field">
        <label class="form__label">Описание</label>
        <textarea
          v-model="description"
          class="form__textarea"
          placeholder="Цели релиза..."
          rows="3"
        />
      </div>
      <p class="form__hint">Дата начала установится автоматически при запуске. Дата окончания рассчитается из настройки длительности релиза.</p>
      <div class="form__actions">
        <BaseButton variant="secondary" @click="$emit('close')">Отмена</BaseButton>
        <BaseButton type="submit" :disabled="!name.trim() || loading">Создать</BaseButton>
      </div>
    </form>
  </BaseModal>
</template>

<style scoped>
.form { display: flex; flex-direction: column; gap: 16px; }
.form__field { display: flex; flex-direction: column; gap: 6px; }
.form__label { font-size: 13px; font-weight: 600; color: var(--color-text-secondary); }
.form__textarea {
  padding: 10px 12px; border: 1px solid var(--color-border); border-radius: 8px;
  font-size: 14px; color: var(--color-text-primary); background: var(--color-input-bg);
  resize: vertical; outline: none; font-family: inherit;
}
.form__textarea:focus { border-color: var(--color-primary); }
.form__hint { margin: 0; font-size: 12px; color: var(--color-text-tertiary); line-height: 1.5; }
.form__actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 4px; }
</style>
