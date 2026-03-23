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
} from '@/types/api'
import type { Board, Column, Card, Comment, Attachment, ActivityEntry } from '@/types/domain'

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
    version: dto.version,
    createdAt: dto.created_at,
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
