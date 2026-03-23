<script setup lang="ts">
import { watch, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import Placeholder from '@tiptap/extension-placeholder'
import Underline from '@tiptap/extension-underline'

const props = defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
  label?: string
  grow?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const editor = useEditor({
  content: props.modelValue || '',
  editable: !props.disabled,
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2, 3] },
    }),
    Underline,
    Link.configure({ openOnClick: false, HTMLAttributes: { class: 'rte-link' } }),
    Image.configure({ inline: true }),
    Placeholder.configure({ placeholder: props.placeholder || '' }),
  ],
  onUpdate({ editor }) {
    const html = editor.getHTML()
    emit('update:modelValue', html === '<p></p>' ? '' : html)
  },
})

watch(() => props.modelValue, (val) => {
  if (editor.value && editor.value.getHTML() !== val) {
    editor.value.commands.setContent(val || '', false)
  }
})

watch(() => props.disabled, (val) => {
  editor.value?.setEditable(!val)
})

onBeforeUnmount(() => editor.value?.destroy())

function toggleBold() { editor.value?.chain().focus().toggleBold().run() }
function toggleItalic() { editor.value?.chain().focus().toggleItalic().run() }
function toggleUnderline() { editor.value?.chain().focus().toggleUnderline().run() }
function toggleStrike() { editor.value?.chain().focus().toggleStrike().run() }
function toggleH1() { editor.value?.chain().focus().toggleHeading({ level: 1 }).run() }
function toggleH2() { editor.value?.chain().focus().toggleHeading({ level: 2 }).run() }
function toggleH3() { editor.value?.chain().focus().toggleHeading({ level: 3 }).run() }
function toggleBulletList() { editor.value?.chain().focus().toggleBulletList().run() }
function toggleOrderedList() { editor.value?.chain().focus().toggleOrderedList().run() }
function toggleCodeBlock() { editor.value?.chain().focus().toggleCodeBlock().run() }
function toggleBlockquote() { editor.value?.chain().focus().toggleBlockquote().run() }
function setHorizontalRule() { editor.value?.chain().focus().setHorizontalRule().run() }

function setLink() {
  const url = window.prompt('URL ссылки:')
  if (url) {
    editor.value?.chain().focus().setLink({ href: url }).run()
  }
}

function addImage() {
  const url = window.prompt('URL изображения:')
  if (url) {
    editor.value?.chain().focus().setImage({ src: url }).run()
  }
}

function isActive(name: string, attrs?: Record<string, unknown>): boolean {
  return editor.value?.isActive(name, attrs) ?? false
}
</script>

<template>
  <div class="rte" :class="{ 'rte--disabled': disabled, 'rte--grow': grow }">
    <label v-if="label" class="rte__label">{{ label }}</label>

    <div v-if="editor && !disabled" class="rte__toolbar">
      <button type="button" :class="{ active: isActive('bold') }" title="Жирный" @click="toggleBold">
        <strong>B</strong>
      </button>
      <button type="button" :class="{ active: isActive('italic') }" title="Курсив" @click="toggleItalic">
        <em>I</em>
      </button>
      <button type="button" :class="{ active: isActive('underline') }" title="Подчёркнутый" @click="toggleUnderline">
        <u>U</u>
      </button>
      <button type="button" :class="{ active: isActive('strike') }" title="Зачёркнутый" @click="toggleStrike">
        <s>S</s>
      </button>

      <span class="rte__sep" />

      <button type="button" :class="{ active: isActive('heading', { level: 1 }) }" title="Заголовок 1" @click="toggleH1">H1</button>
      <button type="button" :class="{ active: isActive('heading', { level: 2 }) }" title="Заголовок 2" @click="toggleH2">H2</button>
      <button type="button" :class="{ active: isActive('heading', { level: 3 }) }" title="Заголовок 3" @click="toggleH3">H3</button>

      <span class="rte__sep" />

      <button type="button" :class="{ active: isActive('bulletList') }" title="Список" @click="toggleBulletList">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><circle cx="4" cy="6" r="1" fill="currentColor"/><circle cx="4" cy="12" r="1" fill="currentColor"/><circle cx="4" cy="18" r="1" fill="currentColor"/></svg>
      </button>
      <button type="button" :class="{ active: isActive('orderedList') }" title="Нумерованный" @click="toggleOrderedList">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="10" y1="6" x2="21" y2="6"/><line x1="10" y1="12" x2="21" y2="12"/><line x1="10" y1="18" x2="21" y2="18"/><text x="2" y="8" font-size="8" fill="currentColor" stroke="none">1</text><text x="2" y="14" font-size="8" fill="currentColor" stroke="none">2</text><text x="2" y="20" font-size="8" fill="currentColor" stroke="none">3</text></svg>
      </button>

      <span class="rte__sep" />

      <button type="button" :class="{ active: isActive('codeBlock') }" title="Код" @click="toggleCodeBlock">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
      </button>
      <button type="button" :class="{ active: isActive('blockquote') }" title="Цитата" @click="toggleBlockquote">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 21c3 0 7-1 7-8V5c0-1.25-.756-2.017-2-2H4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2 1 0 1 0 1 1v1c0 1-1 2-2 2s-1 .008-1 1.031V21z"/><path d="M15 21c3 0 7-1 7-8V5c0-1.25-.757-2.017-2-2h-4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2h.75c0 2.25.25 4-2.75 4v3z"/></svg>
      </button>
      <button type="button" title="Линия" @click="setHorizontalRule">—</button>

      <span class="rte__sep" />

      <button type="button" :class="{ active: isActive('link') }" title="Ссылка" @click="setLink">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
      </button>
      <button type="button" title="Картинка" @click="addImage">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
      </button>
    </div>

    <EditorContent v-if="editor" :editor="editor" class="rte__content" />
  </div>
</template>

<style scoped>
.rte__label {
  display: block;
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-secondary);
  letter-spacing: 0.01em;
  margin-bottom: 6px;
}

.rte__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  padding: 6px 8px;
  border: 1.5px solid var(--color-input-border);
  border-bottom: none;
  border-radius: var(--radius-md) var(--radius-md) 0 0;
  background: var(--color-input-bg);
}

.rte__toolbar button {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: 4px 7px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  transition: all 0.15s;
}

.rte__toolbar button:hover {
  background: var(--color-primary-soft);
  color: var(--color-text);
}

.rte__toolbar button.active {
  background: var(--color-primary);
  color: white;
}

.rte__sep {
  width: 1px;
  height: 20px;
  background: var(--color-border-light);
  align-self: center;
  margin: 0 4px;
}

.rte__content {
  border: 1.5px solid var(--color-input-border);
  border-radius: 0 0 var(--radius-md) var(--radius-md);
  background: var(--color-input-bg);
  min-height: 120px;
  max-height: 400px;
  overflow-y: auto;
  transition: border-color var(--transition-fast);
}

.rte--grow { flex: 1; display: flex; flex-direction: column; }
.rte--grow .rte__content { flex: 1; max-height: none; }

.rte__content:focus-within {
  border-color: var(--color-input-focus);
  box-shadow: var(--shadow-focus);
}

.rte--disabled .rte__content {
  opacity: 0.6;
  pointer-events: none;
}

/* toolbar present → no top radius on content */
.rte__toolbar + .rte__content {
  border-top: none;
}
</style>

<style>
/* Tiptap content — global (not scoped) to style ProseMirror internals */
.rte__content .tiptap {
  padding: 12px 14px;
  outline: none;
  color: var(--color-text);
  font-size: 14px;
  line-height: 1.6;
}

.rte__content .tiptap p { margin: 0 0 8px; }
.rte__content .tiptap p:last-child { margin-bottom: 0; }

.rte__content .tiptap h1 { font-size: 1.5em; font-weight: 700; margin: 16px 0 8px; }
.rte__content .tiptap h2 { font-size: 1.25em; font-weight: 700; margin: 12px 0 6px; }
.rte__content .tiptap h3 { font-size: 1.1em; font-weight: 600; margin: 10px 0 4px; }

.rte__content .tiptap ul,
.rte__content .tiptap ol { padding-left: 24px; margin: 8px 0; }
.rte__content .tiptap li { margin: 2px 0; }

.rte__content .tiptap blockquote {
  border-left: 3px solid var(--color-primary);
  padding-left: 12px;
  margin: 8px 0;
  color: var(--color-text-secondary);
  font-style: italic;
}

.rte__content .tiptap pre {
  background: var(--color-surface-alt, #f3f4f6);
  border-radius: 6px;
  padding: 12px;
  font-family: monospace;
  font-size: 13px;
  overflow-x: auto;
  margin: 8px 0;
}

.rte__content .tiptap code {
  background: var(--color-surface-alt, #f3f4f6);
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 0.9em;
}

.rte__content .tiptap pre code {
  background: none;
  padding: 0;
}

.rte__content .tiptap hr {
  border: none;
  border-top: 1px solid var(--color-border-light);
  margin: 16px 0;
}

.rte__content .tiptap img {
  max-width: 100%;
  border-radius: 6px;
  margin: 8px 0;
}

.rte__content .tiptap a,
.rte__content .tiptap .rte-link {
  color: var(--color-primary);
  text-decoration: underline;
  cursor: pointer;
}

.rte__content .tiptap p.is-editor-empty:first-child::before {
  content: attr(data-placeholder);
  color: var(--color-text-tertiary);
  pointer-events: none;
  float: left;
  height: 0;
}
</style>
