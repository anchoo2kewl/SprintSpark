import { useMemo, useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Task, Milestone, Sprint } from '../../lib/api'

interface TimelineViewProps {
  milestones: Milestone[]
  tasks: Task[]
  sprints: Sprint[]
  projectId: number
  onTaskClick?: (task: Task) => void
}

type ZoomLevel = 'day' | 'week' | 'month'

function addDays(date: Date, days: number): Date {
  const d = new Date(date)
  d.setDate(d.getDate() + days)
  return d
}

function startOfWeek(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay()
  d.setDate(d.getDate() - (day === 0 ? 6 : day - 1))
  d.setHours(0, 0, 0, 0)
  return d
}

function formatDateShort(date: Date): string {
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

function formatMonthYear(date: Date): string {
  return date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' })
}

export default function TimelineView({ milestones, tasks, sprints, projectId, onTaskClick }: TimelineViewProps) {
  const navigate = useNavigate()
  const [zoom, setZoom] = useState<ZoomLevel>('week')
  const [collapsedMilestones, setCollapsedMilestones] = useState<Set<number>>(new Set())
  const scrollRef = useRef<HTMLDivElement>(null)

  // Compute date range from tasks and milestones
  const { dateRange, columns } = useMemo(() => {
    const now = new Date()
    let minDate = addDays(now, -14)
    let maxDate = addDays(now, 90)

    const allDates: Date[] = []
    tasks.forEach(t => {
      if (t.start_date) allDates.push(new Date(t.start_date))
      if (t.due_date) allDates.push(new Date(t.due_date))
    })
    milestones.forEach(m => {
      if (m.target_date) allDates.push(new Date(m.target_date))
    })
    sprints.forEach(s => {
      if (s.start_date) allDates.push(new Date(s.start_date))
      if (s.end_date) allDates.push(new Date(s.end_date))
    })

    if (allDates.length > 0) {
      const sorted = allDates.sort((a, b) => a.getTime() - b.getTime())
      minDate = addDays(sorted[0], -7)
      maxDate = addDays(sorted[sorted.length - 1], 14)
    }

    // Ensure at least 12 weeks of range
    if (maxDate.getTime() - minDate.getTime() < 84 * 24 * 60 * 60 * 1000) {
      maxDate = addDays(minDate, 84)
    }

    const cols: { date: Date; label: string; isToday: boolean }[] = []
    const start = startOfWeek(minDate)
    const today = new Date()
    today.setHours(0, 0, 0, 0)

    if (zoom === 'week') {
      let current = new Date(start)
      while (current <= maxDate) {
        cols.push({
          date: new Date(current),
          label: formatDateShort(current),
          isToday: current.getTime() === today.getTime(),
        })
        current = addDays(current, 7)
      }
    } else if (zoom === 'day') {
      let current = new Date(minDate)
      current.setHours(0, 0, 0, 0)
      while (current <= maxDate) {
        cols.push({
          date: new Date(current),
          label: current.getDate().toString(),
          isToday: current.getTime() === today.getTime(),
        })
        current = addDays(current, 1)
      }
    } else {
      let current = new Date(start.getFullYear(), start.getMonth(), 1)
      while (current <= maxDate) {
        cols.push({
          date: new Date(current),
          label: formatMonthYear(current),
          isToday: current.getMonth() === today.getMonth() && current.getFullYear() === today.getFullYear(),
        })
        current = new Date(current.getFullYear(), current.getMonth() + 1, 1)
      }
    }

    return { dateRange: { min: start, max: maxDate }, columns: cols }
  }, [tasks, milestones, sprints, zoom])

  // Scroll to today on mount
  useEffect(() => {
    if (scrollRef.current && columns.length > 0) {
      const todayIndex = columns.findIndex(c => c.isToday)
      if (todayIndex > 2) {
        const colWidth = zoom === 'day' ? 40 : zoom === 'week' ? 120 : 160
        scrollRef.current.scrollLeft = (todayIndex - 2) * colWidth
      }
    }
  }, [columns, zoom])

  const colWidth = zoom === 'day' ? 40 : zoom === 'week' ? 120 : 160

  // Get position/width for a date range
  const getBarStyle = (startStr: string | null | undefined, endStr: string | null | undefined) => {
    if (!startStr && !endStr) return null
    const start = startStr ? new Date(startStr) : endStr ? addDays(new Date(endStr), -7) : new Date()
    const end = endStr ? new Date(endStr) : addDays(start, 7)

    const totalMs = dateRange.max.getTime() - dateRange.min.getTime()
    if (totalMs <= 0) return null

    const left = ((start.getTime() - dateRange.min.getTime()) / totalMs) * 100
    const width = ((end.getTime() - start.getTime()) / totalMs) * 100

    return { left: `${Math.max(0, left)}%`, width: `${Math.max(1, width)}%` }
  }

  // Group tasks by milestone
  const tasksByMilestone = useMemo(() => {
    const map: Record<number | 'unassigned', Task[]> = { unassigned: [] }
    milestones.forEach(m => { map[m.id] = [] })
    tasks.forEach(t => {
      const key = (t as Task & { milestone_id?: number }).milestone_id || 'unassigned'
      if (!map[key]) map[key] = []
      map[key].push(t)
    })
    return map
  }, [tasks, milestones])

  const toggleCollapse = (id: number) => {
    setCollapsedMilestones(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const handleTaskClick = (task: Task) => {
    if (onTaskClick) onTaskClick(task)
    else if (task.task_number) navigate(`/app/projects/${projectId}/tasks/${task.task_number}`)
  }

  // Today marker position
  const todayPos = (() => {
    const now = new Date()
    const totalMs = dateRange.max.getTime() - dateRange.min.getTime()
    if (totalMs <= 0) return null
    const pct = ((now.getTime() - dateRange.min.getTime()) / totalMs) * 100
    if (pct < 0 || pct > 100) return null
    return `${pct}%`
  })()

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-dark-border-subtle">
        <div className="flex items-center gap-2">
          <span className="text-xs text-dark-text-tertiary uppercase tracking-wider font-medium">Zoom</span>
          {(['day', 'week', 'month'] as ZoomLevel[]).map(z => (
            <button
              key={z}
              onClick={() => setZoom(z)}
              className={`px-2.5 py-1 text-xs rounded transition-colors ${
                zoom === z
                  ? 'bg-primary-500/20 text-primary-400 font-medium'
                  : 'text-dark-text-secondary hover:text-dark-text-primary hover:bg-dark-bg-tertiary'
              }`}
            >
              {z.charAt(0).toUpperCase() + z.slice(1)}
            </button>
          ))}
        </div>
        <div className="text-xs text-dark-text-tertiary">
          {tasks.length} tasks across {milestones.length} milestones
        </div>
      </div>

      {/* Timeline */}
      <div ref={scrollRef} className="flex-1 overflow-auto">
        <div style={{ minWidth: columns.length * colWidth }} className="relative">
          {/* Column headers */}
          <div className="sticky top-0 z-10 flex border-b border-dark-border-subtle bg-dark-bg-secondary">
            <div className="sticky left-0 z-20 w-56 shrink-0 bg-dark-bg-secondary border-r border-dark-border-subtle px-3 py-2">
              <span className="text-xs font-medium text-dark-text-tertiary uppercase tracking-wider">Milestone</span>
            </div>
            <div className="flex relative" style={{ width: columns.length * colWidth }}>
              {columns.map((col, i) => (
                <div
                  key={i}
                  className={`shrink-0 px-2 py-2 text-xs text-center border-r border-dark-border-subtle/30 ${
                    col.isToday ? 'text-primary-400 font-medium bg-primary-500/5' : 'text-dark-text-tertiary'
                  }`}
                  style={{ width: colWidth }}
                >
                  {col.label}
                </div>
              ))}
            </div>
          </div>

          {/* Sprint shading */}
          {sprints.filter(s => s.start_date && s.end_date).map(sprint => {
            const style = getBarStyle(sprint.start_date, sprint.end_date)
            if (!style) return null
            return (
              <div
                key={`sprint-${sprint.id}`}
                className="absolute top-8 bottom-0 bg-primary-500/[0.03] border-x border-primary-500/10 pointer-events-none z-0"
                style={{ left: style.left, width: style.width }}
              />
            )
          })}

          {/* Today marker */}
          {todayPos && (
            <div
              className="absolute top-0 bottom-0 w-px bg-primary-500/40 z-[5] pointer-events-none"
              style={{ left: todayPos, marginLeft: '224px' }}
            />
          )}

          {/* Milestone rows */}
          {milestones.map(milestone => {
            const milestoneTasks = tasksByMilestone[milestone.id] || []
            const isCollapsed = collapsedMilestones.has(milestone.id)
            const completedCount = milestoneTasks.filter(t => t.status === 'done').length
            const pct = milestoneTasks.length > 0 ? Math.round((completedCount / milestoneTasks.length) * 100) : 0

            return (
              <div key={milestone.id}>
                {/* Milestone header row */}
                <div className="flex border-b border-dark-border-subtle/50 hover:bg-dark-bg-tertiary/30 transition-colors">
                  <div className="sticky left-0 z-10 w-56 shrink-0 bg-dark-bg-primary border-r border-dark-border-subtle px-3 py-2 flex items-center gap-2">
                    <button
                      onClick={() => toggleCollapse(milestone.id)}
                      className="text-dark-text-tertiary hover:text-dark-text-primary transition-colors"
                    >
                      <svg className={`w-3.5 h-3.5 transition-transform ${isCollapsed ? '' : 'rotate-90'}`} fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clipRule="evenodd" />
                      </svg>
                    </button>
                    <div className="w-2.5 h-2.5 rounded-full shrink-0" style={{ backgroundColor: milestone.color }} />
                    <span className="text-sm font-medium text-dark-text-primary truncate">{milestone.name}</span>
                    <span className="text-xs text-dark-text-tertiary ml-auto shrink-0">{pct}%</span>
                  </div>
                  <div className="flex-1 relative py-2" style={{ minWidth: columns.length * colWidth }}>
                    {/* Milestone target date marker */}
                    {milestone.target_date && (() => {
                      const totalMs = dateRange.max.getTime() - dateRange.min.getTime()
                      const targetMs = new Date(milestone.target_date).getTime() - dateRange.min.getTime()
                      const pctPos = (targetMs / totalMs) * 100
                      if (pctPos < 0 || pctPos > 100) return null
                      return (
                        <div
                          className="absolute top-1/2 -translate-y-1/2 z-[2]"
                          style={{ left: `${pctPos}%` }}
                          title={`Target: ${milestone.target_date}`}
                        >
                          <div className="w-3 h-3 rotate-45 border-2" style={{ borderColor: milestone.color, backgroundColor: `${milestone.color}33` }} />
                        </div>
                      )
                    })()}
                    {/* Progress bar spanning task date range */}
                    {milestoneTasks.length > 0 && (() => {
                      const dates = milestoneTasks.flatMap(t => [t.start_date, t.due_date].filter(Boolean)).map(d => new Date(d!))
                      if (dates.length === 0) return null
                      const min = new Date(Math.min(...dates.map(d => d.getTime())))
                      const max = new Date(Math.max(...dates.map(d => d.getTime())))
                      const style = getBarStyle(min.toISOString(), max.toISOString())
                      if (!style) return null
                      return (
                        <div className="absolute top-1/2 -translate-y-1/2 h-2 rounded-full opacity-30" style={{ ...style, backgroundColor: milestone.color }} />
                      )
                    })()}
                  </div>
                </div>

                {/* Task rows */}
                {!isCollapsed && milestoneTasks.map(task => {
                  const barStyle = getBarStyle(task.start_date, task.due_date)
                  const isDone = task.status === 'done'
                  return (
                    <div
                      key={task.id}
                      className="flex border-b border-dark-border-subtle/20 hover:bg-dark-bg-tertiary/20 transition-colors cursor-pointer"
                      onClick={() => handleTaskClick(task)}
                    >
                      <div className="sticky left-0 z-10 w-56 shrink-0 bg-dark-bg-primary border-r border-dark-border-subtle px-3 py-1.5 pl-10 flex items-center gap-2">
                        <span className={`text-xs ${isDone ? 'text-dark-text-quaternary' : 'text-dark-text-tertiary'}`}>
                          #{task.task_number}
                        </span>
                        <span className={`text-sm truncate ${isDone ? 'text-dark-text-quaternary line-through' : 'text-dark-text-secondary'}`}>
                          {task.title}
                        </span>
                      </div>
                      <div className="flex-1 relative py-1.5" style={{ minWidth: columns.length * colWidth }}>
                        {barStyle && (
                          <div
                            className={`absolute top-1/2 -translate-y-1/2 h-5 rounded transition-colors ${
                              isDone ? 'opacity-40' : 'hover:brightness-110'
                            }`}
                            style={{
                              ...barStyle,
                              backgroundColor: milestone.color,
                            }}
                            title={`${task.title} (${task.start_date || '?'} - ${task.due_date || '?'})`}
                          >
                            <span className="absolute inset-0 flex items-center px-2 text-[10px] text-white font-medium truncate">
                              {task.title}
                            </span>
                          </div>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            )
          })}

          {/* Unassigned tasks */}
          {tasksByMilestone.unassigned.length > 0 && (
            <div>
              <div className="flex border-b border-dark-border-subtle/50">
                <div className="sticky left-0 z-10 w-56 shrink-0 bg-dark-bg-primary border-r border-dark-border-subtle px-3 py-2 flex items-center gap-2">
                  <div className="w-2.5 h-2.5 rounded-full bg-dark-text-quaternary shrink-0" />
                  <span className="text-sm font-medium text-dark-text-tertiary">No Milestone</span>
                  <span className="text-xs text-dark-text-quaternary ml-auto">{tasksByMilestone.unassigned.length}</span>
                </div>
                <div className="flex-1" style={{ minWidth: columns.length * colWidth }} />
              </div>
              {tasksByMilestone.unassigned.filter(t => t.start_date || t.due_date).map(task => {
                const barStyle = getBarStyle(task.start_date, task.due_date)
                return (
                  <div
                    key={task.id}
                    className="flex border-b border-dark-border-subtle/20 hover:bg-dark-bg-tertiary/20 transition-colors cursor-pointer"
                    onClick={() => handleTaskClick(task)}
                  >
                    <div className="sticky left-0 z-10 w-56 shrink-0 bg-dark-bg-primary border-r border-dark-border-subtle px-3 py-1.5 pl-10 flex items-center gap-2">
                      <span className="text-xs text-dark-text-tertiary">#{task.task_number}</span>
                      <span className="text-sm text-dark-text-secondary truncate">{task.title}</span>
                    </div>
                    <div className="flex-1 relative py-1.5" style={{ minWidth: columns.length * colWidth }}>
                      {barStyle && (
                        <div
                          className="absolute top-1/2 -translate-y-1/2 h-5 rounded bg-dark-text-quaternary/40 hover:brightness-110 transition-colors"
                          style={barStyle}
                          title={task.title}
                        >
                          <span className="absolute inset-0 flex items-center px-2 text-[10px] text-dark-text-secondary font-medium truncate">
                            {task.title}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
