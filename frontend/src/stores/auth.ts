import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { setTokens, registerAuthCallbacks } from '@/api/client'
import * as authApi from '@/api/auth'
import { useUserStore } from './user'
import { router } from '@/router'

const STORAGE_KEY_REFRESH = 'yammi_refresh_token'
const STORAGE_KEY_USER_ID = 'yammi_user_id'

export const useAuthStore = defineStore('auth', () => {
  const userId = ref<string | null>(null)
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const isHydrating = ref(true) // ВАЖНО: true по умолчанию, чтобы блокировать router guard

  const isAuthenticated = computed(() => !!accessToken.value)

  // Синхронизируем токены с API-клиентом
  function syncApiClient(): void {
    setTokens(accessToken.value, refreshToken.value)
  }

  function saveToStorage(): void {
    if (refreshToken.value && userId.value) {
      localStorage.setItem(STORAGE_KEY_REFRESH, refreshToken.value)
      localStorage.setItem(STORAGE_KEY_USER_ID, userId.value)
    } else {
      localStorage.removeItem(STORAGE_KEY_REFRESH)
      localStorage.removeItem(STORAGE_KEY_USER_ID)
    }
  }

  function setAuth(uid: string, access: string, refresh: string): void {
    userId.value = uid
    accessToken.value = access
    refreshToken.value = refresh
    syncApiClient()
    saveToStorage()
  }

  // Восстанавливаем сессию из localStorage (вызывается в main.ts до mount)
  async function hydrate(): Promise<void> {
    isHydrating.value = true

    // Регистрируем callback для auto-refresh из interceptor
    registerAuthCallbacks(
      (access, refresh) => {
        accessToken.value = access
        refreshToken.value = refresh
        syncApiClient()
        saveToStorage()
      },
      () => logout(),
    )

    const savedRefresh = localStorage.getItem(STORAGE_KEY_REFRESH)
    const savedUserId = localStorage.getItem(STORAGE_KEY_USER_ID)

    console.log('[AUTH] Hydrate: savedRefresh =', savedRefresh ? 'EXISTS' : 'NONE')
    console.log('[AUTH] Hydrate: savedUserId =', savedUserId)

    if (!savedRefresh || !savedUserId) {
      console.log('[AUTH] Hydrate: no saved tokens, skipping')
      isHydrating.value = false
      return
    }

    // Проактивно обновляем access token вместо ожидания 401
    try {
      console.log('[AUTH] Hydrate: calling refreshTokens...')
      const result = await authApi.refreshTokens(savedRefresh)
      console.log('[AUTH] Hydrate: refreshTokens SUCCESS')
      setAuth(savedUserId, result.accessToken, result.refreshToken)
    } catch (error) {
      // Refresh token невалиден — очищаем сессию
      console.error('[AUTH] Hydrate: refreshTokens FAILED', error)
      localStorage.removeItem(STORAGE_KEY_REFRESH)
      localStorage.removeItem(STORAGE_KEY_USER_ID)
    } finally {
      isHydrating.value = false
      console.log('[AUTH] Hydrate: DONE, isHydrating = false')
    }
  }

  async function register(email: string, password: string, name: string): Promise<void> {
    const result = await authApi.register(email, password, name)
    setAuth(result.userId, result.accessToken, result.refreshToken)
  }

  async function login(email: string, password: string): Promise<void> {
    const result = await authApi.login(email, password)
    setAuth(result.userId, result.accessToken, result.refreshToken)
  }

  function logout(): void {
    userId.value = null
    accessToken.value = null
    refreshToken.value = null
    syncApiClient()
    saveToStorage()

    // Чистим кэш профиля, чтобы при входе другим пользователем загрузился новый
    useUserStore().clear()

    router.push('/login')
  }

  return {
    userId,
    accessToken,
    refreshToken,
    isAuthenticated,
    isHydrating,
    hydrate,
    register,
    login,
    logout,
  }
})
