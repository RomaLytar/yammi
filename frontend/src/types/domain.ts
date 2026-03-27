// Доменные сущности фронтенда (camelCase).
// Используются в stores, composables и компонентах.

export type Priority = 'low' | 'medium' | 'high' | 'critical'
export type TaskType = 'bug' | 'feature' | 'task' | 'improvement'

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
  dueDate?: string       // ISO date string
  priority: Priority
  taskType: TaskType
  labels?: Label[]
  checklistStats?: { checked: number; total: number }
}

export interface Label {
  id: string
  boardId: string
  name: string
  color: string
  createdAt: string
}

export interface Checklist {
  id: string
  cardId: string
  boardId: string
  title: string
  position: number
  items: ChecklistItem[]
  progress: number
  createdAt: string
  updatedAt: string
}

export interface ChecklistItem {
  id: string
  checklistId: string
  title: string
  isChecked: boolean
  position: number
  createdAt: string
  updatedAt: string
}

export interface CardLink {
  id: string
  parentId: string
  childId: string
  boardId: string
  linkType: string
  childTitle?: string
  childColumnName?: string
  createdAt: string
}

export interface CustomFieldDefinition {
  id: string
  boardId: string
  name: string
  fieldType: 'text' | 'number' | 'date' | 'dropdown'
  options?: string[]
  position: number
  required: boolean
  createdAt: string
  updatedAt: string
}

export interface CustomFieldValue {
  id: string
  cardId: string
  boardId: string
  fieldId: string
  valueText?: string
  valueNumber?: number
  valueDate?: string
  createdAt: string
  updatedAt: string
}

export interface AutomationRule {
  id: string
  boardId: string
  name: string
  enabled: boolean
  triggerType: string
  triggerConfig: Record<string, string>
  actionType: string
  actionConfig: Record<string, string>
  createdBy: string
  createdAt: string
  updatedAt: string
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

export interface BoardSettings {
  boardId: string
  useBoardLabelsOnly: boolean
}

export interface UserLabel {
  id: string
  userId: string
  name: string
  color: string
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

export interface CardTemplate {
  id: string
  boardId: string
  userId: string
  name: string
  title: string
  description: string
  priority: Priority
  taskType: TaskType
  checklistData: { title: string; items: string[] }[]
  labelIds: string[]
  createdAt: string
  updatedAt: string
}

export interface ColumnTemplate {
  id: string
  boardId: string
  userId: string
  name: string
  columnsData: { title: string; position: number }[]
  createdAt: string
  updatedAt: string
}

export interface BoardTemplate {
  id: string
  userId: string
  name: string
  description: string
  columnsData: { title: string; position: number }[]
  labelsData: { name: string; color: string }[]
  createdAt: string
  updatedAt: string
}
