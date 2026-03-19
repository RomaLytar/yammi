import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import type { ErrorResponse, TokenResponse } from '@/types/api'

// --- API Error ---

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// --- Axios instance ---

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10_000,
  headers: { 'Content-Type': 'application/json' },
})

// --- Token management (без прямого импорта store — разрываем циклическую зависимость) ---

let accessToken: string | null = null
let refreshToken: string | null = null
let onTokensUpdated: ((access: string, refresh: string) => void) | null = null
let onAuthFailed: (() => void) | null = null

export function setTokens(access: string | null, refresh: string | null): void {
  accessToken = access
  refreshToken = refresh
}

export function registerAuthCallbacks(
  onUpdate: (access: string, refresh: string) => void,
  onFail: () => void,
): void {
  onTokensUpdated = onUpdate
  onAuthFailed = onFail
}

// --- Request interceptor: добавляет Bearer token ---

const PUBLIC_PATHS = ['/v1/auth/login', '/v1/auth/register', '/v1/auth/refresh', '/v1/auth/public-key']

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const isPublic = PUBLIC_PATHS.some((p) => config.url?.includes(p))
  if (accessToken && !isPublic) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  return config
})

// --- Response interceptor: auto-refresh JWT с очередью ---

let isRefreshing = false
let failedQueue: Array<{
  resolve: (token: string) => void
  reject: (error: unknown) => void
}> = []

function processQueue(error: unknown, token: string | null): void {
  failedQueue.forEach(({ resolve, reject }) => {
    error ? reject(error) : resolve(token!)
  })
  failedQueue = []
}

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ErrorResponse>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (error.response?.status !== 401 || originalRequest._retry) {
      const message = error.response?.data?.error || error.message || 'Ошибка сети'
      const status = error.response?.status || 0
      return Promise.reject(new ApiError(status, message))
    }

    // Не пытаемся рефрешить для самого refresh-запроса
    if (originalRequest.url?.includes('/v1/auth/refresh')) {
      onAuthFailed?.()
      return Promise.reject(new ApiError(401, 'Сессия истекла'))
    }

    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      }).then((token) => {
        originalRequest.headers.Authorization = `Bearer ${token}`
        return api(originalRequest)
      })
    }

    originalRequest._retry = true
    isRefreshing = true

    try {
      if (!refreshToken) throw new Error('no refresh token')

      const { data } = await axios.post<TokenResponse>(
        `${api.defaults.baseURL}/v1/auth/refresh`,
        { refresh_token: refreshToken },
      )

      accessToken = data.access_token
      refreshToken = data.refresh_token
      onTokensUpdated?.(data.access_token, data.refresh_token)
      processQueue(null, data.access_token)

      originalRequest.headers.Authorization = `Bearer ${data.access_token}`
      return api(originalRequest)
    } catch (refreshError) {
      processQueue(refreshError, null)
      onAuthFailed?.()
      return Promise.reject(new ApiError(401, 'Сессия истекла'))
    } finally {
      isRefreshing = false
    }
  },
)

export default api
