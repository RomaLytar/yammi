import api from './client'
import type { AuthResponse, TokenResponse } from '@/types/api'

export interface AuthResult {
  userId: string
  accessToken: string
  refreshToken: string
}

export interface TokenResult {
  accessToken: string
  refreshToken: string
}

export async function register(email: string, password: string, name: string): Promise<AuthResult> {
  const { data } = await api.post<AuthResponse>('/v1/auth/register', { email, password, name })
  return {
    userId: data.user_id,
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
  }
}

export async function login(email: string, password: string): Promise<AuthResult> {
  const { data } = await api.post<AuthResponse>('/v1/auth/login', { email, password })
  return {
    userId: data.user_id,
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
  }
}

export async function refreshTokens(refreshToken: string): Promise<TokenResult> {
  const { data } = await api.post<TokenResponse>('/v1/auth/refresh', {
    refresh_token: refreshToken,
  })
  return {
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
  }
}

export async function revokeToken(refreshToken: string): Promise<void> {
  await api.post('/v1/auth/revoke', { refresh_token: refreshToken })
}
