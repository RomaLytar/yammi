import type { WsMessage } from '@/types/events'

// Маршрутизация WebSocket-событий в stores.
// Транспорт (useWebSocket) вызывает dispatch(), сервис направляет в нужный store.

type Handler = (data: unknown, message: WsMessage) => void

const handlers = new Map<string, Set<Handler>>()
const processedIds = new Set<string>()
const MAX_PROCESSED = 1000

export function registerHandler(eventType: string, handler: Handler): void {
  if (!handlers.has(eventType)) {
    handlers.set(eventType, new Set())
  }
  handlers.get(eventType)!.add(handler)
}

export function unregisterHandler(eventType: string, handler: Handler): void {
  const set = handlers.get(eventType)
  if (set) {
    set.delete(handler)
    if (set.size === 0) handlers.delete(eventType)
  }
}

export function dispatch(message: WsMessage): void {
  // Дедупликация по event_id
  if (processedIds.has(message.event_id)) return
  processedIds.add(message.event_id)
  if (processedIds.size > MAX_PROCESSED) {
    const first = processedIds.values().next().value
    if (first) processedIds.delete(first)
  }

  const set = handlers.get(message.type)
  if (set) {
    for (const handler of set) {
      handler(message.data, message)
    }
  }
}
