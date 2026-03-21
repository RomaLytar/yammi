import api from './client'
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
} from '@/types/api'
import type { Board, Column, Card } from '@/types/domain'

// --- Mappers: snake_case (API) -> camelCase (Domain) ---

function mapBoard(dto: BoardResponse): Board {
  return {
    id: dto.id,
    title: dto.title,
    description: dto.description,
    ownerId: dto.owner_id,
    version: dto.version,
    createdAt: dto.created_at,
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
