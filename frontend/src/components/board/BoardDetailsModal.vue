<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as boardsApi from '@/api/boards'

interface Props {
  boardId: string
  boardTitle: string
}

interface ColumnStat {
  id: string
  title: string
  cardCount: number
}

const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'close'): void }>()

const loading = ref(true)
const columnStats = ref<ColumnStat[]>([])
const memberCount = ref(0)
const totalCards = ref(0)

onMounted(async () => {
  try {
    // Один запрос — GetBoard возвращает columns (с card_count) + members
    const { data } = await boardsApi.getBoardRaw(props.boardId)

    memberCount.value = data.members?.length ?? 0

    let total = 0
    const stats: ColumnStat[] = []
    for (const col of data.columns) {
      stats.push({ id: col.id, title: col.title, cardCount: col.card_count })
      total += col.card_count
    }
    columnStats.value = stats
    totalCards.value = total
  } catch (err) {
    console.error('Failed to load board details:', err)
  } finally {
    loading.value = false
  }
})

function handleBackdrop(e: MouseEvent) {
  if (e.target === e.currentTarget) emit('close')
}
</script>

<template>
  <div class="modal-overlay" @click="handleBackdrop">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ boardTitle }}</h3>
        <button class="modal-close" @click="emit('close')">&times;</button>
      </div>

      <div v-if="loading" class="modal-loading">Загрузка...</div>

      <div v-else class="modal-body">
        <div class="stats-section">
          <div class="stat-row stat-row--summary">
            <span>Всего задач</span>
            <span class="stat-value">{{ totalCards }}</span>
          </div>
          <div class="stat-row stat-row--summary">
            <span>Участников</span>
            <span class="stat-value">{{ memberCount }}</span>
          </div>
        </div>

        <div v-if="columnStats.length > 0" class="columns-section">
          <h4>Колонки</h4>
          <div
            v-for="col in columnStats"
            :key="col.id"
            class="stat-row"
          >
            <span class="column-name">{{ col.title }}</span>
            <span class="stat-badge">{{ col.cardCount }}</span>
          </div>
        </div>

        <div v-else class="empty-hint">Нет колонок</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 24px;
  max-width: 420px;
  width: 90%;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  animation: slideIn 0.2s ease-out;
}

@keyframes slideIn {
  from { transform: scale(0.95) translateY(-20px); opacity: 0; }
  to { transform: scale(1) translateY(0); opacity: 1; }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #111827;
}

.modal-close {
  background: none;
  border: none;
  font-size: 24px;
  color: #9ca3af;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}

.modal-close:hover {
  color: #111827;
}

.modal-loading {
  text-align: center;
  padding: 24px;
  color: #6b7280;
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stats-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f9fafb;
  border-radius: 8px;
  font-size: 14px;
  color: #374151;
}

.stat-row--summary {
  background: var(--color-primary-light);
}

.stat-value {
  font-weight: 600;
  color: var(--color-primary);
}

.columns-section h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.column-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 250px;
}

.stat-badge {
  background: #e5e7eb;
  color: #374151;
  font-weight: 600;
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 12px;
}

.empty-hint {
  text-align: center;
  color: #9ca3af;
  font-size: 14px;
  padding: 12px;
}
</style>
