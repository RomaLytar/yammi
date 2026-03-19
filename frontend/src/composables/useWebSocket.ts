import { ref, onUnmounted } from 'vue'

export interface UseWebSocketOptions {
  url: string
  onMessage: (data: unknown) => void
  onConnect?: () => void
  onDisconnect?: () => void
}

export function useWebSocket(options: UseWebSocketOptions) {
  const connected = ref(false)
  let ws: WebSocket | null = null
  let reconnectAttempt = 0
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let stopped = false

  function getReconnectDelay(attempt: number): number {
    const base = Math.min(1000 * Math.pow(2, attempt), 30_000)
    const jitter = base * 0.2 * (Math.random() * 2 - 1)
    return base + jitter
  }

  function connect(): void {
    if (stopped) return

    ws = new WebSocket(options.url)

    ws.onopen = () => {
      connected.value = true
      reconnectAttempt = 0
      options.onConnect?.()
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        options.onMessage(data)
      } catch {
        console.warn('[WS] Failed to parse message:', event.data)
      }
    }

    ws.onclose = () => {
      connected.value = false
      options.onDisconnect?.()
      scheduleReconnect()
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function scheduleReconnect(): void {
    if (stopped) return
    const delay = getReconnectDelay(reconnectAttempt++)
    reconnectTimer = setTimeout(connect, delay)
  }

  function send(data: unknown): void {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data))
    }
  }

  function disconnect(): void {
    stopped = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
    ws = null
  }

  onUnmounted(disconnect)

  return { connected, connect, disconnect, send }
}
