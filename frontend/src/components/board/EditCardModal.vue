<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue'
import type {
  Card, Comment, Attachment, ActivityEntry,
  Label, Checklist, ChecklistItem, CardLink,
  Priority, TaskType,
} from '@/types/domain'
import BaseModal from '@/components/shared/BaseModal.vue'
import BaseInput from '@/components/shared/BaseInput.vue'
import BaseButton from '@/components/shared/BaseButton.vue'
import BaseSearchSelect from '@/components/shared/BaseSearchSelect.vue'
import RichTextEditor from '@/components/shared/RichTextEditor.vue'
import BaseSpinner from '@/components/shared/BaseSpinner.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import { useBoardStore } from '@/stores/board'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'
import * as boardsApi from '@/api/boards'

interface Props {
  card: Card
  canDelete?: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'update', data: {
    title: string; description: string; assigneeId?: string;
    dueDate?: string; priority?: string; taskType?: string
  }): void
  (e: 'delete'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const boardStore = useBoardStore()
const userStore = useUserStore()
const authStore = useAuthStore()

// --- Sidebar collapsible blocks ---
const infoOpen = ref(true)
const activityOpen = ref(true)
const showAssigneeSelect = ref(false)

// --- Card fields ---
const title = ref(props.card.title)
const description = ref(props.card.description)
const selectedAssignee = ref(props.card.assigneeId || '')
const selectedPriority = ref<Priority>(props.card.priority || 'medium')
const selectedTaskType = ref<TaskType>(props.card.taskType || 'task')
const selectedDueDate = ref(props.card.dueDate || '')
const loading = ref(false)
const showConfirmDelete = ref(false)
const lightboxUrl = ref<string | null>(null)

// --- Labels ---
const cardLabels = ref<Label[]>([])
const labelsLoading = ref(false)
const showLabelPicker = ref(false)

// --- Checklists ---
const checklists = ref<Checklist[]>([])
const checklistsLoading = ref(false)
const checklistsLoaded = ref(false)
const newChecklistTitle = ref('')
const newItemTitles = ref<Record<string, string>>({})

// --- Subtasks (Card Links) ---
const childLinks = ref<CardLink[]>([])
const parentLinks = ref<CardLink[]>([])
const subtasksLoading = ref(false)
const subtasksLoaded = ref(false)
const showSubtaskSearch = ref(false)
const subtaskSearchQuery = ref('')

watch(() => props.card, (newCard) => {
  title.value = newCard.title
  description.value = newCard.description
  selectedAssignee.value = newCard.assigneeId || ''
  selectedPriority.value = newCard.priority || 'medium'
  selectedTaskType.value = newCard.taskType || 'task'
  selectedDueDate.value = newCard.dueDate || ''
})

const creatorName = computed(() => {
  if (props.card.creatorId === userStore.profile?.id) return 'Вы'
  return boardStore.getMemberName(props.card.creatorId)
})

const assigneeOptions = computed(() =>
  boardStore.members.map(m => ({
    value: m.user_id,
    label: boardStore.getMemberName(m.user_id),
    sublabel: m.role === 'owner' ? 'владелец' : boardStore.getMemberEmail(m.user_id),
  }))
)

function handleUpdate() {
  if (!title.value.trim()) return
  loading.value = true
  emit('update', {
    title: title.value.trim(),
    description: description.value.trim(),
    assigneeId: selectedAssignee.value || undefined,
    dueDate: selectedDueDate.value || undefined,
    priority: selectedPriority.value,
    taskType: selectedTaskType.value,
  })
}

function handleDelete() {
  showConfirmDelete.value = true
}

function confirmDelete() {
  loading.value = true
  showConfirmDelete.value = false
  emit('delete')
}

function cancelDelete() {
  showConfirmDelete.value = false
}

function handleClose() {
  if (!loading.value) {
    emit('close')
  }
}

// --- Comments ---
const comments = ref<Comment[]>([])
const commentsLoading = ref(false)
const commentsNextCursor = ref<string | undefined>(undefined)
const commentsLoaded = ref(false)
const commentText = ref('')
const commentSubmitting = ref(false)
const replyingTo = ref<Comment | null>(null)
const editingComment = ref<Comment | null>(null)
const editCommentText = ref('')
const deletingCommentId = ref<string | null>(null)

function getMemberName(userId: string): string {
  if (userId === authStore.userId) return 'Вы'
  return boardStore.getMemberName(userId)
}

async function loadComments() {
  if (!boardStore.boardId) return
  commentsLoading.value = true
  try {
    const result = await boardsApi.listComments(props.card.id, boardStore.boardId, 20)
    comments.value = result.comments
    commentsNextCursor.value = result.nextCursor
    commentsLoaded.value = true
  } catch (err) {
    console.error('Failed to load comments:', err)
  } finally {
    commentsLoading.value = false
  }
}

async function loadMoreComments() {
  if (!boardStore.boardId || !commentsNextCursor.value) return
  commentsLoading.value = true
  try {
    const result = await boardsApi.listComments(
      props.card.id,
      boardStore.boardId,
      20,
      commentsNextCursor.value,
    )
    comments.value.push(...result.comments)
    commentsNextCursor.value = result.nextCursor
  } catch (err) {
    console.error('Failed to load more comments:', err)
  } finally {
    commentsLoading.value = false
  }
}

// Группировка: root комменты + их replies вложенные
const threadedComments = computed(() => {
  const roots: Comment[] = []
  const repliesMap = new Map<string, Comment[]>()

  for (const c of comments.value) {
    if (!c.parentId) {
      roots.push(c)
    } else {
      if (!repliesMap.has(c.parentId)) repliesMap.set(c.parentId, [])
      repliesMap.get(c.parentId)!.push(c)
    }
  }

  // Сортируем replies по дате (старые сверху)
  for (const replies of repliesMap.values()) {
    replies.sort((a, b) => a.createdAt.localeCompare(b.createdAt))
  }

  return { roots, repliesMap }
})

async function submitComment() {
  if (!boardStore.boardId || !commentText.value.trim()) return
  commentSubmitting.value = true
  try {
    const text = commentText.value.trim()
    const isReply = !!replyingTo.value
    const comment = await boardsApi.createComment(
      props.card.id,
      boardStore.boardId,
      text,
      replyingTo.value?.id,
    )
    comments.value.unshift(comment)
    commentText.value = ''
    replyingTo.value = null

    const preview = truncate(text, 30)
    addLocalActivity(
      'comment_added',
      isReply ? `Ответ на комментарий: "${preview}"` : `Комментарий: "${preview}"`,
      { comment_text: preview },
    )
  } catch (err) {
    console.error('Failed to create comment:', err)
  } finally {
    commentSubmitting.value = false
  }
}

function startReply(comment: Comment) {
  replyingTo.value = comment
  editingComment.value = null
}

function cancelReply() {
  replyingTo.value = null
}

function startEditComment(comment: Comment) {
  editingComment.value = comment
  editCommentText.value = comment.content
  replyingTo.value = null
}

function cancelEditComment() {
  editingComment.value = null
  editCommentText.value = ''
}

async function submitEditComment() {
  if (!boardStore.boardId || !editingComment.value || !editCommentText.value.trim()) return
  commentSubmitting.value = true
  try {
    const updated = await boardsApi.updateComment(
      editingComment.value.id,
      boardStore.boardId,
      editCommentText.value.trim(),
    )
    const idx = comments.value.findIndex(c => c.id === updated.id)
    if (idx !== -1) comments.value[idx] = updated
    editingComment.value = null
    editCommentText.value = ''
  } catch (err) {
    console.error('Failed to update comment:', err)
  } finally {
    commentSubmitting.value = false
  }
}

async function handleDeleteComment(commentId: string) {
  if (!boardStore.boardId) return
  deletingCommentId.value = commentId
  try {
    const deleted = comments.value.find(c => c.id === commentId)
    await boardsApi.deleteComment(commentId, boardStore.boardId)
    comments.value = comments.value.filter(c => c.id !== commentId)
    if (deleted) {
      const preview = truncate(deleted.content, 30)
      addLocalActivity('comment_deleted', `Комментарий удалён: "${preview}"`, { comment_text: preview })
    }
  } catch (err) {
    console.error('Failed to delete comment:', err)
  } finally {
    deletingCommentId.value = null
  }
}

// --- Attachments ---
const attachments = ref<Attachment[]>([])
const attachmentsLoading = ref(false)
const attachmentsLoaded = ref(false)
const uploadProgress = ref<number | null>(null)
const uploadError = ref<string | null>(null)
const deletingAttachmentId = ref<string | null>(null)

async function loadAttachments() {
  if (!boardStore.boardId) return
  attachmentsLoading.value = true
  try {
    attachments.value = await boardsApi.listAttachments(props.card.id, boardStore.boardId)
    attachmentsLoaded.value = true
    loadAttachmentPreviews()
  } catch (err) {
    console.error('Failed to load attachments:', err)
  } finally {
    attachmentsLoading.value = false
  }
}

async function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file || !boardStore.boardId) return

  uploadError.value = null
  uploadProgress.value = 0

  try {
    const { attachment, uploadUrl } = await boardsApi.createUploadURL(
      props.card.id,
      boardStore.boardId,
      file.name,
      file.type || 'application/octet-stream',
      file.size,
    )

    await boardsApi.uploadFileToPresignedUrl(uploadUrl, file, (percent) => {
      uploadProgress.value = percent
    })

    const confirmed = await boardsApi.confirmUpload(attachment.id, boardStore.boardId!)
    attachments.value.unshift(confirmed)
    uploadProgress.value = null

    // Загружаем превью для нового файла
    if (isImage(confirmed.mimeType)) {
      try {
        const url = await boardsApi.getDownloadURL(confirmed.id, boardStore.boardId!)
        attachmentUrls.value.set(confirmed.id, url)
      } catch { /* ignore */ }
    }

    // Добавляем запись в историю локально
    addLocalActivity('attachment_added', `Файл "${file.name}" прикреплён`, { file_name: file.name })
  } catch (err) {
    console.error('Failed to upload file:', err)
    uploadError.value = 'Ошибка загрузки файла'
    uploadProgress.value = null
  }

  target.value = ''
}

// Кеш presigned URLs для превью картинок
const attachmentUrls = ref<Map<string, string>>(new Map())

function isImage(mimeType: string): boolean {
  return mimeType.startsWith('image/')
}

async function loadAttachmentPreviews() {
  if (!boardStore.boardId) return
  for (const a of attachments.value) {
    if (isImage(a.mimeType) && !attachmentUrls.value.has(a.id)) {
      try {
        const url = await boardsApi.getDownloadURL(a.id, boardStore.boardId)
        attachmentUrls.value.set(a.id, url)
      } catch { /* ignore */ }
    }
  }
}

async function handleOpenFile(attachment: Attachment) {
  if (!boardStore.boardId) return
  try {
    let url = attachmentUrls.value.get(attachment.id)
    if (!url) {
      url = await boardsApi.getDownloadURL(attachment.id, boardStore.boardId)
      attachmentUrls.value.set(attachment.id, url)
    }
    if (isImage(attachment.mimeType)) {
      lightboxUrl.value = url
    } else {
      window.open(url, '_blank')
    }
  } catch (err) {
    console.error('Failed to open file:', err)
  }
}

async function handleDownload(attachment: Attachment) {
  if (!boardStore.boardId) return
  try {
    const url = await boardsApi.getDownloadURL(attachment.id, boardStore.boardId)
    window.open(url, '_blank')
  } catch (err) {
    console.error('Failed to get download URL:', err)
  }
}

async function handleDeleteAttachment(attachmentId: string) {
  if (!boardStore.boardId) return
  deletingAttachmentId.value = attachmentId
  try {
    const deleted = attachments.value.find(a => a.id === attachmentId)
    await boardsApi.deleteAttachment(attachmentId, boardStore.boardId)
    attachments.value = attachments.value.filter(a => a.id !== attachmentId)
    if (deleted) {
      addLocalActivity('attachment_deleted', `Файл "${deleted.fileName}" удалён`, { file_name: deleted.fileName })
    }
  } catch (err) {
    console.error('Failed to delete attachment:', err)
  } finally {
    deletingAttachmentId.value = null
  }
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' Б'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' КБ'
  return (bytes / (1024 * 1024)).toFixed(1) + ' МБ'
}

function canDeleteAttachment(attachment: Attachment): boolean {
  return attachment.uploaderId === authStore.userId
    || boardStore.board?.ownerId === authStore.userId
}

// --- Activity ---
const activityEntries = ref<ActivityEntry[]>([])
const activityLoading = ref(false)
const activityLoaded = ref(false)
const activityNextCursor = ref<string | undefined>(undefined)

async function loadActivity() {
  if (!boardStore.boardId) return
  activityLoading.value = true
  try {
    const result = await boardsApi.getCardActivity(props.card.id, boardStore.boardId, 20)
    activityEntries.value = result.entries
    activityNextCursor.value = result.nextCursor
    activityLoaded.value = true
  } catch (err) {
    console.error('Failed to load activity:', err)
  } finally {
    activityLoading.value = false
  }
}

async function loadMoreActivity() {
  if (!boardStore.boardId || !activityNextCursor.value) return
  activityLoading.value = true
  try {
    const result = await boardsApi.getCardActivity(
      props.card.id,
      boardStore.boardId,
      20,
      activityNextCursor.value,
    )
    activityEntries.value.push(...result.entries)
    activityNextCursor.value = result.nextCursor
  } catch (err) {
    console.error('Failed to load more activity:', err)
  } finally {
    activityLoading.value = false
  }
}

function truncate(text: string, max: number): string {
  return text.length > max ? text.slice(0, max) + '...' : text
}

// Добавляет запись в историю локально (без перезагрузки)
function addLocalActivity(type: string, description: string, changes: Record<string, string> = {}) {
  activityEntries.value.unshift({
    id: crypto.randomUUID(),
    cardId: props.card.id,
    boardId: boardStore.boardId || '',
    actorId: authStore.userId || '',
    activityType: type,
    description,
    changes,
    createdAt: new Date().toISOString(),
  })
}

function formatActivity(entry: ActivityEntry): string {
  switch (entry.activityType) {
    case 'card_created': return 'Создана'
    case 'card_updated': {
      const parts: string[] = []
      if (entry.changes?.old_title && entry.changes?.new_title) {
        parts.push(`Название: "${entry.changes.old_title}" → "${entry.changes.new_title}"`)
      }
      if (entry.changes?.description_changed) {
        parts.push('Описание изменено')
      }
      return parts.length > 0 ? parts.join('. ') : 'Обновлена'
    }
    case 'card_moved': {
      const from = entry.changes?.from_column_id?.slice(0, 6) || ''
      const to = entry.changes?.to_column_id?.slice(0, 6) || ''
      return from && to ? `Перемещена` : 'Перемещена'
    }
    case 'card_assigned': {
      const assignee = entry.changes?.assignee_id
      const name = assignee ? boardStore.getMemberName(assignee) : ''
      return name ? `Назначена на ${name}` : 'Назначена'
    }
    case 'card_unassigned': return 'Назначение снято'
    case 'card_deleted': return 'Удалена'
    case 'attachment_added': return `Файл "${truncate(entry.changes?.file_name || '', 30)}" прикреплён`
    case 'attachment_deleted': return `Файл "${truncate(entry.changes?.file_name || '', 30)}" удалён`
    case 'comment_added': return entry.changes?.comment_text
      ? `Комментарий: "${truncate(entry.changes.comment_text, 30)}"`
      : 'Комментарий добавлен'
    case 'comment_deleted': return entry.changes?.comment_text
      ? `Комментарий удалён: "${truncate(entry.changes.comment_text, 30)}"`
      : 'Комментарий удалён'
    default: return entry.description
  }
}

function activityIconSvg(type: string): string {
  switch (type) {
    case 'card_created': return '<line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>'
    case 'card_updated': return '<path d="M17 3a2.85 2.85 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/>'
    case 'card_moved': return '<line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>'
    case 'card_assigned': return '<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>'
    case 'card_unassigned': return '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><line x1="17" y1="11" x2="23" y2="11"/>'
    case 'attachment_added': return '<path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48"/>'
    case 'attachment_deleted': return '<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>'
    case 'comment_added': return '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>'
    case 'comment_deleted': return '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/><line x1="9" y1="9" x2="15" y2="15"/><line x1="15" y1="9" x2="9" y2="15"/>'
    default: return '<circle cx="12" cy="12" r="1"/>'
  }
}

// --- Load activity when block opens ---
watch(activityOpen, (open) => {
  if (open && !activityLoaded.value) loadActivity()
})

// --- Labels ---
async function loadCardLabels() {
  if (!boardStore.boardId) return
  labelsLoading.value = true
  try {
    cardLabels.value = await boardsApi.getCardLabels(boardStore.boardId, props.card.id)
  } catch (err) {
    console.error('Failed to load card labels:', err)
  } finally {
    labelsLoading.value = false
  }
}

function isLabelAssigned(labelId: string): boolean {
  return cardLabels.value.some(l => l.id === labelId)
}

async function toggleLabel(label: Label) {
  if (!boardStore.boardId) return
  try {
    if (isLabelAssigned(label.id)) {
      await boardsApi.removeLabelFromCard(boardStore.boardId, props.card.id, label.id)
      cardLabels.value = cardLabels.value.filter(l => l.id !== label.id)
    } else {
      await boardsApi.addLabelToCard(boardStore.boardId, props.card.id, label.id)
      cardLabels.value.push(label)
    }
  } catch (err) {
    console.error('Failed to toggle label:', err)
  }
}

// --- Checklists ---
async function loadChecklists() {
  if (!boardStore.boardId) return
  checklistsLoading.value = true
  try {
    checklists.value = await boardsApi.getChecklists(boardStore.boardId, props.card.id)
    checklistsLoaded.value = true
  } catch (err) {
    console.error('Failed to load checklists:', err)
  } finally {
    checklistsLoading.value = false
  }
}

async function addChecklist() {
  if (!boardStore.boardId || !newChecklistTitle.value.trim()) return
  try {
    const cl = await boardsApi.createChecklist(boardStore.boardId, props.card.id, newChecklistTitle.value.trim())
    checklists.value.push(cl)
    newChecklistTitle.value = ''
  } catch (err) {
    console.error('Failed to create checklist:', err)
  }
}

async function removeChecklist(checklistId: string) {
  if (!boardStore.boardId) return
  try {
    await boardsApi.deleteChecklist(boardStore.boardId, checklistId)
    checklists.value = checklists.value.filter(c => c.id !== checklistId)
  } catch (err) {
    console.error('Failed to delete checklist:', err)
  }
}

async function addChecklistItem(checklistId: string) {
  if (!boardStore.boardId) return
  const itemTitle = newItemTitles.value[checklistId]?.trim()
  if (!itemTitle) return
  try {
    const item = await boardsApi.createChecklistItem(boardStore.boardId, checklistId, itemTitle)
    const cl = checklists.value.find(c => c.id === checklistId)
    if (cl) {
      cl.items.push(item)
      cl.progress = calcProgress(cl.items)
    }
    newItemTitles.value[checklistId] = ''
  } catch (err) {
    console.error('Failed to create checklist item:', err)
  }
}

async function toggleItem(checklistId: string, itemId: string) {
  if (!boardStore.boardId) return
  try {
    const updated = await boardsApi.toggleChecklistItem(boardStore.boardId, itemId)
    const cl = checklists.value.find(c => c.id === checklistId)
    if (cl) {
      const idx = cl.items.findIndex(i => i.id === itemId)
      if (idx !== -1) cl.items[idx] = updated
      cl.progress = calcProgress(cl.items)
    }
  } catch (err) {
    console.error('Failed to toggle checklist item:', err)
  }
}

async function removeChecklistItem(checklistId: string, itemId: string) {
  if (!boardStore.boardId) return
  try {
    await boardsApi.deleteChecklistItem(boardStore.boardId, itemId)
    const cl = checklists.value.find(c => c.id === checklistId)
    if (cl) {
      cl.items = cl.items.filter(i => i.id !== itemId)
      cl.progress = calcProgress(cl.items)
    }
  } catch (err) {
    console.error('Failed to delete checklist item:', err)
  }
}

function calcProgress(items: ChecklistItem[]): number {
  if (items.length === 0) return 0
  return Math.round((items.filter(i => i.isChecked).length / items.length) * 100)
}

// --- Subtasks (Card Links) ---
async function loadSubtasks() {
  if (!boardStore.boardId) return
  subtasksLoading.value = true
  try {
    const [children, parents] = await Promise.all([
      boardsApi.getCardChildren(boardStore.boardId, props.card.id),
      boardsApi.getCardParents(boardStore.boardId, props.card.id),
    ])
    childLinks.value = children
    parentLinks.value = parents
    subtasksLoaded.value = true
  } catch (err) {
    console.error('Failed to load subtasks:', err)
  } finally {
    subtasksLoading.value = false
  }
}

// Simple search: find cards across all columns matching query
const subtaskSearchResults = computed(() => {
  if (!subtaskSearchQuery.value.trim()) return []
  const q = subtaskSearchQuery.value.toLowerCase()
  const results: { id: string; title: string; columnName: string }[] = []
  const linkedIds = new Set([
    props.card.id,
    ...childLinks.value.map(l => l.childId),
    ...parentLinks.value.map(l => l.parentId),
  ])
  for (const col of boardStore.columns) {
    for (const c of col.cards) {
      if (linkedIds.has(c.id)) continue
      if (c.title.toLowerCase().includes(q)) {
        results.push({ id: c.id, title: c.title, columnName: col.title })
      }
    }
  }
  return results.slice(0, 10)
})

async function linkSubtask(childCardId: string) {
  if (!boardStore.boardId) return
  try {
    const link = await boardsApi.linkCards(boardStore.boardId, props.card.id, childCardId)
    // Enrich link with title/column info
    for (const col of boardStore.columns) {
      const c = col.cards.find(card => card.id === childCardId)
      if (c) {
        link.childTitle = c.title
        link.childColumnName = col.title
        break
      }
    }
    childLinks.value.push(link)
    subtaskSearchQuery.value = ''
    showSubtaskSearch.value = false
  } catch (err) {
    console.error('Failed to link card:', err)
  }
}

async function unlinkSubtask(linkId: string) {
  if (!boardStore.boardId) return
  try {
    await boardsApi.unlinkCards(boardStore.boardId, linkId)
    childLinks.value = childLinks.value.filter(l => l.id !== linkId)
  } catch (err) {
    console.error('Failed to unlink card:', err)
  }
}

// --- Load comments, attachments, activity, labels, checklists, subtasks on mount ---
onMounted(() => {
  loadComments()
  loadActivity()
  loadAttachments()
  loadCardLabels()
  loadChecklists()
  loadSubtasks()
})
</script>

<template>
  <BaseModal size="large" @close="handleClose">
    <template #header>
      <div class="ecm-header">
        <div class="ecm-header__icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M17 3a2.85 2.85 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
          </svg>
        </div>
        <h2 class="ecm-header__title">Редактирование</h2>
      </div>
    </template>

    <div class="ecm-layout">
      <!-- Left column: main content -->
      <div class="ecm-main">
        <!-- Title -->
        <BaseInput
          v-model="title"
          :disabled="loading"
          placeholder="Название карточки"
        />

        <!-- Description -->
        <div class="ecm-section">
          <div class="ecm-section__label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <line x1="17" y1="10" x2="3" y2="10" /><line x1="21" y1="6" x2="3" y2="6" /><line x1="21" y1="14" x2="3" y2="14" /><line x1="17" y1="18" x2="3" y2="18" />
            </svg>
            Описание
          </div>
          <RichTextEditor
            v-model="description"
            :disabled="loading"
            placeholder="Описание карточки..."
          />
        </div>

        <!-- Files section -->
        <div class="ecm-section">
          <div class="ecm-section__label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
            </svg>
            Файлы
            <span v-if="attachments.length" class="ecm-badge">{{ attachments.length }}</span>
          </div>

          <!-- Upload area -->
          <div class="ecm-upload">
            <label class="ecm-upload__label">
              <input
                type="file"
                class="ecm-upload__input"
                :disabled="uploadProgress !== null"
                @change="handleFileSelect"
              />
              <span class="ecm-upload__zone">
                <div class="ecm-upload__zone-icon">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                    <polyline points="17 8 12 3 7 8"/>
                    <line x1="12" y1="3" x2="12" y2="15"/>
                  </svg>
                </div>
                <span>Прикрепить файл</span>
              </span>
            </label>
            <div v-if="uploadProgress !== null" class="ecm-upload__progress">
              <div class="ecm-upload__progress-bar">
                <div class="ecm-upload__progress-fill" :style="{ width: uploadProgress + '%' }" />
              </div>
              <span class="ecm-upload__progress-text">{{ uploadProgress }}%</span>
            </div>
            <div v-if="uploadError" class="ecm-upload__error">{{ uploadError }}</div>
          </div>

          <!-- File list -->
          <div v-if="attachmentsLoading && !attachmentsLoaded" class="ecm-section__center">
            <BaseSpinner size="sm" />
          </div>
          <div v-else-if="attachments.length === 0 && attachmentsLoaded" class="ecm-section__empty">
            Файлов пока нет
          </div>
          <div v-else class="ecm-files">
            <div
              v-for="attachment in attachments"
              :key="attachment.id"
              class="ecm-file"
              @click="handleOpenFile(attachment)"
            >
              <!-- Image preview -->
              <div
                v-if="isImage(attachment.mimeType)"
                class="ecm-file__preview"
              >
                <img
                  :src="attachmentUrls.get(attachment.id) || ''"
                  :alt="attachment.fileName"
                  class="ecm-file__image"
                  loading="lazy"
                />
              </div>
              <!-- Non-image file icon -->
              <div v-else class="ecm-file__icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              </div>
              <div class="ecm-file__body">
                <span class="ecm-file__name">{{ attachment.fileName }}</span>
                <span class="ecm-file__meta">{{ formatFileSize(attachment.fileSize) }}</span>
              </div>
              <button
                v-if="canDeleteAttachment(attachment)"
                class="ecm-file__btn ecm-file__btn--danger"
                title="Удалить"
                :disabled="deletingAttachmentId === attachment.id"
                @click.stop="handleDeleteAttachment(attachment.id)"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Checklists section -->
        <div class="ecm-section">
          <div class="ecm-section__label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <polyline points="9 11 12 14 22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
            </svg>
            Чек-листы
            <span v-if="checklists.length" class="ecm-badge">{{ checklists.length }}</span>
          </div>

          <div v-if="checklistsLoading && !checklistsLoaded" class="ecm-section__center">
            <BaseSpinner size="sm" />
          </div>
          <div v-else>
            <!-- Existing checklists -->
            <div v-for="cl in checklists" :key="cl.id" class="ecm-checklist">
              <div class="ecm-checklist__header">
                <span class="ecm-checklist__title">{{ cl.title }}</span>
                <span class="ecm-checklist__progress-text">{{ calcProgress(cl.items) }}%</span>
                <button class="ecm-checklist__remove" title="Удалить чек-лист" @click="removeChecklist(cl.id)">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <div class="ecm-checklist__bar">
                <div class="ecm-checklist__bar-fill" :style="{ width: calcProgress(cl.items) + '%' }" />
              </div>
              <div class="ecm-checklist__items">
                <div v-for="item in cl.items" :key="item.id" class="ecm-checklist-item">
                  <input
                    type="checkbox"
                    :checked="item.isChecked"
                    class="ecm-checklist-item__check"
                    @change="toggleItem(cl.id, item.id)"
                  />
                  <span class="ecm-checklist-item__title" :class="{ 'ecm-checklist-item__title--done': item.isChecked }">{{ item.title }}</span>
                  <button class="ecm-checklist-item__remove" @click="removeChecklistItem(cl.id, item.id)">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </div>
              </div>
              <div class="ecm-checklist__add-item">
                <input
                  v-model="newItemTitles[cl.id]"
                  class="ecm-checklist__add-input"
                  placeholder="Добавить пункт..."
                  @keydown.enter="addChecklistItem(cl.id)"
                />
                <button class="ecm-checklist__add-btn" :disabled="!newItemTitles[cl.id]?.trim()" @click="addChecklistItem(cl.id)">+</button>
              </div>
            </div>

            <!-- Add checklist -->
            <div class="ecm-checklist-add">
              <input
                v-model="newChecklistTitle"
                class="ecm-checklist-add__input"
                placeholder="Новый чек-лист..."
                @keydown.enter="addChecklist"
              />
              <BaseButton size="sm" :disabled="!newChecklistTitle.trim()" @click="addChecklist">
                Добавить
              </BaseButton>
            </div>
          </div>
        </div>

        <!-- Subtasks section -->
        <div class="ecm-section">
          <div class="ecm-section__label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <rect x="3" y="3" width="18" height="18" rx="2"/>
              <path d="M9 12h6M12 9v6"/>
            </svg>
            Подзадачи
            <span v-if="childLinks.length" class="ecm-badge">{{ childLinks.length }}</span>
          </div>

          <div v-if="subtasksLoading && !subtasksLoaded" class="ecm-section__center">
            <BaseSpinner size="sm" />
          </div>
          <div v-else>
            <!-- Parent links -->
            <div v-if="parentLinks.length > 0" class="ecm-subtask-parents">
              <span class="ecm-subtask-parents__label">Родитель:</span>
              <span v-for="p in parentLinks" :key="p.id" class="ecm-subtask-parent">{{ p.childTitle || p.parentId.slice(0,8) }}</span>
            </div>

            <!-- Child links -->
            <div v-if="childLinks.length === 0 && subtasksLoaded" class="ecm-section__empty">
              Подзадач пока нет
            </div>
            <div v-else class="ecm-subtask-list">
              <div v-for="link in childLinks" :key="link.id" class="ecm-subtask-item">
                <div class="ecm-subtask-item__body">
                  <span class="ecm-subtask-item__title">{{ link.childTitle || link.childId.slice(0,8) }}</span>
                  <span v-if="link.childColumnName" class="ecm-subtask-item__column">{{ link.childColumnName }}</span>
                </div>
                <button class="ecm-subtask-item__unlink" title="Отвязать" @click="unlinkSubtask(link.id)">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
            </div>

            <!-- Link subtask search -->
            <button v-if="!showSubtaskSearch" class="ecm-subtask-add-btn" @click="showSubtaskSearch = true">
              + Привязать подзадачу
            </button>
            <div v-else class="ecm-subtask-search">
              <input
                v-model="subtaskSearchQuery"
                class="ecm-subtask-search__input"
                placeholder="Поиск карточки..."
                autofocus
              />
              <button class="ecm-subtask-search__close" @click="showSubtaskSearch = false; subtaskSearchQuery = ''">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
              <div v-if="subtaskSearchResults.length > 0" class="ecm-subtask-search__results">
                <div
                  v-for="result in subtaskSearchResults"
                  :key="result.id"
                  class="ecm-subtask-search__result"
                  @click="linkSubtask(result.id)"
                >
                  <span class="ecm-subtask-search__result-title">{{ result.title }}</span>
                  <span class="ecm-subtask-search__result-col">{{ result.columnName }}</span>
                </div>
              </div>
              <div v-else-if="subtaskSearchQuery.trim()" class="ecm-subtask-search__empty">
                Ничего не найдено
              </div>
            </div>
          </div>
        </div>

        <!-- Comments section -->
        <div class="ecm-section">
          <div class="ecm-section__label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
            Комментарии
            <span v-if="comments.length" class="ecm-badge">{{ comments.length }}</span>
          </div>

          <!-- Comment form -->
          <div class="ecm-comment-form">
            <div v-if="replyingTo" class="ecm-comment-form__reply">
              <span class="ecm-comment-form__reply-text">
                Ответ на: {{ getMemberName(replyingTo.authorId) }}
              </span>
              <button class="ecm-comment-form__reply-cancel" @click="cancelReply">&times;</button>
            </div>
            <textarea
              v-model="commentText"
              class="ecm-comment-form__textarea"
              placeholder="Написать комментарий..."
              rows="3"
              :disabled="commentSubmitting"
            />
            <BaseButton
              size="sm"
              :loading="commentSubmitting"
              :disabled="!commentText.trim()"
              @click="submitComment"
            >
              Отправить
            </BaseButton>
          </div>

          <!-- Comments list -->
          <div v-if="commentsLoading && !commentsLoaded" class="ecm-section__center">
            <BaseSpinner size="sm" />
          </div>
          <div v-else-if="comments.length === 0 && commentsLoaded" class="ecm-section__empty">
            Комментариев пока нет
          </div>
          <div v-else class="ecm-comments">
            <template v-for="root in threadedComments.roots" :key="root.id">
            <!-- Root comment -->
            <div class="ecm-comment">

              <!-- Reuse comment template for root -->
              <div v-if="editingComment?.id === root.id" class="ecm-comment__edit">
                <BaseInput v-model="editCommentText" type="textarea" :disabled="commentSubmitting" />
                <div class="ecm-comment__edit-actions">
                  <BaseButton size="sm" :loading="commentSubmitting" :disabled="!editCommentText.trim()" @click="submitEditComment">Сохранить</BaseButton>
                  <BaseButton size="sm" variant="secondary" @click="cancelEditComment">Отмена</BaseButton>
                </div>
              </div>
              <template v-else>
                <div class="ecm-comment__header">
                  <span class="ecm-comment__author">{{ getMemberName(root.authorId) }}</span>
                  <span class="ecm-comment__time">{{ new Date(root.createdAt).toLocaleString('ru-RU') }}</span>
                </div>
                <div class="ecm-comment__content">{{ root.content }}</div>
                <div class="ecm-comment__actions">
                  <button class="ecm-comment__action" @click="startReply(root)">Ответить</button>
                  <button v-if="root.authorId === authStore.userId" class="ecm-comment__action" @click="startEditComment(root)">Изменить</button>
                  <button v-if="root.authorId === authStore.userId" class="ecm-comment__action ecm-comment__action--danger" :disabled="deletingCommentId === root.id" @click="handleDeleteComment(root.id)">{{ deletingCommentId === root.id ? '...' : 'Удалить' }}</button>
                </div>
              </template>
            </div>

            <!-- Replies -->
            <div
              v-for="reply in threadedComments.repliesMap.get(root.id) || []"
              :key="reply.id"
              class="ecm-comment ecm-comment--reply"
            >
              <div v-if="editingComment?.id === reply.id" class="ecm-comment__edit">
                <BaseInput v-model="editCommentText" type="textarea" :disabled="commentSubmitting" />
                <div class="ecm-comment__edit-actions">
                  <BaseButton size="sm" :loading="commentSubmitting" :disabled="!editCommentText.trim()" @click="submitEditComment">Сохранить</BaseButton>
                  <BaseButton size="sm" variant="secondary" @click="cancelEditComment">Отмена</BaseButton>
                </div>
              </div>
              <template v-else>
                <div class="ecm-comment__header">
                  <span class="ecm-comment__author">{{ getMemberName(reply.authorId) }}</span>
                  <span class="ecm-comment__time">{{ new Date(reply.createdAt).toLocaleString('ru-RU') }}</span>
                </div>
                <div class="ecm-comment__content">{{ reply.content }}</div>
                <div class="ecm-comment__actions">
                  <button v-if="reply.authorId === authStore.userId" class="ecm-comment__action" @click="startEditComment(reply)">Изменить</button>
                  <button v-if="reply.authorId === authStore.userId" class="ecm-comment__action ecm-comment__action--danger" :disabled="deletingCommentId === reply.id" @click="handleDeleteComment(reply.id)">{{ deletingCommentId === reply.id ? '...' : 'Удалить' }}</button>
                </div>
              </template>
            </div>
            </template>
          </div>

          <div v-if="commentsNextCursor" class="ecm-section__load-more">
            <BaseButton
              variant="ghost"
              size="sm"
              :loading="commentsLoading"
              @click="loadMoreComments"
            >
              Загрузить ещё
            </BaseButton>
          </div>
        </div>
      </div>

      <!-- Right column: sidebar -->
      <div class="ecm-sidebar">
        <!-- Сведения (collapsible) -->
        <div class="ecm-block">
          <button class="ecm-block__header" @click="infoOpen = !infoOpen">
            <span class="ecm-block__title">Сведения</span>
            <svg class="ecm-block__chevron" :class="{ 'ecm-block__chevron--open': infoOpen }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="6 9 12 15 18 9"/></svg>
          </button>
          <div v-show="infoOpen" class="ecm-block__body">
            <!-- Исполнитель -->
            <div class="ecm-field">
              <span class="ecm-field__label">Исполнитель</span>
              <div class="ecm-field__user" @click="showAssigneeSelect = !showAssigneeSelect">
                <div class="ecm-avatar" :class="{ 'ecm-avatar--empty': !selectedAssignee }">
                  {{ selectedAssignee ? boardStore.getMemberName(selectedAssignee).charAt(0).toUpperCase() : '?' }}
                </div>
                <span class="ecm-field__name">{{ selectedAssignee ? boardStore.getMemberName(selectedAssignee) : 'Не назначен' }}</span>
              </div>
              <BaseSearchSelect
                v-if="showAssigneeSelect"
                v-model="selectedAssignee"
                :options="assigneeOptions"
                placeholder="Поиск участника..."
                :disabled="loading"
                clearable
                style="margin-top: 6px"
              />
            </div>

            <!-- Автор -->
            <div class="ecm-field">
              <span class="ecm-field__label">Автор</span>
              <div class="ecm-field__user">
                <div class="ecm-avatar">{{ creatorName.charAt(0).toUpperCase() }}</div>
                <span class="ecm-field__name">{{ creatorName }}</span>
              </div>
            </div>

            <!-- Приоритет -->
            <div class="ecm-field">
              <span class="ecm-field__label">Приоритет</span>
              <div class="ecm-priority-btns">
                <button
                  class="ecm-priority-btn"
                  :class="{ 'ecm-priority-btn--active': selectedPriority === 'low' }"
                  style="--btn-color: var(--color-success, #10b981)"
                  @click="selectedPriority = 'low'"
                >
                  <span class="ecm-priority-dot" style="background: var(--color-success, #10b981)" />
                  Низкий
                </button>
                <button
                  class="ecm-priority-btn"
                  :class="{ 'ecm-priority-btn--active': selectedPriority === 'medium' }"
                  style="--btn-color: var(--color-primary, #7c5cfc)"
                  @click="selectedPriority = 'medium'"
                >
                  <span class="ecm-priority-dot" style="background: var(--color-primary, #7c5cfc)" />
                  Средний
                </button>
                <button
                  class="ecm-priority-btn"
                  :class="{ 'ecm-priority-btn--active': selectedPriority === 'high' }"
                  style="--btn-color: #f59e0b"
                  @click="selectedPriority = 'high'"
                >
                  <span class="ecm-priority-dot" style="background: #f59e0b" />
                  Высокий
                </button>
                <button
                  class="ecm-priority-btn"
                  :class="{ 'ecm-priority-btn--active': selectedPriority === 'critical' }"
                  style="--btn-color: var(--color-danger, #ef4444)"
                  @click="selectedPriority = 'critical'"
                >
                  <span class="ecm-priority-dot" style="background: var(--color-danger, #ef4444)" />
                  Крит.
                </button>
              </div>
            </div>

            <!-- Тип задачи -->
            <div class="ecm-field">
              <span class="ecm-field__label">Тип задачи</span>
              <div class="ecm-type-btns">
                <button class="ecm-type-btn" :class="{ 'ecm-type-btn--active': selectedTaskType === 'task' }" title="Задача" @click="selectedTaskType = 'task'">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
                </button>
                <button class="ecm-type-btn" :class="{ 'ecm-type-btn--active': selectedTaskType === 'bug' }" title="Баг" @click="selectedTaskType = 'bug'">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M8 2l1.88 1.88M14.12 3.88L16 2M9 7.13v-1a3.003 3.003 0 1 1 6 0v1"/>
                    <path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6"/>
                    <path d="M12 20v-9M6.53 9C4.6 8.8 3 7.1 3 5M6 13H2M6 17l-4 1M17.47 9c1.93-.2 3.53-1.9 3.53-4M18 13h4M18 17l4 1"/>
                  </svg>
                </button>
                <button class="ecm-type-btn" :class="{ 'ecm-type-btn--active': selectedTaskType === 'feature' }" title="Фича" @click="selectedTaskType = 'feature'">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                </button>
                <button class="ecm-type-btn" :class="{ 'ecm-type-btn--active': selectedTaskType === 'improvement' }" title="Улучшение" @click="selectedTaskType = 'improvement'">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/></svg>
                </button>
              </div>
            </div>

            <!-- Дедлайн -->
            <div class="ecm-field">
              <span class="ecm-field__label">Дедлайн</span>
              <input
                v-model="selectedDueDate"
                type="date"
                class="ecm-date-input"
              />
            </div>

            <!-- Метки -->
            <div class="ecm-field">
              <span class="ecm-field__label">Метки</span>
              <div class="ecm-labels-list">
                <span
                  v-for="label in cardLabels"
                  :key="label.id"
                  class="ecm-label-pill"
                  :style="{ background: label.color }"
                >
                  {{ label.name }}
                  <button class="ecm-label-pill__remove" @click="toggleLabel(label)">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </span>
              </div>
              <button class="ecm-labels-toggle" @click="showLabelPicker = !showLabelPicker">
                {{ showLabelPicker ? 'Скрыть' : '+ Добавить метку' }}
              </button>
              <div v-if="showLabelPicker" class="ecm-label-picker">
                <div
                  v-for="label in boardStore.labels"
                  :key="label.id"
                  class="ecm-label-option"
                  :class="{ 'ecm-label-option--selected': isLabelAssigned(label.id) }"
                  @click="toggleLabel(label)"
                >
                  <span class="ecm-label-option__color" :style="{ background: label.color }" />
                  <span class="ecm-label-option__name">{{ label.name }}</span>
                  <svg v-if="isLabelAssigned(label.id)" class="ecm-label-option__check" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
                </div>
                <div v-if="boardStore.labels.length === 0" class="ecm-section__empty" style="min-height: 32px;">
                  Метки не созданы
                </div>
              </div>
            </div>

            <!-- Дата создания -->
            <div class="ecm-field">
              <span class="ecm-field__label">Дата создания</span>
              <span class="ecm-field__value">{{ new Date(card.createdAt).toLocaleString('ru-RU') }}</span>
            </div>
          </div>
        </div>

        <!-- Action buttons -->
        <div class="ecm-sidebar__actions">
          <BaseButton
            :loading="loading"
            :disabled="!title.trim()"
            @click="handleUpdate"
          >
            Сохранить
          </BaseButton>
          <BaseButton
            v-if="canDelete"
            variant="danger"
            :disabled="loading"
            @click="handleDelete"
          >
            Удалить
          </BaseButton>
        </div>

        <!-- История (collapsible) -->
        <div class="ecm-block">
          <button class="ecm-block__header" @click="activityOpen = !activityOpen">
            <span class="ecm-block__title">История</span>
            <svg class="ecm-block__chevron" :class="{ 'ecm-block__chevron--open': activityOpen }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="6 9 12 15 18 9"/></svg>
          </button>

          <!-- Activity timeline (inside block) -->
          <div v-show="activityOpen" class="ecm-activity">
          <div v-if="activityLoading && !activityLoaded" class="ecm-section__center">
            <BaseSpinner size="sm" />
          </div>
          <div v-else-if="activityEntries.length === 0 && activityLoaded" class="ecm-section__empty">
            Нет записей активности
          </div>
          <div v-else class="ecm-activity__list">
            <div
              v-for="entry in activityEntries"
              :key="entry.id"
              class="ecm-activity__item"
            >
              <div class="ecm-activity__icon">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="activityIconSvg(entry.activityType)" />
              </div>
              <div class="ecm-activity__body">
                <span class="ecm-activity__description">{{ formatActivity(entry) }}</span>
                <span class="ecm-activity__meta">
                  {{ getMemberName(entry.actorId) }}
                  &middot;
                  {{ new Date(entry.createdAt).toLocaleString('ru-RU') }}
                </span>
              </div>
            </div>
          </div>
          <div v-if="activityNextCursor" class="ecm-section__load-more">
            <BaseButton
              variant="ghost"
              size="sm"
              :loading="activityLoading"
              @click="loadMoreActivity"
            >
              Загрузить ещё
            </BaseButton>
          </div>
        </div>
        </div><!-- /ecm-block История -->
      </div>
    </div>

    <ConfirmModal
      v-if="showConfirmDelete"
      title="Удалить карточку?"
      message="Вы уверены, что хотите удалить эту карточку? Это действие нельзя отменить."
      confirm-text="Удалить"
      cancel-text="Отмена"
      variant="danger"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />

    <!-- Lightbox -->
    <Teleport to="body">
      <div v-if="lightboxUrl" class="lightbox" @click="lightboxUrl = null">
        <img :src="lightboxUrl" class="lightbox__image" @click.stop />
        <button class="lightbox__close" @click="lightboxUrl = null">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>
    </Teleport>
  </BaseModal>
</template>

<style scoped>
/* ===== Header ===== */
.ecm-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ecm-header__icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--color-primary-soft);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ecm-header__title {
  font-size: var(--font-size-lg, 18px);
  font-weight: 700;
  letter-spacing: -0.02em;
  margin: 0;
}

.ecm-layout {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 32px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.ecm-main {
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  max-height: calc(86vh - 80px);
  padding-right: 16px;
}

.ecm-main::-webkit-scrollbar { width: 5px; }
.ecm-main::-webkit-scrollbar-track { background: transparent; }
.ecm-main::-webkit-scrollbar-thumb { background: var(--color-border, #d1d5db); border-radius: 3px; }

.ecm-sidebar {
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
  overflow-x: hidden;
  padding-left: 20px;
  border-left: 1px solid var(--color-border-light, #e5e7eb);
  min-height: 0;
}

.ecm-sidebar::-webkit-scrollbar { width: 5px; }
.ecm-sidebar::-webkit-scrollbar-track { background: transparent; }
.ecm-sidebar::-webkit-scrollbar-thumb { background: var(--color-border, #d1d5db); border-radius: 3px; }

/* ===== Collapsible blocks ===== */
.ecm-block {
  border: 1px solid var(--color-border-light, #e5e7eb);
  border-radius: 10px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.ecm-block:last-child {
  flex: 1;
  min-height: 0;
}

.ecm-block__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 10px 14px;
  background: var(--color-input-bg, #f9fafb);
  border: none;
  cursor: pointer;
  transition: background 0.15s;
}
.ecm-block__header:hover { background: var(--color-primary-soft, #eef2ff); }

.ecm-block__title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary, #6b7280);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.ecm-block__chevron {
  color: var(--color-text-tertiary, #9ca3af);
  transition: transform 0.2s;
}
.ecm-block__chevron--open { transform: rotate(180deg); }

.ecm-block__body {
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* ===== Sidebar fields with avatars ===== */
.ecm-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ecm-field__label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-tertiary, #9ca3af);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.ecm-field__user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
}
.ecm-field__user:hover { background: var(--color-input-bg, #f3f4f6); }

.ecm-field__name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text, #111827);
}

.ecm-field__value {
  font-size: 14px;
  color: var(--color-text, #111827);
  padding: 0 8px;
}

.ecm-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--gradient-primary, linear-gradient(135deg, #6b7c4e, var(--color-primary)));
  color: white;
  font-size: 13px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ecm-avatar--empty {
  background: var(--color-border, #d1d5db);
  color: var(--color-text-tertiary, #9ca3af);
}

/* ===== Sidebar actions (no jump!) ===== */
.ecm-sidebar__actions {
  display: flex;
  gap: 8px;
}
.ecm-sidebar__actions > * {
  flex: 1;
}

/* ===== Sidebar tab ===== */
.ecm-sidebar__tabs {
  display: flex;
  border-bottom: 1px solid var(--color-border, #e5e7eb);
  margin-top: 4px;
}

.ecm-sidebar__tab {
  padding: 8px 16px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.15s;
  margin-bottom: -1px;
}

.ecm-sidebar__tab:hover {
  color: var(--color-text, #111827);
}

.ecm-sidebar__tab--active {
  color: var(--color-primary, #6b7c4e);
  border-bottom-color: var(--color-primary, #6b7c4e);
}

/* ===== Sections ===== */
.ecm-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.ecm-section__label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-tertiary);
}

.ecm-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: var(--color-primary);
  color: white;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: none;
}

.ecm-section__center {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60px;
}

.ecm-section__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60px;
  color: var(--color-text-tertiary, #9ca3af);
  font-size: 13px;
}

.ecm-section__load-more {
  display: flex;
  justify-content: center;
  margin-top: 8px;
}

/* ===== Upload area ===== */
.ecm-upload {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ecm-upload__label {
  display: block;
  cursor: pointer;
}

.ecm-upload__input {
  display: none;
}

.ecm-upload__zone {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border: 1.5px dashed var(--color-border, #d1d5db);
  border-radius: var(--radius-md, 10px);
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-tertiary, #9ca3af);
  transition: all 0.2s;
}

.ecm-upload__zone-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: var(--color-surface-alt);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
}

.ecm-upload__zone:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.ecm-upload__zone:hover .ecm-upload__zone-icon {
  background: var(--color-primary);
  color: white;
}

.ecm-upload__progress {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ecm-upload__progress-bar {
  flex: 1;
  height: 6px;
  background: var(--color-border, #e5e7eb);
  border-radius: 3px;
  overflow: hidden;
}

.ecm-upload__progress-fill {
  height: 100%;
  background: var(--color-primary, #6b7c4e);
  border-radius: 3px;
  transition: width 0.2s;
}

.ecm-upload__progress-text {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary, #6b7280);
  min-width: 36px;
  text-align: right;
}

.ecm-upload__error {
  font-size: 13px;
  color: var(--color-danger, #ef4444);
}

/* ===== File list ===== */
.ecm-files {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ecm-file {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: var(--color-surface-alt, #f9fafb);
  border-radius: var(--radius-sm, 8px);
  transition: background 0.15s;
  cursor: pointer;
}
.ecm-file:hover { background: var(--color-primary-soft, #eef2ff); }

.ecm-file__preview {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  flex-shrink: 0;
  border: 1px solid var(--color-border-light, #e5e7eb);
}

.ecm-file__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ecm-file__icon {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-input-bg, #f3f4f6);
  border-radius: 8px;
  cursor: pointer;
  flex-shrink: 0;
  color: var(--color-text-tertiary, #9ca3af);
}
.ecm-file__icon:hover { color: var(--color-primary, #6b7c4e); }

.ecm-file__body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.ecm-file__name {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-primary, #6b7c4e);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}
.ecm-file__name:hover { text-decoration: underline; }

.ecm-file__meta {
  font-size: 12px;
  color: var(--color-text-tertiary, #9ca3af);
}

.ecm-file__actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.ecm-file__btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  background: none;
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 6px;
  color: var(--color-text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.15s;
}

.ecm-file__btn:hover {
  background: var(--color-primary-soft, rgba(99, 102, 241, 0.1));
  border-color: var(--color-primary, #6b7c4e);
  color: var(--color-primary, #6b7c4e);
}

.ecm-file__btn--danger:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: var(--color-danger, #ef4444);
  color: var(--color-danger, #ef4444);
}

.ecm-file__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ===== Comment form ===== */
.ecm-comment-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ecm-comment-form__textarea {
  width: 100%;
  padding: 10px 14px;
  border: 1.5px solid var(--color-input-border, #d1d5db);
  border-radius: var(--radius-md, 10px);
  background: var(--color-input-bg, #f9fafb);
  color: var(--color-text, #111827);
  font-family: inherit;
  font-size: 14px;
  line-height: 1.5;
  resize: vertical;
  min-height: 60px;
  outline: none;
  transition: all 0.15s;
  box-sizing: border-box;
}

.ecm-comment-form__textarea:focus {
  border-color: var(--color-input-focus, #6b7c4e);
  background: var(--color-surface, #fff);
  box-shadow: var(--shadow-focus, 0 0 0 3px rgba(99, 102, 241, 0.15));
}

.ecm-comment-form__reply {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: var(--color-surface-alt, #f3f4f6);
  border-radius: 6px;
  font-size: 13px;
  color: var(--color-text-secondary, #6b7280);
}

.ecm-comment-form__reply-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ecm-comment-form__reply-cancel {
  background: none;
  border: none;
  font-size: 18px;
  color: var(--color-text-tertiary, #9ca3af);
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}

.ecm-comment-form__reply-cancel:hover {
  color: var(--color-text, #111827);
}

/* ===== Comments list ===== */
.ecm-comments {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.ecm-comment {
  padding: 10px 12px;
  background: var(--color-surface-alt, #f9fafb);
  border-radius: var(--radius-sm, 8px);
}

.ecm-comment--reply {
  margin-left: 24px;
  border-left: 2px solid var(--color-border, #e5e7eb);
}

.ecm-comment__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.ecm-comment__author {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text, #111827);
}

.ecm-comment__time {
  font-size: 12px;
  color: var(--color-text-tertiary, #9ca3af);
}

.ecm-comment__content {
  font-size: 14px;
  color: var(--color-text-secondary, #374151);
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.ecm-comment__actions {
  display: flex;
  gap: 12px;
  margin-top: 6px;
}

.ecm-comment__action {
  background: none;
  border: none;
  font-size: 12px;
  color: var(--color-text-tertiary, #9ca3af);
  cursor: pointer;
  padding: 0;
  transition: color 0.15s;
}

.ecm-comment__action:hover {
  color: var(--color-primary, #6b7c4e);
}

.ecm-comment__action--danger:hover {
  color: var(--color-danger, #ef4444);
}

.ecm-comment__action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ecm-comment__edit-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

/* ===== Activity ===== */
.ecm-activity {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 14px;
  overflow-y: auto;
  flex: 1;
}

.ecm-activity__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.ecm-activity__item {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.ecm-activity__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--color-input-bg, #f3f4f6);
  border: 1.5px solid var(--color-border, #d1d5db);
  border-radius: 50%;
  color: var(--color-text-tertiary, #9ca3af);
  flex-shrink: 0;
}

.ecm-activity__body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.ecm-activity__description {
  font-size: 13px;
  color: var(--color-text, #111827);
  line-height: 1.4;
}

.ecm-activity__meta {
  font-size: 11px;
  color: var(--color-text-tertiary, #9ca3af);
}

/* ===== Lightbox ===== */
.lightbox {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 20000;
  cursor: pointer;
  padding: 40px;
}

.lightbox__image {
  max-width: 90vw;
  max-height: 90vh;
  object-fit: contain;
  border-radius: 8px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  cursor: default;
}

.lightbox__close {
  position: absolute;
  top: 16px;
  right: 16px;
  background: rgba(255, 255, 255, 0.15);
  border: none;
  color: white;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s;
}
.lightbox__close:hover { background: rgba(255, 255, 255, 0.3); }

/* ===== Priority & Type buttons ===== */
.ecm-priority-btns {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.ecm-priority-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border: 1.5px solid var(--color-border, #e5e7eb);
  border-radius: 6px;
  background: var(--color-surface, #fff);
  color: var(--color-text-secondary, #6b7280);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.ecm-priority-btn:hover {
  border-color: var(--btn-color);
}

.ecm-priority-btn--active {
  border-color: var(--btn-color);
  background: color-mix(in srgb, var(--btn-color) 8%, transparent);
  color: var(--color-text, #111827);
}

.ecm-priority-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.ecm-type-btns {
  display: flex;
  gap: 4px;
}

.ecm-type-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1.5px solid var(--color-border, #e5e7eb);
  border-radius: 6px;
  background: var(--color-surface, #fff);
  color: var(--color-text-tertiary, #9ca3af);
  cursor: pointer;
  transition: all 0.15s;
}

.ecm-type-btn:hover {
  border-color: var(--color-primary, #7c5cfc);
  color: var(--color-primary, #7c5cfc);
}

.ecm-type-btn--active {
  border-color: var(--color-primary, #7c5cfc);
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
  color: var(--color-primary, #7c5cfc);
}

.ecm-date-input {
  padding: 6px 10px;
  border: 1.5px solid var(--color-input-border, #e5e7eb);
  border-radius: 6px;
  background: var(--color-input-bg, #f9fafb);
  color: var(--color-text, #111827);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.15s;
  width: 100%;
  box-sizing: border-box;
}

.ecm-date-input:focus {
  border-color: var(--color-input-focus, #7c5cfc);
}

/* ===== Labels ===== */
.ecm-labels-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.ecm-label-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 10px;
  color: white;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.6;
}

.ecm-label-pill__remove {
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  padding: 0;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  transition: color 0.15s;
}

.ecm-label-pill__remove:hover {
  color: white;
}

.ecm-labels-toggle {
  background: none;
  border: none;
  color: var(--color-primary, #7c5cfc);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  padding: 2px 0;
  transition: color 0.15s;
}

.ecm-labels-toggle:hover {
  text-decoration: underline;
}

.ecm-label-picker {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 6px;
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 8px;
  background: var(--color-surface, #fff);
  max-height: 160px;
  overflow-y: auto;
}

.ecm-label-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.1s;
}

.ecm-label-option:hover {
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
}

.ecm-label-option--selected {
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
}

.ecm-label-option__color {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  flex-shrink: 0;
}

.ecm-label-option__name {
  flex: 1;
  font-size: 13px;
  color: var(--color-text, #111827);
}

.ecm-label-option__check {
  color: var(--color-primary, #7c5cfc);
  flex-shrink: 0;
}

/* ===== Checklists ===== */
.ecm-checklist {
  padding: 12px;
  border: 1px solid var(--color-border-light, #f3f4f6);
  border-radius: 8px;
  margin-bottom: 10px;
  background: var(--color-surface-alt, #fafafa);
}

.ecm-checklist__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.ecm-checklist__title {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text, #111827);
}

.ecm-checklist__progress-text {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-tertiary, #9ca3af);
}

.ecm-checklist__remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: none;
  border: none;
  color: var(--color-text-tertiary, #9ca3af);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s;
}

.ecm-checklist__remove:hover {
  color: var(--color-danger, #ef4444);
  background: var(--color-danger-soft, rgba(239, 68, 68, 0.08));
}

.ecm-checklist__bar {
  height: 4px;
  background: var(--color-border, #e5e7eb);
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 8px;
}

.ecm-checklist__bar-fill {
  height: 100%;
  background: var(--color-success, #10b981);
  border-radius: 2px;
  transition: width 0.2s;
}

.ecm-checklist__items {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ecm-checklist-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.ecm-checklist-item__check {
  width: 16px;
  height: 16px;
  accent-color: var(--color-primary, #7c5cfc);
  cursor: pointer;
  flex-shrink: 0;
}

.ecm-checklist-item__title {
  flex: 1;
  font-size: 13px;
  color: var(--color-text, #111827);
}

.ecm-checklist-item__title--done {
  text-decoration: line-through;
  color: var(--color-text-tertiary, #9ca3af);
}

.ecm-checklist-item__remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  background: none;
  border: none;
  color: var(--color-text-tertiary, #9ca3af);
  cursor: pointer;
  border-radius: 3px;
  opacity: 0;
  transition: all 0.15s;
}

.ecm-checklist-item:hover .ecm-checklist-item__remove {
  opacity: 1;
}

.ecm-checklist-item__remove:hover {
  color: var(--color-danger, #ef4444);
}

.ecm-checklist__add-item {
  display: flex;
  gap: 6px;
  margin-top: 6px;
}

.ecm-checklist__add-input {
  flex: 1;
  padding: 5px 8px;
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 5px;
  background: var(--color-surface, #fff);
  color: var(--color-text, #111827);
  font-size: 12px;
  font-family: inherit;
  outline: none;
}

.ecm-checklist__add-input:focus {
  border-color: var(--color-input-focus, #7c5cfc);
}

.ecm-checklist__add-input::placeholder {
  color: var(--color-text-tertiary, #9ca3af);
}

.ecm-checklist__add-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
  border: none;
  border-radius: 5px;
  color: var(--color-primary, #7c5cfc);
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s;
}

.ecm-checklist__add-btn:hover:not(:disabled) {
  background: var(--color-primary, #7c5cfc);
  color: white;
}

.ecm-checklist__add-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.ecm-checklist-add {
  display: flex;
  gap: 8px;
  align-items: center;
}

.ecm-checklist-add__input {
  flex: 1;
  padding: 7px 10px;
  border: 1.5px dashed var(--color-border, #e5e7eb);
  border-radius: 6px;
  background: var(--color-surface, #fff);
  color: var(--color-text, #111827);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.15s;
}

.ecm-checklist-add__input:focus {
  border-color: var(--color-primary, #7c5cfc);
}

.ecm-checklist-add__input::placeholder {
  color: var(--color-text-tertiary, #9ca3af);
}

/* ===== Subtasks ===== */
.ecm-subtask-parents {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.ecm-subtask-parents__label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-tertiary, #9ca3af);
}

.ecm-subtask-parent {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-primary, #7c5cfc);
  padding: 2px 8px;
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
  border-radius: 4px;
}

.ecm-subtask-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ecm-subtask-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  background: var(--color-surface-alt, #fafafa);
  border-radius: 6px;
  border: 1px solid var(--color-border-light, #f3f4f6);
  transition: background 0.15s;
}

.ecm-subtask-item:hover {
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
}

.ecm-subtask-item__body {
  flex: 1;
  min-width: 0;
}

.ecm-subtask-item__title {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text, #111827);
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ecm-subtask-item__column {
  font-size: 11px;
  color: var(--color-text-tertiary, #9ca3af);
}

.ecm-subtask-item__unlink {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  background: none;
  border: none;
  color: var(--color-text-tertiary, #9ca3af);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s;
  flex-shrink: 0;
}

.ecm-subtask-item__unlink:hover {
  color: var(--color-danger, #ef4444);
  background: var(--color-danger-soft, rgba(239, 68, 68, 0.08));
}

.ecm-subtask-add-btn {
  background: none;
  border: 1.5px dashed var(--color-border, #e5e7eb);
  border-radius: 6px;
  padding: 8px 12px;
  color: var(--color-text-tertiary, #9ca3af);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  width: 100%;
  text-align: left;
  transition: all 0.15s;
  margin-top: 8px;
}

.ecm-subtask-add-btn:hover {
  border-color: var(--color-primary, #7c5cfc);
  color: var(--color-primary, #7c5cfc);
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
}

.ecm-subtask-search {
  position: relative;
  margin-top: 8px;
}

.ecm-subtask-search__input {
  width: 100%;
  padding: 8px 32px 8px 10px;
  border: 1.5px solid var(--color-input-border, #e5e7eb);
  border-radius: 6px;
  background: var(--color-input-bg, #f9fafb);
  color: var(--color-text, #111827);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
}

.ecm-subtask-search__input:focus {
  border-color: var(--color-input-focus, #7c5cfc);
}

.ecm-subtask-search__close {
  position: absolute;
  top: 8px;
  right: 8px;
  background: none;
  border: none;
  color: var(--color-text-tertiary, #9ca3af);
  cursor: pointer;
  padding: 2px;
}

.ecm-subtask-search__results {
  margin-top: 4px;
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 6px;
  background: var(--color-surface, #fff);
  max-height: 200px;
  overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.ecm-subtask-search__result {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  cursor: pointer;
  transition: background 0.1s;
}

.ecm-subtask-search__result:hover {
  background: var(--color-primary-soft, rgba(124, 92, 252, 0.08));
}

.ecm-subtask-search__result-title {
  font-size: 13px;
  color: var(--color-text, #111827);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ecm-subtask-search__result-col {
  font-size: 11px;
  color: var(--color-text-tertiary, #9ca3af);
  flex-shrink: 0;
  margin-left: 8px;
}

.ecm-subtask-search__empty {
  padding: 10px;
  text-align: center;
  font-size: 13px;
  color: var(--color-text-tertiary, #9ca3af);
}
</style>
