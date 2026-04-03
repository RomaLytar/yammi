import api from './client'
import axios from 'axios'
import type {
  CreateBoardRequest,
  UpdateBoardRequest,
  BoardResponse,
  ListBoardsResponse,
  GetBoardResponse,
  CreateColumnRequest,
  UpdateColumnRequest,
  ColumnResponse,
  CreateCardRequest,
  UpdateCardRequest,
  MoveCardRequest,
  DeleteCardsRequest,
  CardResponse,
  AddMemberRequest,
  MemberResponse,
  CommentResponse,
  AttachmentResponse,
  ActivityEntryResponse,
  LabelResponse,
  ChecklistResponse,
  ChecklistItemResponse,
  CardLinkResponse,
  CustomFieldDefinitionResponse,
  CustomFieldValueResponse,
  AutomationRuleResponse,
  BoardSettingsResponse,
  UserLabelResponse,
  AvailableLabelsResponse,
  BoardTemplateResponse,
} from '@/types/api'
import type {
  Board, Column, Card, Comment, Attachment, ActivityEntry,
  Label, Checklist, ChecklistItem, CardLink,
  CustomFieldDefinition, CustomFieldValue, AutomationRule,
  BoardSettings, UserLabel,
  BoardTemplate,
} from '@/types/domain'

// --- Mappers: snake_case (API) -> camelCase (Domain) ---

function mapBoard(dto: BoardResponse): Board {
  return {
    id: dto.id,
    title: dto.title,
    description: dto.description,
    ownerId: dto.owner_id,
    version: dto.version,
    createdAt: dto.created_at,
    ownerName: dto.owner_name,
    ownerAvatarUrl: dto.owner_avatar_url,
  }
}

function mapColumn(dto: ColumnResponse): Omit<Column, 'cards'> {
  return {
    id: dto.id,
    title: dto.title,
    position: dto.position,
  }
}

function mapCard(dto: CardResponse): Card {
  return {
    id: dto.id,
    title: dto.title,
    description: dto.description,
    position: dto.position,
    columnId: dto.column_id,
    assigneeId: dto.assignee_id,
    creatorId: dto.creator_id,
    releaseId: dto.release_id,
    version: dto.version,
    createdAt: dto.created_at,
    dueDate: dto.due_date,
    priority: (dto.priority as Card['priority']) || 'medium',
    taskType: (dto.task_type as Card['taskType']) || 'task',
  }
}

function mapLabel(dto: LabelResponse): Label {
  return {
    id: dto.id,
    boardId: dto.board_id,
    name: dto.name,
    color: dto.color,
    createdAt: dto.created_at,
  }
}

function mapChecklistItem(dto: ChecklistItemResponse): ChecklistItem {
  return {
    id: dto.id,
    checklistId: dto.checklist_id,
    title: dto.title,
    isChecked: dto.is_checked,
    position: dto.position,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

function mapChecklist(dto: ChecklistResponse): Checklist {
  return {
    id: dto.id,
    cardId: dto.card_id,
    boardId: dto.board_id,
    title: dto.title,
    position: dto.position,
    items: (dto.items || []).map(mapChecklistItem),
    progress: dto.progress,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

function mapCardLink(dto: CardLinkResponse): CardLink {
  return {
    id: dto.id,
    parentId: dto.parent_id,
    childId: dto.child_id,
    boardId: dto.board_id,
    linkType: dto.link_type,
    childTitle: dto.child_title,
    childColumnName: dto.child_column_name,
    createdAt: dto.created_at,
  }
}

function mapCustomFieldDefinition(dto: CustomFieldDefinitionResponse): CustomFieldDefinition {
  return {
    id: dto.id,
    boardId: dto.board_id,
    name: dto.name,
    fieldType: dto.field_type,
    options: dto.options,
    position: dto.position,
    required: dto.required,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

function mapCustomFieldValue(dto: CustomFieldValueResponse): CustomFieldValue {
  return {
    id: dto.id,
    cardId: dto.card_id,
    boardId: dto.board_id,
    fieldId: dto.field_id,
    valueText: dto.value_text,
    valueNumber: dto.value_number,
    valueDate: dto.value_date,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

function mapAutomationRule(dto: AutomationRuleResponse): AutomationRule {
  return {
    id: dto.id,
    boardId: dto.board_id,
    name: dto.name,
    enabled: dto.enabled,
    triggerType: dto.trigger_type,
    triggerConfig: dto.trigger_config,
    actionType: dto.action_type,
    actionConfig: dto.action_config,
    createdBy: dto.created_by,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

// --- API Functions ---

export async function createBoard(req: CreateBoardRequest): Promise<Board> {
  const { data } = await api.post<{ board: BoardResponse }>('/v1/boards', req)
  return mapBoard(data.board)
}

export async function getBoards(
  limit = 20,
  cursor?: string,
  ownerOnly = false,
  search = '',
  sortBy = 'updated_at',
): Promise<{ boards: Board[]; nextCursor?: string }> {
  const params = new URLSearchParams({ limit: limit.toString() })
  if (cursor) params.append('cursor', cursor)
  if (ownerOnly) params.append('owner_only', 'true')
  if (search) params.append('search', search)
  if (sortBy && sortBy !== 'updated_at') params.append('sort_by', sortBy)

  const { data } = await api.get<ListBoardsResponse>(`/v1/boards?${params}`)
  return {
    boards: data.boards.map(mapBoard),
    nextCursor: data.next_cursor,
  }
}

export async function getBoardRaw(boardId: string) {
  return api.get<GetBoardResponse>(`/v1/boards/${boardId}`)
}

export async function getBoard(boardId: string): Promise<{ board: Board; columns: Column[] }> {
  const { data } = await api.get<GetBoardResponse>(`/v1/boards/${boardId}`)

  // Группируем карточки по колонкам
  const cardsMap = new Map<string, Card[]>()

  const columns: Column[] = data.columns.map((col) => {
    const column: Column = {
      ...mapColumn(col),
      cards: cardsMap.get(col.id) || [],
    }
    return column
  })

  return {
    board: mapBoard(data.board),
    columns,
  }
}

export async function updateBoard(boardId: string, req: UpdateBoardRequest): Promise<Board> {
  const { data } = await api.put<{ board: BoardResponse }>(`/v1/boards/${boardId}`, req)
  return mapBoard(data.board)
}

export async function deleteBoards(boardIds: string[]): Promise<void> {
  await api.post('/v1/boards/delete', { board_ids: boardIds })
}

// --- Columns ---

export async function createColumn(boardId: string, req: CreateColumnRequest): Promise<Omit<Column, 'cards'>> {
  const { data } = await api.post<{ column: ColumnResponse }>(`/v1/boards/${boardId}/columns`, req)
  return mapColumn(data.column)
}

export async function getColumns(boardId: string): Promise<Array<Omit<Column, 'cards'>>> {
  const { data } = await api.get<{ columns: ColumnResponse[] }>(`/v1/boards/${boardId}/columns`)
  return data.columns.map(mapColumn)
}

export async function updateColumn(columnId: string, req: UpdateColumnRequest): Promise<Omit<Column, 'cards'>> {
  const { data } = await api.put<{ column: ColumnResponse }>(`/v1/columns/${columnId}`, req)
  return mapColumn(data.column)
}

export async function deleteColumn(columnId: string, boardId: string): Promise<void> {
  await api.delete(`/v1/columns/${columnId}`, { data: { board_id: boardId } })
}

export async function reorderColumns(boardId: string, columnIds: string[]): Promise<void> {
  await api.post(`/v1/boards/${boardId}/columns/reorder`, { column_ids: columnIds })
}

// --- Cards ---

export async function createCard(columnId: string, req: CreateCardRequest): Promise<Card> {
  const { data } = await api.post<{ card: CardResponse }>(`/v1/columns/${columnId}/cards`, req)
  return mapCard(data.card)
}

export async function getCards(columnId: string, boardId: string): Promise<Card[]> {
  const { data } = await api.get<{ cards: CardResponse[] }>(
    `/v1/columns/${columnId}/cards?board_id=${boardId}`
  )
  return data.cards.map(mapCard)
}

export async function searchBoardCards(
  boardId: string,
  filters: { search?: string; assigneeIds?: string[]; priority?: string; taskType?: string },
): Promise<Card[]> {
  const params = new URLSearchParams()
  if (filters.search) params.set('search', filters.search)
  if (filters.assigneeIds?.length) {
    for (const id of filters.assigneeIds) params.append('assignee_id', id)
  }
  if (filters.priority) params.set('priority', filters.priority)
  if (filters.taskType) params.set('task_type', filters.taskType)
  const { data } = await api.get<{ cards: CardResponse[] }>(
    `/v1/boards/${boardId}/cards/search?${params.toString()}`
  )
  return data.cards.map(mapCard)
}

export async function getCard(cardId: string): Promise<Card> {
  const { data } = await api.get<{ card: CardResponse }>(`/v1/cards/${cardId}`)
  return mapCard(data.card)
}

export async function updateCard(cardId: string, req: UpdateCardRequest): Promise<Card> {
  const { data } = await api.put<{ card: CardResponse }>(`/v1/cards/${cardId}`, req)
  return mapCard(data.card)
}

export async function deleteCards(req: DeleteCardsRequest): Promise<void> {
  await api.post('/v1/cards/delete', req)
}

export async function moveCard(cardId: string, req: MoveCardRequest): Promise<Card> {
  const { data } = await api.put<{ card: CardResponse }>(`/v1/cards/${cardId}/move`, req)
  return mapCard(data.card)
}

// --- Card Assignment ---

export async function assignCard(cardId: string, boardId: string, assigneeId: string): Promise<Card> {
  const { data } = await api.put<{ card: CardResponse }>(`/v1/cards/${cardId}/assign`, {
    board_id: boardId,
    assignee_id: assigneeId,
  })
  return mapCard(data.card)
}

export async function unassignCard(cardId: string, boardId: string): Promise<Card> {
  const { data } = await api.delete<{ card: CardResponse }>(
    `/v1/cards/${cardId}/assign?board_id=${boardId}`
  )
  return mapCard(data.card)
}

// --- Members ---

export async function addMember(boardId: string, req: AddMemberRequest): Promise<void> {
  await api.post(`/v1/boards/${boardId}/members`, req)
}

export async function removeMember(boardId: string, userId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/members/${userId}`)
}

export async function getMembers(boardId: string): Promise<MemberResponse[]> {
  const { data } = await api.get<{ members: MemberResponse[] }>(`/v1/boards/${boardId}/members`)
  return data.members
}

// --- Mappers: Comments, Attachments, Activity ---

function mapComment(dto: CommentResponse): Comment {
  return {
    id: dto.id,
    cardId: dto.card_id,
    boardId: dto.board_id,
    authorId: dto.author_id,
    parentId: dto.parent_id,
    content: dto.content,
    replyCount: dto.reply_count,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

function mapAttachment(dto: AttachmentResponse): Attachment {
  return {
    id: dto.id,
    cardId: dto.card_id,
    boardId: dto.board_id,
    fileName: dto.file_name,
    fileSize: dto.file_size,
    mimeType: dto.mime_type,
    uploaderId: dto.uploader_id,
    createdAt: dto.created_at,
  }
}

function mapActivity(dto: ActivityEntryResponse): ActivityEntry {
  return {
    id: dto.id,
    cardId: dto.card_id,
    boardId: dto.board_id,
    actorId: dto.actor_id,
    activityType: dto.activity_type,
    description: dto.description,
    changes: dto.changes,
    createdAt: dto.created_at,
  }
}

// --- Comments ---

export async function createComment(
  cardId: string,
  boardId: string,
  content: string,
  parentId?: string,
): Promise<Comment> {
  const body: { board_id: string; content: string; parent_id?: string } = {
    board_id: boardId,
    content,
  }
  if (parentId) body.parent_id = parentId
  const { data } = await api.post<{ comment: CommentResponse }>(
    `/v1/cards/${cardId}/comments`,
    body,
  )
  return mapComment(data.comment)
}

export async function listComments(
  cardId: string,
  boardId: string,
  limit?: number,
  cursor?: string,
): Promise<{ comments: Comment[]; nextCursor?: string }> {
  const params = new URLSearchParams({ board_id: boardId })
  if (limit) params.append('limit', limit.toString())
  if (cursor) params.append('cursor', cursor)
  const { data } = await api.get<{ comments: CommentResponse[]; next_cursor?: string }>(
    `/v1/cards/${cardId}/comments?${params}`,
  )
  return {
    comments: data.comments.map(mapComment),
    nextCursor: data.next_cursor,
  }
}

export async function updateComment(
  commentId: string,
  boardId: string,
  content: string,
): Promise<Comment> {
  const { data } = await api.put<{ comment: CommentResponse }>(
    `/v1/comments/${commentId}`,
    { board_id: boardId, content },
  )
  return mapComment(data.comment)
}

export async function deleteComment(commentId: string, boardId: string): Promise<void> {
  await api.delete(`/v1/comments/${commentId}?board_id=${boardId}`)
}

export async function getCommentCount(cardId: string, boardId: string): Promise<number> {
  const { data } = await api.get<{ count: number }>(
    `/v1/cards/${cardId}/comments/count?board_id=${boardId}`,
  )
  return data.count
}

// --- Attachments ---

export async function createUploadURL(
  cardId: string,
  boardId: string,
  fileName: string,
  contentType: string,
  fileSize: number,
): Promise<{ attachment: Attachment; uploadUrl: string }> {
  const { data } = await api.post<{ attachment: AttachmentResponse; upload_url: string }>(
    `/v1/cards/${cardId}/attachments/upload-url`,
    {
      board_id: boardId,
      file_name: fileName,
      content_type: contentType,
      file_size: fileSize,
    },
  )
  return {
    attachment: mapAttachment(data.attachment),
    uploadUrl: data.upload_url,
  }
}

export async function confirmUpload(attachmentId: string, boardId: string): Promise<Attachment> {
  const { data } = await api.post<{ attachment: AttachmentResponse }>(
    `/v1/attachments/${attachmentId}/confirm`,
    { board_id: boardId },
  )
  return mapAttachment(data.attachment)
}

export async function getDownloadURL(attachmentId: string, boardId: string): Promise<string> {
  const { data } = await api.get<{ download_url: string }>(
    `/v1/attachments/${attachmentId}/download-url?board_id=${boardId}`,
  )
  return data.download_url
}

export async function listAttachments(cardId: string, boardId: string): Promise<Attachment[]> {
  const { data } = await api.get<{ attachments: AttachmentResponse[] }>(
    `/v1/cards/${cardId}/attachments?board_id=${boardId}`,
  )
  return data.attachments.map(mapAttachment)
}

export async function deleteAttachment(attachmentId: string, boardId: string): Promise<void> {
  await api.delete(`/v1/attachments/${attachmentId}?board_id=${boardId}`)
}

export async function uploadFileToPresignedUrl(
  uploadUrl: string,
  file: File,
  onProgress?: (percent: number) => void,
): Promise<void> {
  await axios.put(uploadUrl, file, {
    headers: { 'Content-Type': file.type },
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded * 100) / e.total))
      }
    },
  })
}

// --- Labels ---

export async function createLabel(boardId: string, name: string, color: string): Promise<Label> {
  const { data } = await api.post<{ label: LabelResponse }>(
    `/v1/boards/${boardId}/labels`,
    { name, color },
  )
  return mapLabel(data.label)
}

export async function listLabels(boardId: string): Promise<Label[]> {
  const { data } = await api.get<{ labels: LabelResponse[] }>(
    `/v1/boards/${boardId}/labels`,
  )
  return (data.labels || []).map(mapLabel)
}

export async function updateLabel(boardId: string, labelId: string, name: string, color: string): Promise<Label> {
  const { data } = await api.put<{ label: LabelResponse }>(
    `/v1/boards/${boardId}/labels/${labelId}`,
    { name, color },
  )
  return mapLabel(data.label)
}

export async function deleteLabel(boardId: string, labelId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/labels/${labelId}`)
}

export async function addLabelToCard(boardId: string, cardId: string, labelId: string): Promise<void> {
  await api.post(`/v1/boards/${boardId}/cards/${cardId}/labels`, { label_id: labelId })
}

export async function removeLabelFromCard(boardId: string, cardId: string, labelId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/cards/${cardId}/labels/${labelId}`)
}

export async function getCardLabels(boardId: string, cardId: string): Promise<Label[]> {
  const { data } = await api.get<{ labels: LabelResponse[] }>(
    `/v1/boards/${boardId}/cards/${cardId}/labels`,
  )
  return (data.labels || []).map(mapLabel)
}

// --- Checklists ---

export async function createChecklist(boardId: string, cardId: string, title: string): Promise<Checklist> {
  const { data } = await api.post<{ checklist: ChecklistResponse }>(
    `/v1/boards/${boardId}/cards/${cardId}/checklists`,
    { title },
  )
  return mapChecklist(data.checklist)
}

export async function getChecklists(boardId: string, cardId: string): Promise<Checklist[]> {
  const { data } = await api.get<{ checklists: ChecklistResponse[] }>(
    `/v1/boards/${boardId}/cards/${cardId}/checklists`,
  )
  return (data.checklists || []).map(mapChecklist)
}

export async function updateChecklist(boardId: string, checklistId: string, title: string): Promise<Checklist> {
  const { data } = await api.put<{ checklist: ChecklistResponse }>(
    `/v1/boards/${boardId}/checklists/${checklistId}`,
    { title },
  )
  return mapChecklist(data.checklist)
}

export async function deleteChecklist(boardId: string, checklistId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/checklists/${checklistId}`)
}

export async function createChecklistItem(boardId: string, checklistId: string, title: string): Promise<ChecklistItem> {
  const { data } = await api.post<{ item: ChecklistItemResponse }>(
    `/v1/boards/${boardId}/checklists/${checklistId}/items`,
    { title },
  )
  return mapChecklistItem(data.item)
}

export async function updateChecklistItem(boardId: string, itemId: string, title: string): Promise<ChecklistItem> {
  const { data } = await api.put<{ item: ChecklistItemResponse }>(
    `/v1/boards/${boardId}/checklist-items/${itemId}`,
    { title },
  )
  return mapChecklistItem(data.item)
}

export async function deleteChecklistItem(boardId: string, itemId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/checklist-items/${itemId}`)
}

export async function toggleChecklistItem(boardId: string, itemId: string): Promise<boolean> {
  const { data } = await api.put<{ is_checked: boolean }>(
    `/v1/boards/${boardId}/checklist-items/${itemId}/toggle`,
  )
  return data.is_checked
}

// --- Card Links ---

export async function linkCards(boardId: string, parentCardId: string, childCardId: string): Promise<CardLink> {
  const { data } = await api.post<{ link: CardLinkResponse }>(
    `/v1/boards/${boardId}/cards/${parentCardId}/links`,
    { child_id: childCardId },
  )
  return mapCardLink(data.link)
}

export async function unlinkCards(boardId: string, linkId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/card-links/${linkId}`)
}

export async function getCardChildren(boardId: string, cardId: string): Promise<CardLink[]> {
  const { data } = await api.get<{ links: CardLinkResponse[] }>(
    `/v1/boards/${boardId}/cards/${cardId}/children`,
  )
  return (data.links || []).map(mapCardLink)
}

export async function getCardParents(boardId: string, cardId: string): Promise<CardLink[]> {
  const { data } = await api.get<{ links: CardLinkResponse[] }>(
    `/v1/boards/${boardId}/cards/${cardId}/parents`,
  )
  return (data.links || []).map(mapCardLink)
}

// --- Custom Fields ---

export async function createCustomField(
  boardId: string,
  data_: { name: string; field_type: string; options?: string[]; required?: boolean },
): Promise<CustomFieldDefinition> {
  const { data } = await api.post<{ field: CustomFieldDefinitionResponse }>(
    `/v1/boards/${boardId}/custom-fields`,
    data_,
  )
  return mapCustomFieldDefinition(data.field)
}

export async function listCustomFields(boardId: string): Promise<CustomFieldDefinition[]> {
  const { data } = await api.get<{ fields: CustomFieldDefinitionResponse[] }>(
    `/v1/boards/${boardId}/custom-fields`,
  )
  return (data.fields || []).map(mapCustomFieldDefinition)
}

export async function updateCustomField(
  boardId: string,
  fieldId: string,
  data_: { name?: string; options?: string[]; required?: boolean },
): Promise<CustomFieldDefinition> {
  const { data } = await api.put<{ field: CustomFieldDefinitionResponse }>(
    `/v1/boards/${boardId}/custom-fields/${fieldId}`,
    data_,
  )
  return mapCustomFieldDefinition(data.field)
}

export async function deleteCustomField(boardId: string, fieldId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/custom-fields/${fieldId}`)
}

export async function setCustomFieldValue(
  boardId: string,
  cardId: string,
  fieldId: string,
  value: { value_text?: string; value_number?: number; value_date?: string },
): Promise<void> {
  await api.put(
    `/v1/boards/${boardId}/cards/${cardId}/custom-fields/${fieldId}`,
    value,
  )
}

export async function getCardCustomFields(boardId: string, cardId: string): Promise<CustomFieldValue[]> {
  const { data } = await api.get<{ values: CustomFieldValueResponse[] }>(
    `/v1/boards/${boardId}/cards/${cardId}/custom-fields`,
  )
  return (data.values || []).map(mapCustomFieldValue)
}

// --- Automation ---

export async function createAutomationRule(
  boardId: string,
  data_: {
    name: string
    trigger_type: string
    trigger_config: Record<string, string>
    action_type: string
    action_config: Record<string, string>
  },
): Promise<AutomationRule> {
  const { data } = await api.post<{ rule: AutomationRuleResponse }>(
    `/v1/boards/${boardId}/automations`,
    data_,
  )
  return mapAutomationRule(data.rule)
}

export async function listAutomationRules(boardId: string): Promise<AutomationRule[]> {
  const { data } = await api.get<{ rules: AutomationRuleResponse[] }>(
    `/v1/boards/${boardId}/automations`,
  )
  return (data.rules || []).map(mapAutomationRule)
}

export async function updateAutomationRule(
  boardId: string,
  ruleId: string,
  data_: {
    name?: string
    enabled?: boolean
    trigger_type?: string
    trigger_config?: Record<string, string>
    action_type?: string
    action_config?: Record<string, string>
  },
): Promise<AutomationRule> {
  const { data } = await api.put<{ rule: AutomationRuleResponse }>(
    `/v1/boards/${boardId}/automations/${ruleId}`,
    data_,
  )
  return mapAutomationRule(data.rule)
}

export async function deleteAutomationRule(boardId: string, ruleId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/automations/${ruleId}`)
}

export async function getAutomationHistory(boardId: string, ruleId: string): Promise<any[]> {
  const { data } = await api.get<{ history: any[] }>(
    `/v1/boards/${boardId}/automations/${ruleId}/history`,
  )
  return data.history || []
}

// --- Board Settings ---

function mapBoardSettings(dto: BoardSettingsResponse): BoardSettings {
  return {
    boardId: dto.board_id,
    useBoardLabelsOnly: dto.use_board_labels_only,
    doneColumnId: dto.done_column_id,
    sprintDurationDays: dto.sprint_duration_days || 14,
    releasesEnabled: dto.releases_enabled || false,
  }
}

function mapUserLabel(dto: UserLabelResponse): UserLabel {
  return {
    id: dto.id,
    userId: dto.user_id,
    name: dto.name,
    color: dto.color,
    createdAt: dto.created_at,
  }
}

export async function getBoardSettings(boardId: string): Promise<BoardSettings> {
  const { data } = await api.get<{ settings: BoardSettingsResponse }>(
    `/v1/boards/${boardId}/settings`,
  )
  return mapBoardSettings(data.settings)
}

export async function updateBoardSettings(
  boardId: string,
  useBoardLabelsOnly: boolean,
  doneColumnId?: string,
  sprintDurationDays?: number,
  releasesEnabled?: boolean,
): Promise<BoardSettings> {
  const body: Record<string, unknown> = { use_board_labels_only: useBoardLabelsOnly }
  if (doneColumnId !== undefined) {
    body.done_column_id = doneColumnId || ''
  }
  if (sprintDurationDays !== undefined) {
    body.sprint_duration_days = sprintDurationDays
  }
  if (releasesEnabled !== undefined) {
    body.releases_enabled = releasesEnabled
  }
  const { data } = await api.put<{ settings: BoardSettingsResponse }>(
    `/v1/boards/${boardId}/settings`,
    body,
  )
  return mapBoardSettings(data.settings)
}

// --- User Labels (global) ---

export async function createUserLabel(name: string, color: string): Promise<UserLabel> {
  const { data } = await api.post<{ label: UserLabelResponse }>(
    '/v1/user-labels',
    { name, color },
  )
  return mapUserLabel(data.label)
}

export async function listUserLabels(): Promise<UserLabel[]> {
  const { data } = await api.get<{ labels: UserLabelResponse[] }>('/v1/user-labels')
  return (data.labels || []).map(mapUserLabel)
}

export async function updateUserLabel(labelId: string, name: string, color: string): Promise<UserLabel> {
  const { data } = await api.put<{ label: UserLabelResponse }>(
    `/v1/user-labels/${labelId}`,
    { name, color },
  )
  return mapUserLabel(data.label)
}

export async function deleteUserLabel(labelId: string): Promise<void> {
  await api.delete(`/v1/user-labels/${labelId}`)
}

// --- Available Labels (merged board + global) ---

export async function getAvailableLabels(boardId: string): Promise<{
  boardLabels: Label[]
  globalLabels: UserLabel[]
  useBoardLabelsOnly: boolean
}> {
  const { data } = await api.get<AvailableLabelsResponse>(
    `/v1/boards/${boardId}/available-labels`,
  )
  return {
    boardLabels: (data.board_labels || []).map(mapLabel),
    globalLabels: (data.user_labels || []).map(mapUserLabel),
    useBoardLabelsOnly: data.use_board_labels_only,
  }
}

// --- Activity ---

export async function getCardActivity(
  cardId: string,
  boardId: string,
  limit?: number,
  cursor?: string,
): Promise<{ entries: ActivityEntry[]; nextCursor?: string }> {
  const params = new URLSearchParams({ board_id: boardId })
  if (limit) params.append('limit', limit.toString())
  if (cursor) params.append('cursor', cursor)
  const { data } = await api.get<{ entries: ActivityEntryResponse[]; next_cursor?: string }>(
    `/v1/cards/${cardId}/activity?${params}`,
  )
  return {
    entries: data.entries.map(mapActivity),
    nextCursor: data.next_cursor,
  }
}
// --- Template Mappers ---

function mapBoardTemplate(dto: BoardTemplateResponse): BoardTemplate {
  return {
    id: dto.id,
    userId: dto.user_id,
    name: dto.name,
    description: dto.description,
    columnsData: dto.columns_data || [],
    labelsData: dto.labels_data || [],
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

// --- Board Templates ---

export async function createBoardTemplate(
  data_: {
    name: string
    description: string
    columns_data: { title: string; position: number }[]
    labels_data: { name: string; color: string }[]
  },
): Promise<BoardTemplate> {
  const { data } = await api.post<{ template: BoardTemplateResponse }>(
    '/v1/board-templates',
    data_,
  )
  return mapBoardTemplate(data.template)
}

export async function listBoardTemplates(): Promise<BoardTemplate[]> {
  const { data } = await api.get<{ templates: BoardTemplateResponse[] }>(
    '/v1/board-templates',
  )
  return (data.templates || []).map(mapBoardTemplate)
}

export async function deleteBoardTemplate(templateId: string): Promise<void> {
  await api.delete(`/v1/board-templates/${templateId}`)
}

export async function createBoardFromTemplate(
  templateId: string,
  title: string,
): Promise<Board> {
  const { data } = await api.post<{ board: BoardResponse }>(
    '/v1/boards/from-template',
    { template_id: templateId, title },
  )
  return mapBoard(data.board)
}
