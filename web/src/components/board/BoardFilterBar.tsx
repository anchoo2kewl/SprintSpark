import { useEffect, useRef, useState } from 'react'
import type { Task, Sprint, Tag } from '../../lib/api'

// ── Board filter bar (GitHub-style) ──────────────────────────────────────────
type BoardFilterKey = 'sprint' | 'assignee' | 'priority' | 'tag' | 'task_id'

export interface BoardFilterState {
  sprintId: number | null
  assigneeId: number | null
  priority: string
  tagId: number | null
  taskIds: number[]
}

export const EMPTY_BOARD_FILTERS: BoardFilterState = {
  sprintId: null,
  assigneeId: null,
  priority: '',
  tagId: null,
  taskIds: [],
}

export function applyBoardFilters(tasks: Task[], f: BoardFilterState): Task[] {
  return tasks.filter(t => {
    if (f.sprintId   !== null && t.sprint_id !== f.sprintId) return false
    if (f.assigneeId !== null) {
      const inMulti = t.assignees?.some(a => a.user_id === f.assigneeId)
      const inLegacy = t.assignee_id === f.assigneeId
      if (!inMulti && !inLegacy) return false
    }
    if (f.priority && t.priority !== f.priority) return false
    if (f.tagId !== null && !t.tags?.some(tag => tag.id === f.tagId)) return false
    if (f.taskIds.length > 0 && !f.taskIds.includes(t.task_number ?? 0)) return false
    return true
  })
}

interface BoardFilterBarProps {
  sprints: Sprint[]
  assignees: { id: number; name: string }[]
  tags: Tag[]
  sprintId: number | null
  assigneeId: number | null
  priority: string
  tagId: number | null
  taskIds: number[]
  onChange: (patch: {
    sprintId?: number | null
    assigneeId?: number | null
    priority?: string
    tagId?: number | null
    taskIds?: number[]
  }) => void
}

const PRIORITY_OPTIONS = [
  { value: 'urgent', label: 'Urgent', color: '#ef4444' },
  { value: 'high',   label: 'High',   color: '#f97316' },
  { value: 'medium', label: 'Medium', color: '#eab308' },
  { value: 'low',    label: 'Low',    color: '#6b7280' },
]

export default function BoardFilterBar({ sprints, assignees, tags, sprintId, assigneeId, priority, tagId, taskIds, onChange }: BoardFilterBarProps) {
  const [open, setOpen] = useState(false)
  const [category, setCategory] = useState<BoardFilterKey | null>(null)
  const [search, setSearch] = useState('')
  const [taskIdInput, setTaskIdInput] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const h = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false); setCategory(null); setSearch(''); setTaskIdInput('')
      }
    }
    document.addEventListener('mousedown', h)
    return () => document.removeEventListener('mousedown', h)
  }, [])

  const CATEGORIES: { id: BoardFilterKey; label: string; icon: React.ReactNode }[] = [
    { id: 'sprint',   label: 'Sprint',   icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg> },
    { id: 'assignee', label: 'Assignee', icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg> },
    { id: 'priority', label: 'Priority', icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 4h13M3 8h9m-9 4h6m4 0l4-4m0 0l4 4m-4-4v12" /></svg> },
    { id: 'tag',      label: 'Tag',      icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z" /></svg> },
    { id: 'task_id',  label: 'Task ID',  icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" /></svg> },
  ]

  // Active chips
  const chips: { key: BoardFilterKey; label: string }[] = []
  if (sprintId)         chips.push({ key: 'sprint',   label: `sprint:"${sprints.find(s => s.id === sprintId)?.name ?? sprintId}"` })
  if (assigneeId)       chips.push({ key: 'assignee', label: `assignee:"${assignees.find(a => a.id === assigneeId)?.name ?? assigneeId}"` })
  if (priority)         chips.push({ key: 'priority', label: `priority:${priority}` })
  if (tagId)            chips.push({ key: 'tag',      label: `tag:"${tags.find(t => t.id === tagId)?.name ?? tagId}"` })
  if (taskIds.length)   chips.push({ key: 'task_id',  label: `id:${taskIds.join(',')}` })

  const removeChip = (key: BoardFilterKey) => {
    if (key === 'sprint')   onChange({ sprintId: null })
    if (key === 'assignee') onChange({ assigneeId: null })
    if (key === 'priority') onChange({ priority: '' })
    if (key === 'tag')      onChange({ tagId: null })
    if (key === 'task_id')  onChange({ taskIds: [] })
  }

  const q = search.toLowerCase()

  let options: { value: string; label: string; sub?: string; color?: string }[] = []
  if (category === 'sprint') {
    options = sprints
      .filter(s => !q || s.name.toLowerCase().includes(q))
      .map(s => ({ value: String(s.id), label: s.name, sub: s.status === 'completed' ? 'completed' : s.status === 'active' ? 'active' : undefined }))
  } else if (category === 'assignee') {
    options = [
      { value: 'none', label: 'No assignee' },
      ...assignees
        .filter(a => !q || a.name.toLowerCase().includes(q))
        .map(a => ({ value: String(a.id), label: a.name }))
    ]
  } else if (category === 'priority') {
    options = PRIORITY_OPTIONS.map(p => ({ value: p.value, label: p.label, color: p.color }))
  } else if (category === 'tag') {
    options = tags
      .filter(t => !q || t.name.toLowerCase().includes(q))
      .map(t => ({ value: String(t.id), label: t.name, color: t.color }))
  }

  const activeVal = (key: BoardFilterKey) => {
    if (key === 'sprint')   return sprintId ? sprints.find(s => s.id === sprintId)?.name : undefined
    if (key === 'assignee') return assigneeId ? assignees.find(a => a.id === assigneeId)?.name : undefined
    if (key === 'priority') return priority || undefined
    if (key === 'tag')      return tagId ? tags.find(t => t.id === tagId)?.name : undefined
    if (key === 'task_id')  return taskIds.length ? `${taskIds.length} selected` : undefined
  }

  const selectOption = (cat: BoardFilterKey, value: string) => {
    if (cat === 'sprint')   onChange({ sprintId: value && value !== 'none' ? Number(value) : null })
    if (cat === 'assignee') onChange({ assigneeId: value && value !== 'none' ? Number(value) : null })
    if (cat === 'priority') onChange({ priority: value })
    if (cat === 'tag')      onChange({ tagId: value ? Number(value) : null })
    setOpen(false); setCategory(null); setSearch('')
  }

  const addTaskId = () => {
    const n = parseInt(taskIdInput.trim(), 10)
    if (!isNaN(n) && n > 0 && !taskIds.includes(n)) {
      onChange({ taskIds: [...taskIds, n] })
    }
    setTaskIdInput('')
  }

  const removeTaskId = (id: number) => {
    onChange({ taskIds: taskIds.filter(x => x !== id) })
  }

  const hasFilters = chips.length > 0

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => { setOpen(v => !v); setCategory(null); setSearch(''); setTaskIdInput('') }}
        className={`flex items-center gap-2 px-3 py-1.5 rounded-lg border text-sm transition-colors focus:outline-none ${
          hasFilters
            ? 'bg-primary-500/10 border-primary-500/40 text-primary-400'
            : 'bg-transparent border-dark-border-medium text-dark-text-tertiary hover:border-dark-border-strong hover:text-dark-text-secondary'
        }`}
      >
        <svg className="w-3.5 h-3.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L13 13.414V19a1 1 0 01-.553.894l-4 2A1 1 0 017 21v-7.586L3.293 6.707A1 1 0 013 6V4z" />
        </svg>
        {hasFilters ? (
          <div className="flex items-center gap-1.5 flex-wrap">
            {chips.map(c => (
              <span key={c.key} className="inline-flex items-center gap-1 font-mono text-xs">
                {c.label}
                <span
                  role="button"
                  onClick={e => { e.stopPropagation(); removeChip(c.key) }}
                  className="text-dark-text-tertiary hover:text-dark-text-primary cursor-pointer"
                >×</span>
              </span>
            ))}
          </div>
        ) : (
          <span>Filter</span>
        )}
      </button>

      {open && (
        <div className="absolute top-full left-0 mt-1 w-64 bg-dark-bg-elevated border border-dark-border-medium rounded-lg shadow-linear-xl z-50 overflow-hidden text-sm">
          {!category ? (
            <>
              <div className="px-3 py-2 text-xs text-dark-text-tertiary font-semibold border-b border-dark-border-subtle">Filter by</div>
              {CATEGORIES.map(cat => (
                <button
                  key={cat.id}
                  type="button"
                  onClick={() => { setCategory(cat.id); setSearch('') }}
                  className="w-full flex items-center gap-3 px-3 py-2.5 text-dark-text-secondary hover:bg-primary-500/10 transition-colors"
                >
                  <span className="text-dark-text-tertiary w-4 flex-shrink-0">{cat.icon}</span>
                  <span className="flex-1 text-left">{cat.label}</span>
                  {activeVal(cat.id) && <span className="text-xs text-primary-400 font-mono truncate max-w-[80px]">{activeVal(cat.id)}</span>}
                  <svg className="w-3 h-3 text-dark-text-tertiary flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
                </button>
              ))}
              {hasFilters && (
                <div className="border-t border-dark-border-subtle px-3 py-2">
                  <button
                    type="button"
                    onClick={() => { onChange({ sprintId: null, assigneeId: null, priority: '', tagId: null, taskIds: [] }); setOpen(false) }}
                    className="text-xs text-danger-400 hover:text-danger-300"
                  >
                    Clear all filters
                  </button>
                </div>
              )}
            </>
          ) : category === 'task_id' ? (
            <>
              <div className="flex items-center gap-2 px-3 py-2 border-b border-dark-border-subtle">
                <button type="button" onClick={() => { setCategory(null); setTaskIdInput('') }} className="text-dark-text-tertiary hover:text-dark-text-secondary">
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" /></svg>
                </button>
                <span className="text-xs text-dark-text-tertiary font-semibold uppercase tracking-wide">Filter by Task ID</span>
              </div>
              <div className="px-3 py-2 border-b border-dark-border-subtle">
                <div className="flex items-center gap-2">
                  <input
                    autoFocus
                    type="number"
                    min="1"
                    placeholder="Enter task # and press Enter"
                    value={taskIdInput}
                    onChange={e => setTaskIdInput(e.target.value)}
                    onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addTaskId() } }}
                    className="flex-1 text-sm bg-transparent text-dark-text-secondary placeholder:text-dark-text-quaternary outline-none [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                  />
                  <button
                    type="button"
                    onClick={addTaskId}
                    disabled={!taskIdInput.trim()}
                    className="text-xs px-2 py-1 rounded bg-primary-500/20 text-primary-400 hover:bg-primary-500/30 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                  >
                    Add
                  </button>
                </div>
              </div>
              <div className="px-3 py-2 min-h-[2.5rem]">
                {taskIds.length === 0 ? (
                  <p className="text-xs text-dark-text-quaternary">No task IDs selected</p>
                ) : (
                  <div className="flex flex-wrap gap-1.5">
                    {taskIds.map(id => (
                      <span key={id} className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-primary-500/10 text-primary-400 font-mono text-xs">
                        #{id}
                        <button type="button" onClick={() => removeTaskId(id)} className="text-primary-400/60 hover:text-primary-400 leading-none">×</button>
                      </span>
                    ))}
                  </div>
                )}
              </div>
              {taskIds.length > 0 && (
                <div className="border-t border-dark-border-subtle px-3 py-2">
                  <button type="button" onClick={() => onChange({ taskIds: [] })} className="text-xs text-danger-400 hover:text-danger-300">
                    Clear task IDs
                  </button>
                </div>
              )}
            </>
          ) : (
            <>
              <div className="flex items-center gap-2 px-3 py-2 border-b border-dark-border-subtle">
                <button type="button" onClick={() => { setCategory(null); setSearch('') }} className="text-dark-text-tertiary hover:text-dark-text-secondary">
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" /></svg>
                </button>
                <span className="text-xs text-dark-text-tertiary font-semibold uppercase tracking-wide">Filter by {category}</span>
              </div>
              {category !== 'priority' && (
                <div className="px-3 py-2 border-b border-dark-border-subtle">
                  <input
                    autoFocus
                    type="text"
                    placeholder={`Search ${category}s…`}
                    value={search}
                    onChange={e => setSearch(e.target.value)}
                    className="w-full text-sm bg-transparent text-dark-text-secondary placeholder:text-dark-text-quaternary outline-none"
                  />
                </div>
              )}
              <div className="max-h-52 overflow-y-auto">
                {options.length === 0 ? (
                  <div className="px-3 py-3 text-dark-text-tertiary">No results</div>
                ) : options.map(opt => {
                  const isActive =
                    (category === 'sprint'   && sprintId   === Number(opt.value)) ||
                    (category === 'assignee' && assigneeId === Number(opt.value)) ||
                    (category === 'priority' && priority   === opt.value) ||
                    (category === 'tag'      && tagId      === Number(opt.value))
                  return (
                    <button
                      key={opt.value}
                      type="button"
                      onClick={() => selectOption(category, opt.value)}
                      className={`w-full flex items-center gap-3 px-3 py-2 hover:bg-primary-500/10 transition-colors ${isActive ? 'text-primary-400' : 'text-dark-text-secondary'}`}
                    >
                      <span className={`w-3 h-3 rounded-full border-2 flex-shrink-0 ${isActive ? 'bg-primary-500 border-primary-500' : 'border-dark-border-strong'}`} />
                      {opt.color && <span className="w-2.5 h-2.5 rounded-full flex-shrink-0" style={{ backgroundColor: opt.color }} />}
                      <span className="flex-1 text-left">{opt.label}</span>
                      {opt.sub && <span className="text-xs text-dark-text-tertiary">{opt.sub}</span>}
                    </button>
                  )
                })}
              </div>
            </>
          )}
        </div>
      )}
    </div>
  )
}
