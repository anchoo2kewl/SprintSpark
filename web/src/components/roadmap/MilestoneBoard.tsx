import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Task, Milestone, MilestoneProgress } from '../../lib/api'

interface MilestoneBoardProps {
  milestones: Milestone[]
  tasks: Task[]
  projectId: number
  progressMap: Record<number, MilestoneProgress>
  onMilestoneEdit: (milestone: Milestone) => void
  onRefresh: () => void
}

const PRIORITY_COLORS: Record<string, string> = {
  urgent: '#ef4444',
  high: '#f97316',
  medium: '#eab308',
  low: '#6b7280',
}

export default function MilestoneBoard({ milestones, tasks, projectId, progressMap, onMilestoneEdit, onRefresh: _onRefresh }: MilestoneBoardProps) {
  const navigate = useNavigate()

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

  return (
    <div className="flex-1 overflow-x-auto p-4">
      <div className="flex gap-4" style={{ minWidth: (milestones.length + 1) * 320 }}>
        {milestones.map(milestone => {
          const milestoneTasks = tasksByMilestone[milestone.id] || []
          const progress = progressMap[milestone.id]
          const pct = progress ? progress.percentage : 0

          return (
            <div
              key={milestone.id}
              className="w-80 shrink-0 flex flex-col bg-dark-bg-secondary rounded-lg border border-dark-border-subtle"
            >
              {/* Column header */}
              <div className="p-3 border-b border-dark-border-subtle">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: milestone.color }} />
                    <h3 className="text-sm font-semibold text-dark-text-primary truncate">{milestone.name}</h3>
                    <span className="text-xs text-dark-text-tertiary">{milestoneTasks.length}</span>
                  </div>
                  <button
                    onClick={() => onMilestoneEdit(milestone)}
                    className="text-dark-text-quaternary hover:text-dark-text-secondary transition-colors p-0.5"
                    title="Edit milestone"
                  >
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                    </svg>
                  </button>
                </div>
                {/* Progress bar */}
                <div className="h-1.5 bg-dark-bg-tertiary rounded-full overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all duration-300"
                    style={{ width: `${pct}%`, backgroundColor: milestone.color }}
                  />
                </div>
                <div className="flex items-center justify-between mt-1">
                  <span className="text-[10px] text-dark-text-quaternary">{Math.round(pct)}% complete</span>
                  {milestone.target_date && (
                    <span className="text-[10px] text-dark-text-quaternary">
                      Due {new Date(milestone.target_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                    </span>
                  )}
                </div>
              </div>

              {/* Task cards */}
              <div className="flex-1 overflow-y-auto p-2 space-y-1.5 max-h-[calc(100vh-280px)]">
                {milestoneTasks.map(task => {
                  const isDone = (task.status || '') === 'done'
                  return (
                    <div
                      key={task.id}
                      onClick={() => task.task_number && navigate(`/app/projects/${projectId}/tasks/${task.task_number}`)}
                      className={`p-2.5 rounded-md border border-dark-border-subtle/50 hover:border-dark-border-medium cursor-pointer transition-all ${
                        isDone ? 'opacity-50' : 'hover:bg-dark-bg-tertiary/30'
                      }`}
                    >
                      <div className="flex items-start gap-2">
                        <div className="w-1 h-4 rounded-full mt-0.5 shrink-0" style={{ backgroundColor: PRIORITY_COLORS[task.priority || 'medium'] || '#6b7280' }} />
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center gap-1.5 mb-0.5">
                            <span className="text-[10px] text-dark-text-quaternary">#{task.task_number}</span>
                            {isDone && (
                              <svg className="w-3 h-3 text-success-500" fill="currentColor" viewBox="0 0 20 20">
                                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                              </svg>
                            )}
                          </div>
                          <p className={`text-sm leading-snug ${isDone ? 'text-dark-text-quaternary line-through' : 'text-dark-text-primary'}`}>
                            {task.title}
                          </p>
                          <div className="flex items-center gap-2 mt-1.5">
                            {task.assignees && task.assignees.length > 0 && (
                              <span className="text-[10px] text-dark-text-tertiary truncate max-w-[120px]">
                                {task.assignees.map(a => a.user_name).join(', ')}
                              </span>
                            )}
                            {task.due_date && (
                              <span className={`text-[10px] ${new Date(task.due_date) < new Date() && !isDone ? 'text-danger-400' : 'text-dark-text-quaternary'}`}>
                                {new Date(task.due_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  )
                })}
                {milestoneTasks.length === 0 && (
                  <div className="text-center py-6 text-xs text-dark-text-quaternary">
                    No tasks assigned
                  </div>
                )}
              </div>
            </div>
          )
        })}

        {/* Unassigned column */}
        {tasksByMilestone.unassigned.length > 0 && (
          <div className="w-80 shrink-0 flex flex-col bg-dark-bg-secondary rounded-lg border border-dark-border-subtle/50 border-dashed">
            <div className="p-3 border-b border-dark-border-subtle">
              <div className="flex items-center gap-2">
                <div className="w-2.5 h-2.5 rounded-full bg-dark-text-quaternary" />
                <h3 className="text-sm font-semibold text-dark-text-tertiary">No Milestone</h3>
                <span className="text-xs text-dark-text-quaternary">{tasksByMilestone.unassigned.length}</span>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto p-2 space-y-1.5 max-h-[calc(100vh-280px)]">
              {tasksByMilestone.unassigned.map(task => (
                <div
                  key={task.id}
                  onClick={() => task.task_number && navigate(`/app/projects/${projectId}/tasks/${task.task_number}`)}
                  className="p-2.5 rounded-md border border-dark-border-subtle/30 hover:border-dark-border-subtle cursor-pointer transition-all"
                >
                  <div className="flex items-center gap-1.5 mb-0.5">
                    <span className="text-[10px] text-dark-text-quaternary">#{task.task_number}</span>
                  </div>
                  <p className="text-sm text-dark-text-secondary leading-snug">{task.title}</p>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
