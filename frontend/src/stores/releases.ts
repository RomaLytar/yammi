import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Release, Card } from '@/types/domain'
import * as releasesApi from '@/api/releases'
import { ApiError } from '@/api/client'

export const useReleasesStore = defineStore('releases', () => {
  const releases = ref<Release[]>([])
  const activeRelease = ref<Release | null>(null)
  const backlog = ref<Card[]>([])
  const releaseCards = ref<Card[]>([])
  const currentRelease = ref<Release | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const draftReleases = computed(() => releases.value.filter(r => r.status === 'draft'))
  const completedReleases = computed(() => releases.value.filter(r => r.status === 'completed'))

  async function fetchReleases(boardId: string): Promise<void> {
    try {
      loading.value = true
      error.value = null
      releases.value = await releasesApi.listReleases(boardId)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка загрузки релизов'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchActiveRelease(boardId: string): Promise<void> {
    try {
      activeRelease.value = await releasesApi.getActiveRelease(boardId)
    } catch {
      activeRelease.value = null
    }
  }

  async function fetchRelease(boardId: string, releaseId: string): Promise<void> {
    try {
      loading.value = true
      error.value = null
      currentRelease.value = await releasesApi.getRelease(boardId, releaseId)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка загрузки релиза'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchReleaseCards(boardId: string, releaseId: string): Promise<void> {
    try {
      releaseCards.value = await releasesApi.getReleaseCards(boardId, releaseId)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка загрузки карточек релиза'
      throw err
    }
  }

  async function fetchBacklog(boardId: string): Promise<void> {
    try {
      loading.value = true
      error.value = null
      backlog.value = await releasesApi.getBacklog(boardId)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка загрузки бэклога'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createRelease(boardId: string, name: string, description: string): Promise<Release> {
    try {
      error.value = null
      const release = await releasesApi.createRelease(boardId, { name, description })
      releases.value.unshift(release)
      return release
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка создания релиза'
      throw err
    }
  }

  async function updateRelease(boardId: string, releaseId: string, name: string, description: string, version: number): Promise<void> {
    try {
      error.value = null
      const updated = await releasesApi.updateRelease(boardId, releaseId, { name, description, version })
      const idx = releases.value.findIndex(r => r.id === releaseId)
      if (idx !== -1) releases.value[idx] = updated
      if (currentRelease.value?.id === releaseId) currentRelease.value = updated
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка обновления релиза'
      throw err
    }
  }

  async function deleteRelease(boardId: string, releaseId: string): Promise<void> {
    try {
      error.value = null
      await releasesApi.deleteRelease(boardId, releaseId)
      releases.value = releases.value.filter(r => r.id !== releaseId)
      if (currentRelease.value?.id === releaseId) currentRelease.value = null
      if (activeRelease.value?.id === releaseId) activeRelease.value = null
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка удаления релиза'
      throw err
    }
  }

  async function startRelease(boardId: string, releaseId: string): Promise<void> {
    try {
      error.value = null
      const updated = await releasesApi.startRelease(boardId, releaseId)
      const idx = releases.value.findIndex(r => r.id === releaseId)
      if (idx !== -1) releases.value[idx] = updated
      activeRelease.value = updated
      if (currentRelease.value?.id === releaseId) currentRelease.value = updated

      // Перезагружаем карточки доски — фильтр по активному релизу обновится
      const { useBoardStore } = await import('@/stores/board')
      const boardStore = useBoardStore()
      if (boardStore.boardId) {
        await boardStore.fetchBoard(boardStore.boardId)
      }
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка запуска релиза'
      throw err
    }
  }

  async function completeRelease(boardId: string, releaseId: string): Promise<void> {
    try {
      error.value = null
      const updated = await releasesApi.completeRelease(boardId, releaseId)
      const idx = releases.value.findIndex(r => r.id === releaseId)
      if (idx !== -1) releases.value[idx] = updated
      if (activeRelease.value?.id === releaseId) activeRelease.value = null
      if (currentRelease.value?.id === releaseId) currentRelease.value = updated

      // Перезагружаем карточки доски — бэкенд сбросил release_id у незавершённых задач
      const { useBoardStore } = await import('@/stores/board')
      const boardStore = useBoardStore()
      if (boardStore.boardId) {
        await boardStore.fetchBoard(boardStore.boardId)
      }
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка завершения релиза'
      throw err
    }
  }

  async function assignCard(boardId: string, releaseId: string, cardId: string): Promise<void> {
    try {
      error.value = null
      await releasesApi.assignCardToRelease(boardId, releaseId, cardId)
      // Убираем из бэклога если там была
      backlog.value = backlog.value.filter(c => c.id !== cardId)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка назначения карточки'
      throw err
    }
  }

  async function removeCard(boardId: string, releaseId: string, cardId: string): Promise<void> {
    try {
      error.value = null
      await releasesApi.removeCardFromRelease(boardId, releaseId, cardId)
      releaseCards.value = releaseCards.value.filter(c => c.id !== cardId)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Ошибка удаления карточки из релиза'
      throw err
    }
  }

  function clear(): void {
    releases.value = []
    activeRelease.value = null
    backlog.value = []
    releaseCards.value = []
    currentRelease.value = null
    error.value = null
  }

  return {
    releases,
    activeRelease,
    backlog,
    releaseCards,
    currentRelease,
    draftReleases,
    completedReleases,
    loading,
    error,
    fetchReleases,
    fetchActiveRelease,
    fetchRelease,
    fetchReleaseCards,
    fetchBacklog,
    createRelease,
    updateRelease,
    deleteRelease,
    startRelease,
    completeRelease,
    assignCard,
    removeCard,
    clear,
  }
})
