import api from './client'
import type {
  NotificationResponse,
  ListNotificationsResponse,
  NotificationSettingsResponse,
} from '@/types/api'
import type { Notification, NotificationSettings } from '@/types/domain'

function mapNotification(dto: NotificationResponse): Notification {
  return {
    id: dto.id,
    type: dto.type,
    title: dto.title,
    message: dto.message,
    metadata: dto.metadata,
    isRead: dto.is_read,
    createdAt: dto.created_at,
  }
}

export async function getNotifications(
  limit = 20,
  cursor?: string,
  type?: string,
  search?: string,
): Promise<{ notifications: Notification[]; nextCursor?: string; totalUnread: number }> {
  const params = new URLSearchParams({ limit: limit.toString() })
  if (cursor) params.append('cursor', cursor)
  if (type) params.append('type', type)
  if (search) params.append('search', search)
  const { data } = await api.get<ListNotificationsResponse>(`/v1/notifications?${params}`)
  return {
    notifications: data.notifications.map(mapNotification),
    nextCursor: data.next_cursor,
    totalUnread: data.total_unread,
  }
}

export async function markAsRead(notificationIds: string[]): Promise<void> {
  await api.post('/v1/notifications/read', { notification_ids: notificationIds })
}

export async function markAllAsRead(): Promise<void> {
  await api.post('/v1/notifications/read-all')
}

export async function getUnreadCount(): Promise<number> {
  const { data } = await api.get<{ count: number }>('/v1/notifications/unread-count')
  return data.count
}

export async function getSettings(): Promise<NotificationSettings> {
  const { data } = await api.get<{ settings: NotificationSettingsResponse }>('/v1/notifications/settings')
  return {
    enabled: data.settings.enabled,
    realtimeEnabled: data.settings.realtime_enabled,
  }
}

export async function updateSettings(enabled: boolean, realtimeEnabled: boolean): Promise<NotificationSettings> {
  const { data } = await api.put<{ settings: NotificationSettingsResponse }>('/v1/notifications/settings', {
    enabled,
    realtime_enabled: realtimeEnabled,
  })
  return {
    enabled: data.settings.enabled,
    realtimeEnabled: data.settings.realtime_enabled,
  }
}
