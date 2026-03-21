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
  authorId: string
  text: string
  createdAt: string
}
