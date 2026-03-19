<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/api/client'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'

const router = useRouter()
const authStore = useAuthStore()

const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleSubmit(): Promise<void> {
  error.value = ''
  loading.value = true
  try {
    await authStore.register(email.value, password.value, name.value)
    router.push('/boards')
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Ошибка регистрации'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="register-page">
    <div class="register-page__brand">
      <h1 class="register-page__logo">Yammi</h1>
      <p class="register-page__subtitle">Создайте новый аккаунт</p>
    </div>

    <form class="register-page__form" @submit.prevent="handleSubmit">
      <BaseInput v-model="name" label="Имя" placeholder="Ваше имя" />
      <BaseInput v-model="email" label="Email" type="email" placeholder="you@example.com" />
      <BaseInput v-model="password" label="Пароль" type="password" placeholder="Минимум 8 символов" />

      <p v-if="error" class="register-page__error">{{ error }}</p>

      <BaseButton type="submit" :loading="loading" block>Создать аккаунт</BaseButton>
    </form>

    <p class="register-page__footer">
      Уже есть аккаунт? <RouterLink to="/login">Войти</RouterLink>
    </p>
  </div>
</template>

<style scoped>
.register-page {
  background: var(--color-surface);
  padding: var(--space-xl) var(--space-xl) var(--space-lg);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--color-border-light);
}

.register-page__brand {
  text-align: center;
  margin-bottom: var(--space-xl);
}

.register-page__logo {
  font-size: var(--font-size-2xl);
  font-weight: 700;
  letter-spacing: var(--letter-spacing-tight);
  background: var(--gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.register-page__subtitle {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  margin-top: var(--space-xs);
}

.register-page__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.register-page__error {
  color: var(--color-danger);
  font-size: var(--font-size-xs);
  text-align: center;
  background: var(--color-danger-soft);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
}

.register-page__footer {
  text-align: center;
  margin-top: var(--space-lg);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}
</style>
