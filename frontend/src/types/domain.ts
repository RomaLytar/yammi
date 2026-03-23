// Доменные сущности фронтенда (camelCase).
// Используются в stores, composables и компонентах.

export interface UserProfile {
  id: string
  email: string
  name: string
  avatarUrl: string
  bio: string
  createdAt: string
  updatedAt: string
}

export interface Board {
  id: string
  title: string
  description: string
  ownerId: string
  version: number
  createdAt: string
  ownerName?: string
  ownerAvatarUrl?: string
}

export interface Column {
  id: string
  title: string
  position: number
  cards: Card[]
}

export interface Card {
  id: string
  title: string
  description: string
  position: string // lexorank позиция (строка типа "a", "am", "b")
  columnId: string
  assigneeId?: string
  creatorId: string
  version: number  // optimistic locking version
  createdAt: string
}

export interface Comment {
  id: string
  cardId: string
  boardId: string
  authorId: string
  parentId?: string
  content: string
  replyCount: number
  createdAt: string
  updatedAt: string
}

export interface Attachment {
  id: string
  cardId: string
  boardId: string
  fileName: string
  fileSize: number
  mimeType: string
  uploaderId: string
  createdAt: string
}

export interface ActivityEntry {
  id: string
  cardId: string
  boardId: string
  actorId: string
  activityType: string
  description: string
  changes: Record<string, string>
  createdAt: string
}

export interface Notification {
  id: string
  type: string
  title: string
  message: string
  metadata: Record<string, string>
  isRead: boolean
  createdAt: string
}

export interface NotificationSettings {
  enabled: boolean
  realtimeEnabled: boolean
}
