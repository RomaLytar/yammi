import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Notification, NotificationSettings } from '@/types/domain'
import * as notificationsApi from '@/api/notifications'

export const useNotificationsStore = defineStore('notifications', () => {
  const notifications = ref<Notification[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const nextCursor = ref<string | undefined>(undefined)
  const hasMore = ref(true)
  const unreadCount = ref(0)
  const settings = ref<NotificationSettings>({ enabled: true, realtimeEnabled: true })
  const typeFilter = ref('')
  const search = ref('')
  const toasts = ref<Notification[]>([])

  async function fetchNotifications(reset = false): Promise<void> {
    if (loading.value) return
    try {
      loading.value = true
      error.value = null
      const cursor = reset ? undefined : nextCursor.value
      const result = await notificationsApi.getNotifications(
        20, cursor, typeFilter.value || undefined, search.value || undefined
      )
      if (reset) {
        notifications.value = result.notifications
      } else {
        notifications.value.push(...result.notifications)
      }
      nextCursor.value = result.nextCursor
      hasMore.value = !!result.nextCursor
      unreadCount.value = result.totalUnread
    } catch (err) {
      error.value = 'Ошибка загрузки уведомлений'
    } finally {
      loading.value = false
    }
  }

  async function fetchUnreadCount(): Promise<void> {
    try {
      unreadCount.value = await notificationsApi.getUnreadCount()
    } catch { /* ignore */ }
  }

  async function markAsRead(ids: string[]): Promise<void> {
    await notificationsApi.markAsRead(ids)
    for (const n of notifications.value) {
      if (ids.includes(n.id)) n.isRead = true
    }
    unreadCount.value = Math.max(0, unreadCount.value - ids.length)
  }

  async function markAllAsRead(): Promise<void> {
    await notificationsApi.markAllAsRead()
    for (const n of notifications.value) n.isRead = true
    unreadCount.value = 0
  }

  async function fetchSettings(): Promise<void> {
    try {
      settings.value = await notificationsApi.getSettings()
    } catch { /* ignore */ }
  }

  async function updateSettings(enabled: boolean, realtimeEnabled: boolean): Promise<void> {
    settings.value = await notificationsApi.updateSettings(enabled, realtimeEnabled)
  }

  // Called from WebSocket when a new notification arrives
  function addRealtimeNotification(notification: Notification): void {
    notifications.value.unshift(notification)
    unreadCount.value++

    // Показываем toast если realtime включён
    if (settings.value.realtimeEnabled) {
      toasts.value.push(notification)
      // Максимум 5 toast-ов одновременно
      if (toasts.value.length > 5) {
        toasts.value.shift()
      }
    }
  }

  function removeToast(id: string): void {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  function clear(): void {
    notifications.value = []
    nextCursor.value = undefined
    hasMore.value = true
    unreadCount.value = 0
    error.value = null
  }

  return {
    notifications, loading, error, hasMore, unreadCount, settings,
    typeFilter, search, toasts,
    fetchNotifications, fetchUnreadCount, markAsRead, markAllAsRead,
    fetchSettings, updateSettings, addRealtimeNotification, removeToast, clear,
  }
})
