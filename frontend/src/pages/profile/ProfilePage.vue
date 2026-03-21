<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { ApiError } from '@/api/client'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseAvatar from '@/components/shared/BaseAvatar.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

const authStore = useAuthStore()
const userStore = useUserStore()

const name = ref('')
const avatarUrl = ref('')
const bio = ref('')
const error = ref('')
const saving = ref(false)
const saved = ref(false)

onMounted(async () => {
  if (!userStore.profile) {
    await userStore.fetchProfile()
  }
  if (userStore.profile) {
    name.value = userStore.profile.name
    avatarUrl.value = userStore.profile.avatarUrl
    bio.value = userStore.profile.bio
  }
})

async function handleSave(): Promise<void> {
  error.value = ''
  saved.value = false
  saving.value = true
  try {
    await userStore.updateProfile({
      name: name.value,
      avatarUrl: avatarUrl.value,
      bio: bio.value,
    })
    saved.value = true
    setTimeout(() => (saved.value = false), 3000)
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Ошибка сохранения'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="profile-page">
    <h1 class="profile-page__title">Профиль</h1>

    <BaseSpinner v-if="userStore.loading && !userStore.profile" />

    <form v-else-if="userStore.profile" class="profile-page__form" @submit.prevent="handleSave">
      <div class="profile-page__avatar-section">
        <BaseAvatar :name="userStore.profile.name" :src="userStore.profile.avatarUrl || undefined" size="lg" />
        <div class="profile-page__user-info">
          <span class="profile-page__name">{{ userStore.profile.name }}</span>
          <span class="profile-page__email">{{ userStore.profile.email }}</span>
        </div>
        <button class="profile-page__logout" @click="authStore.logout" title="Выйти из аккаунта">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
        </button>
      </div>

      <BaseInput v-model="name" label="Имя" />
      <BaseInput v-model="avatarUrl" label="URL аватара" placeholder="https://..." />
      <BaseInput v-model="bio" label="О себе" placeholder="Расскажите о себе" />

      <p v-if="error" class="profile-page__error">{{ error }}</p>
      <p v-if="saved" class="profile-page__saved">Сохранено</p>

      <BaseButton type="submit" :loading="saving">Сохранить</BaseButton>
    </form>
  </div>
</template>

<style scoped>
.profile-page {
  max-width: 520px;
  margin: 0 auto;
}

.profile-page__title {
  font-size: var(--font-size-xl);
  font-weight: 700;
  letter-spacing: var(--letter-spacing-tight);
  margin-bottom: var(--space-lg);
}

.profile-page__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  background: var(--color-surface);
  padding: var(--space-xl);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  border: 1px solid var(--color-border-light);
}

.profile-page__avatar-section {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid var(--color-border-light);
  margin-bottom: var(--space-sm);
}

.profile-page__user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}

.profile-page__logout {
  background: var(--color-danger-soft);
  border: 1px solid transparent;
  color: var(--color-danger);
  width: 40px;
  height: 40px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}
.profile-page__logout:hover {
  background: var(--color-danger);
  color: white;
}

.profile-page__name {
  font-weight: 600;
  font-size: var(--font-size-md);
}

.profile-page__email {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.profile-page__error {
  color: var(--color-danger);
  font-size: var(--font-size-xs);
  background: var(--color-danger-soft);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
}

.profile-page__saved {
  color: var(--color-success);
  font-size: var(--font-size-xs);
  background: var(--color-success-soft);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
}
</style>
