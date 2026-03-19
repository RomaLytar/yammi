import type { WsMessage } from '@/types/events'

// Маршрутизация WebSocket-событий в stores.
// Транспорт (useWebSocket) вызывает dispatch(), сервис направляет в нужный store.

type Handler = (data: unknown) => void

const handlers = new Map<string, Handler>()
const processedIds = new Set<string>()
const MAX_PROCESSED = 1000

export function registerHandler(eventType: string, handler: Handler): void {
  handlers.set(eventType, handler)
}

export function dispatch(message: WsMessage): void {
  // Дедупликация по event_id
  if (processedIds.has(message.event_id)) return
  processedIds.add(message.event_id)
  if (processedIds.size > MAX_PROCESSED) {
    const first = processedIds.values().next().value
    if (first) processedIds.delete(first)
  }

  const handler = handlers.get(message.type)
  if (handler) {
    handler(message.data)
  }
}
