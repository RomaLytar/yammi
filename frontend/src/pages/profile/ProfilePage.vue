<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { ApiError } from '@/api/client'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseAvatar from '@/components/shared/BaseAvatar.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'

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
