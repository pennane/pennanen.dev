'use client'

import { Extension, type Editor as Tiptap, type Range } from '@tiptap/core'
import FileHandler from '@tiptap/extension-file-handler'
import Image from '@tiptap/extension-image'
import { Markdown } from '@tiptap/markdown'
import { EditorContent, ReactRenderer, useEditor } from '@tiptap/react'
import { BubbleMenu } from '@tiptap/react/menus'
import StarterKit from '@tiptap/starter-kit'
import Suggestion, { type SuggestionKeyDownProps, type SuggestionProps } from '@tiptap/suggestion'
import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { readImageWidth, withImageWidth } from '../../../lib/log/image-width'
import type { ItemLink } from '../../../lib/log/types'
import styles from './editor.module.css'

const SAVE_AFTER_MS = 800
const QUERY_LENGTH_TO_SEARCH = 2
const IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']

export type ProseEditorProps = {
  initial: string
  onSave: (markdown: string) => void
  onImage: (file: File, insert: (src: string) => void) => void
  findLinks: (query: string) => Promise<ItemLink[]>
}

type PickerHandle = { onKeyDown: (event: KeyboardEvent) => boolean }

const LinkPicker = forwardRef<PickerHandle, SuggestionProps<ItemLink>>(function LinkPicker(
  { items, command },
  ref
) {
  const [pick, setPick] = useState({ items, index: 0 })
  const chosen = pick.items === items ? pick.index : 0
  const choose = (index: number) => setPick({ items, index: (index + items.length) % items.length })
  useImperativeHandle(ref, () => ({
    onKeyDown: (event) => {
      if (items.length === 0) return false
      if (event.key === 'ArrowDown') choose(chosen + 1)
      else if (event.key === 'ArrowUp') choose(chosen - 1)
      else if (event.key === 'Enter') command(items[chosen])
      else return false
      return true
    }
  }))
  if (items.length === 0) return null
  return (
    <ul className={styles.picker}>
      {items.map((item, i) => (
        <li key={item.href}>
          <button
            type="button"
            aria-pressed={i === chosen}
            onMouseDown={(event) => event.preventDefault()}
            onClick={() => command(item)}
          >
            {item.title}
          </button>
        </li>
      ))}
    </ul>
  )
})

const SMALLEST_IMAGE_PX = 80

const titlePart = (title: unknown) => (typeof title === 'string' && title ? ` "${title}"` : '')

const ResizableImage = Image.extend({
  parseMarkdown: (token, helpers) => {
    const { src, width } = readImageWidth(token.href ?? '')
    return helpers.createNode('image', { src, width, title: token.title, alt: token.text })
  },
  renderMarkdown: (node) => {
    const href = withImageWidth({
      src: node.attrs?.src ?? '',
      width: node.attrs?.width ?? undefined
    })
    return `![${node.attrs?.alt ?? ''}](${href}${titlePart(node.attrs?.title)})`
  }
}).configure({
  resize: {
    enabled: true,
    directions: ['left', 'right'],
    minWidth: SMALLEST_IMAGE_PX,
    minHeight: SMALLEST_IMAGE_PX,
    alwaysPreserveAspectRatio: true
  }
})

const placeBelow = (element: Element, rect: DOMRect | null | undefined) => {
  if (!(element instanceof HTMLElement) || !rect) return
  element.style.position = 'fixed'
  element.style.left = `${rect.left}px`
  element.style.top = `${rect.bottom + 4}px`
}

const linkTo = (editor: Tiptap, range: Range, link: ItemLink) =>
  editor
    .chain()
    .focus()
    .insertContentAt(range, [
      { type: 'text', text: link.title, marks: [{ type: 'link', attrs: { href: link.href } }] },
      { type: 'text', text: ' ' }
    ])
    .run()

const itemLinks = (findLinks: (query: string) => Promise<ItemLink[]>) =>
  Extension.create({
    name: 'itemLinks',
    addProseMirrorPlugins() {
      return [
        Suggestion<ItemLink>({
          editor: this.editor,
          char: '@',
          items: ({ query }) => (query.length < QUERY_LENGTH_TO_SEARCH ? [] : findLinks(query)),
          command: ({ editor, range, props }) => linkTo(editor, range, props),
          render: () => {
            let picker: ReactRenderer<PickerHandle, SuggestionProps<ItemLink>> | undefined
            return {
              onStart: (props) => {
                picker = new ReactRenderer(LinkPicker, { props, editor: props.editor })
                document.body.appendChild(picker.element)
                placeBelow(picker.element, props.clientRect?.())
              },
              onUpdate: (props) => {
                picker?.updateProps(props)
                if (picker) placeBelow(picker.element, props.clientRect?.())
              },
              onKeyDown: ({ event }: SuggestionKeyDownProps) =>
                picker?.ref?.onKeyDown(event) ?? false,
              onExit: () => {
                picker?.element.remove()
                picker?.destroy()
              }
            }
          }
        })
      ]
    }
  })

export const ProseEditor = ({ initial, onSave, onImage, findLinks }: ProseEditorProps) => {
  const pending = useRef<ReturnType<typeof setTimeout>>(undefined)
  const markdown = useRef(initial)
  const latest = useRef({ onSave, onImage })
  useEffect(() => {
    latest.current = { onSave, onImage }
  })

  const insertImages = (editor: Tiptap, files: File[], at: number) =>
    files.forEach((file) =>
      latest.current.onImage(file, (src) =>
        editor.chain().focus().insertContentAt(at, { type: 'image', attrs: { src } }).run()
      )
    )

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit.configure({ link: { openOnClick: false, autolink: true } }),
      ResizableImage,
      Markdown,
      FileHandler.configure({
        allowedMimeTypes: IMAGE_TYPES,
        onPaste: (editor, files) => insertImages(editor, files, editor.state.selection.anchor),
        onDrop: (editor, files, at) => insertImages(editor, files, at)
      }),
      itemLinks(findLinks)
    ],
    content: initial,
    contentType: 'markdown',
    editorProps: { attributes: { class: styles.prose } },
    onUpdate: ({ editor }) => {
      markdown.current = editor.getMarkdown()
      clearTimeout(pending.current)
      pending.current = setTimeout(() => {
        pending.current = undefined
        latest.current.onSave(markdown.current)
      }, SAVE_AFTER_MS)
    }
  })

  useEffect(
    () => () => {
      if (pending.current === undefined) return
      clearTimeout(pending.current)
      latest.current.onSave(markdown.current)
    },
    []
  )

  if (!editor) return null
  return (
    <>
      <BubbleMenu editor={editor} className={styles.bubble}>
        <button
          type="button"
          aria-pressed={editor.isActive('bold')}
          onClick={() => editor.chain().focus().toggleBold().run()}
        >
          <b>b</b>
        </button>
        <button
          type="button"
          aria-pressed={editor.isActive('italic')}
          onClick={() => editor.chain().focus().toggleItalic().run()}
        >
          <i>i</i>
        </button>
        <button
          type="button"
          aria-pressed={editor.isActive('heading', { level: 2 })}
          onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
        >
          heading
        </button>
        <button
          type="button"
          aria-pressed={editor.isActive('link')}
          onClick={() => {
            if (editor.isActive('link')) {
              editor.chain().focus().unsetLink().run()
              return
            }
            const href = window.prompt('link to')
            if (href) editor.chain().focus().setLink({ href }).run()
          }}
        >
          link
        </button>
      </BubbleMenu>
      <EditorContent editor={editor} />
    </>
  )
}
