import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as usersApi from '@/api/users'
import { useAuthStore } from './auth'
import type { UserProfile } from '@/types/domain'

export const useUserStore = defineStore('user', () => {
  const profile = ref<UserProfile | null>(null)
  const loading = ref(false)

  async function fetchProfile(): Promise<void> {
    const authStore = useAuthStore()
    if (!authStore.userId) return

    loading.value = true
    try {
      profile.value = await usersApi.getProfile(authStore.userId)
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(fields: {
    name: string
    avatarUrl: string
    bio: string
  }): Promise<void> {
    const authStore = useAuthStore()
    if (!authStore.userId) return

    loading.value = true
    try {
      profile.value = await usersApi.updateProfile(authStore.userId, fields)
    } finally {
      loading.value = false
    }
  }

  function clear(): void {
    profile.value = null
  }

  return { profile, loading, fetchProfile, updateProfile, clear }
})
