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
  release_id?: string
  due_date?: string
  priority?: string
  task_type?: string
}

export interface UpdateCardRequest {
  board_id: string
  title: string
  description: string
  assignee_id?: string
  due_date?: string
  priority?: string
  task_type?: string
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
  release_id?: string
  version: number
  created_at: string
  updated_at: string
  due_date?: string
  priority?: string
  task_type?: string
}

// --- Labels ---

export interface LabelResponse {
  id: string
  board_id: string
  name: string
  color: string
  created_at: string
}

// --- Checklists ---

export interface ChecklistItemResponse {
  id: string
  checklist_id: string
  title: string
  is_checked: boolean
  position: number
  created_at: string
  updated_at: string
}

export interface ChecklistResponse {
  id: string
  card_id: string
  board_id: string
  title: string
  position: number
  items: ChecklistItemResponse[]
  progress: number
  created_at: string
  updated_at: string
}

// --- Card Links ---

export interface CardLinkResponse {
  id: string
  parent_id: string
  child_id: string
  board_id: string
  link_type: string
  child_title?: string
  child_column_name?: string
  created_at: string
}

// --- Custom Fields ---

export interface CustomFieldDefinitionResponse {
  id: string
  board_id: string
  name: string
  field_type: 'text' | 'number' | 'date' | 'dropdown'
  options?: string[]
  position: number
  required: boolean
  created_at: string
  updated_at: string
}

export interface CustomFieldValueResponse {
  id: string
  card_id: string
  board_id: string
  field_id: string
  value_text?: string
  value_number?: number
  value_date?: string
  created_at: string
  updated_at: string
}

// --- Automation ---

export interface AutomationRuleResponse {
  id: string
  board_id: string
  name: string
  enabled: boolean
  trigger_type: string
  trigger_config: Record<string, string>
  action_type: string
  action_config: Record<string, string>
  created_by: string
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

// --- Board Settings ---

export interface BoardSettingsResponse {
  board_id: string
  use_board_labels_only: boolean
  done_column_id?: string
  sprint_duration_days: number
  releases_enabled: boolean
}

// --- User Labels (global) ---

export interface UserLabelResponse {
  id: string
  user_id: string
  name: string
  color: string
  created_at: string
}

export interface AvailableLabelsResponse {
  board_labels: LabelResponse[]
  user_labels: UserLabelResponse[]
  use_board_labels_only: boolean
}

// --- Releases ---

export interface ReleaseResponse {
  id: string
  board_id: string
  name: string
  description: string
  status: string
  start_date?: string
  end_date?: string
  started_at?: string
  completed_at?: string
  created_by: string
  version: number
  created_at: string
  updated_at: string
}

export interface CreateReleaseRequest {
  name: string
  description?: string
  start_date?: string
  end_date?: string
}

export interface UpdateReleaseRequest {
  name: string
  description?: string
  start_date?: string
  end_date?: string
  version: number
}

// --- Errors ---

export interface ErrorResponse {
  error: string
}

// --- Templates ---

export interface BoardTemplateResponse {
  id: string
  user_id: string
  name: string
  description: string
  columns_data: { title: string; position: number }[]
  labels_data: { name: string; color: string }[]
  created_at: string
  updated_at: string
}
