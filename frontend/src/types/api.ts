// Типы запросов/ответов API Gateway.
// Зеркало бэкенд DTO (snake_case). Маппинг в camelCase — в api/*.ts.

// --- Auth ---

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RefreshRequest {
  refresh_token: string
}

export interface RevokeRequest {
  refresh_token: string
}

export interface AuthResponse {
  user_id: string
  access_token: string
  refresh_token: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
}

// --- User ---

export interface UpdateProfileRequest {
  name: string
  avatar_url: string
  bio: string
}

export interface ProfileResponse {
  id: string
  email: string
  name: string
  avatar_url: string
  bio: string
  created_at: string
  updated_at: string
}

// --- Errors ---

export interface ErrorResponse {
  error: string
}
