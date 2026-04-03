import api from './client'
import type {
  ReleaseResponse,
  CreateReleaseRequest,
  UpdateReleaseRequest,
  CardResponse,
} from '@/types/api'
import type { Release, Card } from '@/types/domain'

// --- Mappers ---

function mapRelease(dto: ReleaseResponse): Release {
  return {
    id: dto.id,
    boardId: dto.board_id,
    name: dto.name,
    description: dto.description,
    status: dto.status as Release['status'],
    startDate: dto.start_date,
    endDate: dto.end_date,
    startedAt: dto.started_at,
    completedAt: dto.completed_at,
    createdBy: dto.created_by,
    version: dto.version,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
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
    priority: (dto.priority || 'medium') as Card['priority'],
    taskType: (dto.task_type || 'task') as Card['taskType'],
  }
}

// --- API functions ---

export async function createRelease(boardId: string, req: CreateReleaseRequest): Promise<Release> {
  const { data } = await api.post<{ release: ReleaseResponse }>(
    `/v1/boards/${boardId}/releases`,
    req,
  )
  return mapRelease(data.release)
}

export async function getRelease(boardId: string, releaseId: string): Promise<Release> {
  const { data } = await api.get<{ release: ReleaseResponse }>(
    `/v1/boards/${boardId}/releases/${releaseId}`,
  )
  return mapRelease(data.release)
}

export async function listReleases(boardId: string): Promise<Release[]> {
  const { data } = await api.get<{ releases: ReleaseResponse[] }>(
    `/v1/boards/${boardId}/releases`,
  )
  return (data.releases || []).map(mapRelease)
}

export async function updateRelease(
  boardId: string,
  releaseId: string,
  req: UpdateReleaseRequest,
): Promise<Release> {
  const { data } = await api.put<{ release: ReleaseResponse }>(
    `/v1/boards/${boardId}/releases/${releaseId}`,
    req,
  )
  return mapRelease(data.release)
}

export async function deleteRelease(boardId: string, releaseId: string): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/releases/${releaseId}`)
}

export async function startRelease(boardId: string, releaseId: string): Promise<Release> {
  const { data } = await api.post<{ release: ReleaseResponse }>(
    `/v1/boards/${boardId}/releases/${releaseId}/start`,
  )
  return mapRelease(data.release)
}

export async function completeRelease(boardId: string, releaseId: string): Promise<Release> {
  const { data } = await api.post<{ release: ReleaseResponse }>(
    `/v1/boards/${boardId}/releases/${releaseId}/complete`,
  )
  return mapRelease(data.release)
}

export async function getActiveRelease(boardId: string): Promise<Release | null> {
  try {
    const { data } = await api.get<{ release: ReleaseResponse | null }>(
      `/v1/boards/${boardId}/releases/active`,
    )
    return data.release ? mapRelease(data.release) : null
  } catch {
    return null
  }
}

export async function getReleaseCards(boardId: string, releaseId: string): Promise<Card[]> {
  const { data } = await api.get<{ cards: CardResponse[] }>(
    `/v1/boards/${boardId}/releases/${releaseId}/cards`,
  )
  return (data.cards || []).map(mapCard)
}

export async function assignCardToRelease(
  boardId: string,
  releaseId: string,
  cardId: string,
): Promise<void> {
  await api.post(`/v1/boards/${boardId}/releases/${releaseId}/cards`, { card_id: cardId })
}

export async function removeCardFromRelease(
  boardId: string,
  releaseId: string,
  cardId: string,
): Promise<void> {
  await api.delete(`/v1/boards/${boardId}/releases/${releaseId}/cards/${cardId}`)
}

export async function getBacklog(boardId: string): Promise<Card[]> {
  const { data } = await api.get<{ cards: CardResponse[] }>(
    `/v1/boards/${boardId}/backlog`,
  )
  return (data.cards || []).map(mapCard)
}
