import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'
import { dispatch, registerHandler } from '@/services/realtimeService'
import type { WsMessage } from '@/types/events'

// Синглтон WebSocket — одно соединение на всё приложение
let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempt = 0
let stopped = false
let handlersRegistered = false

const connected = ref(false)

function getReconnectDelay(attempt: number): number {
  const base = Math.min(1000 * Math.pow(2, attempt), 30_000)
  const jitter = base * 0.2 * (Math.random() * 2 - 1)
  return base + jitter
}

function doConnect() {
  if (stopped || ws?.readyState === WebSocket.OPEN || ws?.readyState === WebSocket.CONNECTING) return

  const authStore = useAuthStore()
  if (!authStore.accessToken) return

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${window.location.host}/ws?token=${authStore.accessToken}`

  ws = new WebSocket(url)

  ws.onopen = () => {
    connected.value = true
    reconnectAttempt = 0
    console.log('[WS] Connected')
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      dispatch(data as WsMessage)
    } catch {
      console.warn('[WS] Failed to parse message:', event.data)
    }
  }

  ws.onclose = () => {
    connected.value = false
    scheduleReconnect()
  }

  ws.onerror = () => {
    ws?.close()
  }
}

function scheduleReconnect() {
  if (stopped) return
  const delay = getReconnectDelay(reconnectAttempt++)
  reconnectTimer = setTimeout(doConnect, delay)
}

function send(data: unknown) {
  if (ws?.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(data))
  }
}

function registerGlobalHandlers() {
  if (handlersRegistered) return
  handlersRegistered = true

  const notificationsStore = useNotificationsStore()

  registerHandler('notification', (data: unknown) => {
    const n = data as { id: string; type: string; title: string; message?: string; metadata?: Record<string, string> }
    notificationsStore.addRealtimeNotification({
      id: n.id,
      type: n.type,
      title: n.title,
      message: n.message || '',
      metadata: n.metadata || {},
      isRead: false,
      createdAt: new Date().toISOString(),
    })
  })

  registerHandler('unread_count', (data: unknown) => {
    const d = data as { count: number }
    notificationsStore.unreadCount = d.count
  })
}

export function useRealtimeConnection() {
  function connect() {
    stopped = false
    registerGlobalHandlers()
    doConnect()
  }

  function disconnect() {
    stopped = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
    ws = null
  }

  function subscribeBoard(boardId: string) {
    send({ type: 'subscribe', board_id: boardId })
  }

  function unsubscribeBoard(boardId: string) {
    send({ type: 'unsubscribe', board_id: boardId })
  }

  return { connected, connect, disconnect, subscribeBoard, unsubscribeBoard }
}
