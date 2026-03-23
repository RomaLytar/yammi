// Типы запросов/ответов API Gateway.
// Зеркало бэкенд DTO (snake_case). Маппинг в camelCase — в api/*.ts.

// --- Auth ---

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RefreshRequest {
  refresh_token: string
}

export interface RevokeRequest {
  refresh_token: string
}

export interface AuthResponse {
  user_id: string
  access_token: string
  refresh_token: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
}

// --- User ---

export interface UpdateProfileRequest {
  name: string
  avatar_url: string
  bio: string
}

export interface ProfileResponse {
  id: string
  email: string
  name: string
  avatar_url: string
  bio: string
  created_at: string
  updated_at: string
}

// --- Board ---

export interface CreateBoardRequest {
  title: string
  description: string
}

export interface UpdateBoardRequest {
  title: string
  description: string
  version: number
}

export interface BoardResponse {
  id: string
  title: string
  description: string
  owner_id: string
  version: number
  created_at: string
  updated_at: string
  owner_name?: string
  owner_avatar_url?: string
}

export interface ListBoardsResponse {
  boards: BoardResponse[]
  next_cursor?: string
}

export interface GetBoardResponse {
  board: BoardResponse
  columns: ColumnResponse[]
  members: MemberResponse[]
}

export interface CreateColumnRequest {
  title: string
  position: number
}

export interface UpdateColumnRequest {
  title: string
}

export interface ColumnResponse {
  id: string
  board_id: string
  title: string
  position: number
  created_at: string
  updated_at: string
  card_count: number
}

export interface CreateCardRequest {
  board_id: string
  title: string
  description: string
  position: string
  assignee_id?: string
}

export interface UpdateCardRequest {
  board_id: string
  title: string
  description: string
  assignee_id?: string
}

export interface MoveCardRequest {
  board_id: string
  from_column_id: string
  to_column_id: string
  position: string  // lexorank position (string, not index)
  version: number   // optimistic locking version
}

export interface DeleteCardsRequest {
  card_ids: string[]
  board_id: string
}

export interface CardResponse {
  id: string
  column_id: string
  title: string
  description: string
  position: string
  assignee_id?: string
  creator_id: string
  version: number
  created_at: string
  updated_at: string
}

export interface AddMemberRequest {
  user_id: string
  role: 'owner' | 'member'
}

export interface MemberResponse {
  user_id: string
  role: 'owner' | 'member'
  added_at: string
  name: string
  email: string
  avatar_url: string
}

// --- User Search ---

export interface SearchUserItem {
  id: string
  email: string
  name: string
  avatar_url: string
}

export interface SearchUsersResponse {
  users: SearchUserItem[]
}

// --- Notifications ---

export interface NotificationResponse {
  id: string
  type: string
  title: string
  message: string
  metadata: Record<string, string>
  is_read: boolean
  created_at: string
}

export interface ListNotificationsResponse {
  notifications: NotificationResponse[]
  next_cursor?: string
  total_unread: number
}

export interface NotificationSettingsResponse {
  enabled: boolean
  realtime_enabled: boolean
}

// --- Comments ---

export interface CommentResponse {
  id: string
  card_id: string
  board_id: string
  author_id: string
  parent_id?: string
  content: string
  reply_count: number
  created_at: string
  updated_at: string
}

// --- Attachments ---

export interface AttachmentResponse {
  id: string
  card_id: string
  board_id: string
  file_name: string
  file_size: number
  mime_type: string
  uploader_id: string
  created_at: string
}

// --- Activity ---

export interface ActivityEntryResponse {
  id: string
  card_id: string
  board_id: string
  actor_id: string
  activity_type: string
  description: string
  changes: Record<string, string>
  created_at: string
}

// --- Errors ---

export interface ErrorResponse {
  error: string
}
