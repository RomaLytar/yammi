import api from './client'
import type { ProfileResponse } from '@/types/api'
import type { UserProfile } from '@/types/domain'

function mapProfile(data: ProfileResponse): UserProfile {
  return {
    id: data.id,
    email: data.email,
    name: data.name,
    avatarUrl: data.avatar_url,
    bio: data.bio,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

export async function getProfile(userId: string): Promise<UserProfile> {
  const { data } = await api.get<ProfileResponse>(`/v1/users/${userId}`)
  return mapProfile(data)
}

export async function updateProfile(
  userId: string,
  fields: { name: string; avatarUrl: string; bio: string },
): Promise<UserProfile> {
  const { data } = await api.put<ProfileResponse>(`/v1/users/${userId}`, {
    name: fields.name,
    avatar_url: fields.avatarUrl,
    bio: fields.bio,
  })
  return mapProfile(data)
}

export async function deleteUser(userId: string): Promise<void> {
  await api.delete(`/v1/users/${userId}`)
}
