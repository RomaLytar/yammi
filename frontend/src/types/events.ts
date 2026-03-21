// Типы WebSocket-сообщений от WS Gateway.

export type WsEventType =
  | 'card.moved'
  | 'card.created'
  | 'card.updated'
  | 'card.deleted'
  | 'column.created'
  | 'column.updated'
  | 'column.deleted'
  | 'columns.reordered'
  | 'board.created'
  | 'board.updated'
  | 'board.deleted'
  | 'member.added'
  | 'member.removed'
  | 'notification'
  | 'unread_count'
  | 'comment.added'

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
  new_position: string
  actor_id: string
}

export interface CardCreatedData {
  card_id: string
  column_id: string
  board_id: string
  title: string
  description: string
  position: string
  actor_id: string
}

export interface CardUpdatedData {
  card_id: string
  board_id: string
  title: string
  description: string
  assignee_id?: string
  actor_id: string
}

export interface CardDeletedData {
  card_id: string
  column_id: string
  board_id: string
  actor_id: string
}

export interface ColumnCreatedData {
  column_id: string
  board_id: string
  title: string
  position: number
  actor_id: string
}

export interface ColumnDeletedData {
  column_id: string
  board_id: string
  actor_id: string
}

export interface ColumnUpdatedData {
  column_id: string
  board_id: string
  title: string
  actor_id: string
}

export interface ColumnsReorderedData {
  board_id: string
  columns: string[]
  actor_id: string
}

export interface BoardCreatedData {
  board_id: string
  owner_id: string
  title: string
  description: string
}

export interface BoardUpdatedData {
  board_id: string
  title: string
  description: string
  actor_id: string
}

export interface BoardDeletedData {
  board_id: string
  actor_id: string
}

export interface MemberAddedData {
  board_id: string
  user_id: string
  actor_id: string
  role: string
  board_title: string
}

export interface MemberRemovedData {
  board_id: string
  user_id: string
  actor_id: string
  board_title: string
}

export interface NotificationData {
  id: string
  type: string
  title: string
  message: string
  metadata: Record<string, string>
}
