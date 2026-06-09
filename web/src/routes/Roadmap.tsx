import { useEffect, useState, useMemo, useCallback } from 'react'
import { api, type Task, type Milestone, type Sprint, type MilestoneProgress, type Tag } from '../lib/api'
import MilestoneModal from '../components/roadmap/MilestoneModal'
import TimelineView from '../components/roadmap/TimelineView'
import MilestoneBoard from '../components/roadmap/MilestoneBoard'
import RoadmapList from '../components/roadmap/RoadmapList'
import BoardFilterBar, { applyBoardFilters } from '../components/board/BoardFilterBar'
import { useSync } from '../state/SyncContext'

type RoadmapView = 'timeline' | 'board' | 'list'

interface RoadmapProps {
  projectId: number
  tasks: Task[]
}

export default function Roadmap({ projectId, tasks }: RoadmapProps) {
  const { registerSyncTask } = useSync()
  const [view, setView] = useState<RoadmapView>(() => {
    return (localStorage.getItem(`roadmap_view_${projectId}`) as RoadmapView) || 'timeline'
  })
  const [milestones, setMilestones] = useState<Milestone[]>([])
  const [sprints, setSprints] = useState<Sprint[]>([])
  const [tags, setTags] = useState<Tag[]>([])
  const [progressMap, setProgressMap] = useState<Record<number, MilestoneProgress>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Milestone modal state
  const [showMilestoneModal, setShowMilestoneModal] = useState(false)
  const [editingMilestone, setEditingMilestone] = useState<Milestone | null>(null)

  // Roadmap filters — kept on a separate localStorage key from the board filters
  // so toggling filters in the roadmap doesn't stomp the board's saved state.
  const [filterSprint, setFilterSprint] = useState<number | null>(null)
  const [filterAssignee, setFilterAssignee] = useState<number | null>(null)
  const [filterPriority, setFilterPriority] = useState('')
  const [filterTag, setFilterTag] = useState<number | null>(null)
  const [filterTaskIds, setFilterTaskIds] = useState<number[]>([])

  useEffect(() => {
    try {
      const raw = localStorage.getItem(`taskai_roadmap_filters_${projectId}`)
      if (raw) {
        const s = JSON.parse(raw)
        setFilterSprint(s.sprint ?? null)
        setFilterAssignee(s.assignee ?? null)
        setFilterPriority(s.priority ?? '')
        setFilterTag(s.tag ?? null)
        setFilterTaskIds(Array.isArray(s.taskIds) ? s.taskIds : [])
      } else {
        setFilterSprint(null)
        setFilterAssignee(null)
        setFilterPriority('')
        setFilterTag(null)
        setFilterTaskIds([])
      }
    } catch { /* ignore */ }
  }, [projectId])

  useEffect(() => {
    localStorage.setItem(`taskai_roadmap_filters_${projectId}`, JSON.stringify({
      sprint: filterSprint,
      assignee: filterAssignee,
      priority: filterPriority,
      tag: filterTag,
      taskIds: filterTaskIds,
    }))
  }, [projectId, filterSprint, filterAssignee, filterPriority, filterTag, filterTaskIds])

  const fetchMilestones = useCallback(async (silent = false) => {
    try {
      if (!silent) setLoading(true)
      setError(null)
      const [ms, sp, tg] = await Promise.all([
        api.getMilestones(projectId),
        api.getSprints(projectId),
        api.getTags(projectId).catch(() => [] as Tag[]),
      ])
      setMilestones(ms)
      setSprints(sp)
      setTags(tg)

      // Fetch progress for each milestone
      const progressEntries = await Promise.all(
        ms.map(async m => {
          try {
            const p = await api.getMilestoneProgress(m.id)
            return [m.id, p] as const
          } catch {
            return null
          }
        })
      )
      const map: Record<number, MilestoneProgress> = {}
      progressEntries.forEach(entry => {
        if (entry) map[entry[0]] = entry[1]
      })
      setProgressMap(map)
    } catch (err) {
      if (silent) throw err
      setError(err instanceof Error ? err.message : 'Failed to load roadmap data')
    } finally {
      if (!silent) setLoading(false)
    }
  }, [projectId])

  useEffect(() => {
    void fetchMilestones()
  }, [fetchMilestones])

  useEffect(() => {
    return registerSyncTask(`project:${projectId}:roadmap`, () => fetchMilestones(true))
  }, [fetchMilestones, projectId, registerSyncTask])

  const setViewPersisted = (v: RoadmapView) => {
    setView(v)
    localStorage.setItem(`roadmap_view_${projectId}`, v)
  }

  const handleMilestoneSaved = (milestone: Milestone) => {
    setMilestones(prev => {
      const exists = prev.find(m => m.id === milestone.id)
      if (exists) return prev.map(m => m.id === milestone.id ? milestone : m)
      return [...prev, milestone]
    })
    setEditingMilestone(null)
    // Refresh progress
    void fetchMilestones()
  }

  const handleMilestoneEdit = (milestone: Milestone) => {
    setEditingMilestone(milestone)
    setShowMilestoneModal(true)
  }

  // Active milestones for display
  const activeMilestones = useMemo(() => milestones.filter(m => m.status !== 'cancelled'), [milestones])
  const completedCount = useMemo(() => milestones.filter(m => m.status === 'completed').length, [milestones])

  // Unique assignees derived from the loaded tasks (mirrors ProjectDetail).
  const uniqueAssignees = useMemo(() => {
    const map = new Map<number, string>()
    tasks.forEach(t => {
      if (t.assignees?.length) {
        t.assignees.forEach(a => {
          if (a.user_id != null && !map.has(a.user_id)) map.set(a.user_id, a.user_name ?? `User ${a.user_id}`)
        })
      } else if (t.assignee_id && !map.has(t.assignee_id)) {
        map.set(t.assignee_id, t.assignee_name || `User ${t.assignee_id}`)
      }
    })
    return Array.from(map.entries()).map(([id, name]) => ({ id, name }))
  }, [tasks])

  const filteredTasks = useMemo(
    () => applyBoardFilters(tasks, {
      sprintId: filterSprint,
      assigneeId: filterAssignee,
      priority: filterPriority,
      tagId: filterTag,
      taskIds: filterTaskIds,
    }),
    [tasks, filterSprint, filterAssignee, filterPriority, filterTag, filterTaskIds]
  )

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex items-center gap-3">
          <svg className="w-5 h-5 animate-spin text-primary-400" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          <span className="text-sm text-dark-text-secondary">Loading roadmap...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-sm text-danger-400 mb-2">{error}</p>
          <button
            onClick={() => { setError(null); setLoading(true); fetchMilestones() }}
            className="text-xs text-primary-400 hover:text-primary-300"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Roadmap header */}
      <div className="flex items-center justify-between px-6 py-3 border-b border-dark-border-subtle">
        <div className="flex items-center gap-4">
          <h2 className="text-sm font-semibold text-dark-text-primary">Roadmap</h2>

          {/* View switcher */}
          <div className="flex items-center bg-dark-bg-tertiary rounded-md p-0.5">
            {[
              { key: 'timeline' as const, label: 'Timeline', icon: (
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              )},
              { key: 'board' as const, label: 'Board', icon: (
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
                </svg>
              )},
              { key: 'list' as const, label: 'List', icon: (
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" />
                </svg>
              )},
            ].map(v => (
              <button
                key={v.key}
                onClick={() => setViewPersisted(v.key)}
                className={`flex items-center gap-1.5 px-2.5 py-1 rounded text-xs transition-colors ${
                  view === v.key
                    ? 'bg-dark-bg-primary text-dark-text-primary shadow-sm'
                    : 'text-dark-text-tertiary hover:text-dark-text-secondary'
                }`}
                title={v.label}
              >
                {v.icon}
                <span className="hidden sm:inline">{v.label}</span>
              </button>
            ))}
          </div>

          <BoardFilterBar
            sprints={sprints}
            assignees={uniqueAssignees}
            tags={tags}
            sprintId={filterSprint}
            assigneeId={filterAssignee}
            priority={filterPriority}
            tagId={filterTag}
            taskIds={filterTaskIds}
            onChange={patch => {
              if ('sprintId'   in patch) setFilterSprint(patch.sprintId ?? null)
              if ('assigneeId' in patch) setFilterAssignee(patch.assigneeId ?? null)
              if ('priority'   in patch) setFilterPriority(patch.priority ?? '')
              if ('tagId'      in patch) setFilterTag(patch.tagId ?? null)
              if ('taskIds'    in patch) setFilterTaskIds(patch.taskIds ?? [])
            }}
          />
        </div>

        <div className="flex items-center gap-3">
          {/* Stats */}
          <div className="hidden sm:flex items-center gap-3 text-xs text-dark-text-tertiary">
            <span>{activeMilestones.length} milestones</span>
            {completedCount > 0 && <span className="text-success-400">{completedCount} completed</span>}
          </div>

          {/* Add milestone */}
          <button
            onClick={() => { setEditingMilestone(null); setShowMilestoneModal(true) }}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-primary-500/10 text-primary-400 hover:bg-primary-500/20 rounded-md text-xs font-medium transition-colors"
          >
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Milestone
          </button>
        </div>
      </div>

      {/* Empty state */}
      {milestones.length === 0 ? (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center max-w-sm">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-primary-500/10 flex items-center justify-center">
              <svg className="w-8 h-8 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-dark-text-primary mb-1">Plan your roadmap</h3>
            <p className="text-sm text-dark-text-tertiary mb-4">
              Create milestones to group tasks into deliverables and track progress toward your project goals.
            </p>
            <button
              onClick={() => { setEditingMilestone(null); setShowMilestoneModal(true) }}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary-500 text-white rounded-lg text-sm font-medium hover:bg-primary-600 transition-colors"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Create First Milestone
            </button>
          </div>
        </div>
      ) : (
        <div className="flex-1 overflow-hidden">
          {view === 'timeline' && (
            <TimelineView
              milestones={activeMilestones}
              tasks={filteredTasks}
              sprints={sprints}
              projectId={projectId}
            />
          )}
          {view === 'board' && (
            <MilestoneBoard
              milestones={activeMilestones}
              tasks={filteredTasks}
              projectId={projectId}
              progressMap={progressMap}
              onMilestoneEdit={handleMilestoneEdit}
            />
          )}
          {view === 'list' && (
            <RoadmapList
              milestones={activeMilestones}
              tasks={filteredTasks}
              sprints={sprints}
              projectId={projectId}
            />
          )}
        </div>
      )}

      {/* Milestone Modal */}
      <MilestoneModal
        isOpen={showMilestoneModal}
        onClose={() => { setShowMilestoneModal(false); setEditingMilestone(null) }}
        onSaved={handleMilestoneSaved}
        projectId={projectId}
        milestone={editingMilestone}
      />
    </div>
  )
}
