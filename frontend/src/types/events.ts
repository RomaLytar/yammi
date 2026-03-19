// Типы WebSocket-сообщений от WS Gateway.

export type WsEventType =
  | 'card.moved'
  | 'card.created'
  | 'card.updated'
  | 'card.deleted'
  | 'column.created'
  | 'column.deleted'
  | 'comment.added'
  | 'board.updated'
  | 'notification'

export interface WsMessage<T = unknown> {
  type: WsEventType
  event_id: string
  board_id?: string
  data: T
}

export interface CardMovedData {
  card_id: string
  from_column_id: string
  to_column_id: string
  position: number
}

export interface CardCreatedData {
  card_id: string
  column_id: string
  title: string
}
