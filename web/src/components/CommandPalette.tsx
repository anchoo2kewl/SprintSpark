/**
 * Command Palette (Cmd+K / Ctrl+K)
 * Linear-style command palette for quick navigation and actions
 */

import { useState, useEffect, useMemo, Fragment } from 'react'
import { useNavigate } from 'react-router-dom'
import { Dialog, Transition, Combobox } from '@headlessui/react'
import { useSync } from '../state/SyncContext'
import { useAuth } from '../state/AuthContext'
import type { TaskDocument, ProjectDocument } from '../lib/db/schema'

interface Command {
  id: string
  name: string
  description?: string
  icon: string
  action: () => void
  category: 'navigation' | 'actions' | 'search' | 'create'
  keywords?: string[]
}

export default function CommandPalette() {
  const [isOpen, setIsOpen] = useState(false)
  const [query, setQuery] = useState('')
  const { db } = useSync()
  const { logout } = useAuth()
  const navigate = useNavigate()

  const [tasks, setTasks] = useState<TaskDocument[]>([])
  const [projects, setProjects] = useState<ProjectDocument[]>([])

  // Load data from local database
  useEffect(() => {
    if (!db) return

    const tasksSub = db.tasks
      .find({ selector: { _deleted: { $ne: true } }, limit: 50 })
      .$.subscribe(docs => setTasks(docs.map(d => d.toJSON())))

    const projectsSub = db.projects
      .find({ selector: { _deleted: { $ne: true } } })
      .$.subscribe(docs => setProjects(docs.map(d => d.toJSON())))

    return () => {
      tasksSub.unsubscribe()
      projectsSub.unsubscribe()
    }
  }, [db])

  // Keyboard shortcut to open/close palette
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd+K (Mac) or Ctrl+K (Windows/Linux)
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setIsOpen(open => !open)
      }
      // Escape to close
      if (e.key === 'Escape') {
        setIsOpen(false)
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  // Reset query when closed
  useEffect(() => {
    if (!isOpen) {
      setQuery('')
    }
  }, [isOpen])

  // Define available commands
  const staticCommands: Command[] = [
    // Navigation
    {
      id: 'nav-projects',
      name: 'Go to Projects',
      description: 'View all projects',
      icon: '📁',
      category: 'navigation',
      keywords: ['projects', 'list'],
      action: () => {
        navigate('/app')
        setIsOpen(false)
      },
    },
    {
      id: 'nav-cycles',
      name: 'Go to Cycles',
      description: 'View sprint cycles',
      icon: '🔄',
      category: 'navigation',
      keywords: ['sprints', 'cycles', 'iterations'],
      action: () => {
        navigate('/app/sprints')
        setIsOpen(false)
      },
    },
    {
      id: 'nav-tags',
      name: 'Go to Tags',
      description: 'Manage labels and tags',
      icon: '🏷️',
      category: 'navigation',
      keywords: ['tags', 'labels'],
      action: () => {
        navigate('/app/tags')
        setIsOpen(false)
      },
    },
    {
      id: 'nav-settings',
      name: 'Go to Settings',
      description: 'Account and preferences',
      icon: '⚙️',
      category: 'navigation',
      keywords: ['settings', 'preferences', 'config'],
      action: () => {
        navigate('/app/settings')
        setIsOpen(false)
      },
    },
    // Actions
    {
      id: 'action-logout',
      name: 'Logout',
      description: 'Sign out of your account',
      icon: '🚪',
      category: 'actions',
      keywords: ['logout', 'signout', 'exit'],
      action: () => {
        logout()
        navigate('/login')
        setIsOpen(false)
      },
    },
  ]

  // Dynamic commands from data
  const dynamicCommands: Command[] = useMemo(() => {
    const commands: Command[] = []

    // Add projects
    projects.forEach(project => {
      commands.push({
        id: `project-${project.id}`,
        name: project.name,
        description: project.description || 'Open project',
        icon: '📂',
        category: 'search',
        keywords: ['project', project.name.toLowerCase()],
        action: () => {
          navigate(`/app/projects/${project.id}`)
          setIsOpen(false)
        },
      })
    })

    // Add recent tasks (limit to 20 for performance)
    tasks.slice(0, 20).forEach(task => {
      commands.push({
        id: `task-${task.id}`,
        name: task.title,
        description: `${task.status} · Project ${task.project_id}`,
        icon: task.status === 'done' ? '✅' : task.status === 'in_progress' ? '🔄' : '📝',
        category: 'search',
        keywords: ['task', 'issue', task.title.toLowerCase()],
        action: () => {
          navigate(`/app/projects/${task.project_id}/tasks/${task.id}`)
          setIsOpen(false)
        },
      })
    })

    return commands
  }, [projects, tasks, navigate])

  const allCommands = [...staticCommands, ...dynamicCommands]

  // Filter commands based on query
  const filteredCommands = useMemo(() => {
    if (!query) return allCommands

    const lowerQuery = query.toLowerCase()

    return allCommands.filter(cmd => {
      // Search in name
      if (cmd.name.toLowerCase().includes(lowerQuery)) return true
      // Search in description
      if (cmd.description?.toLowerCase().includes(lowerQuery)) return true
      // Search in keywords
      if (cmd.keywords?.some(k => k.includes(lowerQuery))) return true
      return false
    })
  }, [query, allCommands])

  // Group commands by category
  const groupedCommands = useMemo(() => {
    const groups = {
      navigation: [] as Command[],
      create: [] as Command[],
      actions: [] as Command[],
      search: [] as Command[],
    }

    filteredCommands.forEach(cmd => {
      groups[cmd.category].push(cmd)
    })

    return groups
  }, [filteredCommands])

  const categoryLabels: Record<string, string> = {
    navigation: 'Navigation',
    create: 'Create',
    actions: 'Actions',
    search: 'Search Results',
  }

  return (
    <Transition.Root show={isOpen} as={Fragment} afterLeave={() => setQuery('')}>
      <Dialog className="relative z-50" onClose={setIsOpen}>
        {/* Backdrop */}
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-200"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-150"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black bg-opacity-40 backdrop-blur-sm transition-opacity" />
        </Transition.Child>

        <div className="fixed inset-0 z-10 overflow-y-auto p-4 sm:p-6 md:p-20">
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-200"
            enterFrom="opacity-0 scale-95"
            enterTo="opacity-100 scale-100"
            leave="ease-in duration-150"
            leaveFrom="opacity-100 scale-100"
            leaveTo="opacity-0 scale-95"
          >
            <Dialog.Panel className="mx-auto max-w-2xl transform rounded-lg bg-white shadow-linear-lg ring-1 ring-black ring-opacity-5 transition-all">
              <Combobox onChange={(command: Command | null) => command?.action()}>
                {/* Search input */}
                <div className="relative">
                  <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                    <span className="text-gray-400 text-xl">🔍</span>
                  </div>
                  <Combobox.Input
                    className="h-14 w-full border-0 bg-transparent pl-12 pr-4 text-gray-900 placeholder-gray-400 focus:ring-0 text-base"
                    placeholder="Search or type a command..."
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setQuery(e.target.value)}
                    autoFocus
                  />
                  <div className="absolute inset-y-0 right-0 flex items-center pr-4">
                    <kbd className="hidden sm:inline-block px-2 py-1 text-xs font-semibold text-gray-500 bg-gray-100 rounded">
                      Esc
                    </kbd>
                  </div>
                </div>

                {/* Results */}
                {filteredCommands.length > 0 && (
                  <Combobox.Options
                    static
                    className="max-h-96 scroll-py-2 overflow-y-auto border-t border-gray-100"
                  >
                    {Object.entries(groupedCommands).map(([category, commands]) => {
                      if (commands.length === 0) return null

                      return (
                        <div key={category} className="p-2">
                          <div className="px-3 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wide">
                            {categoryLabels[category as keyof typeof categoryLabels]}
                          </div>
                          {commands.map((command) => (
                            <Combobox.Option
                              key={command.id}
                              value={command}
                              className={({ active }: { active: boolean }) =>
                                `flex cursor-pointer select-none items-center rounded-md px-3 py-2 ${
                                  active ? 'bg-primary-500 text-white' : 'text-gray-900'
                                }`
                              }
                            >
                              {({ active }: { active: boolean }) => (
                                <>
                                  <span className="mr-3 text-xl">{command.icon}</span>
                                  <div className="flex-1 min-w-0">
                                    <p className={`text-sm font-medium truncate ${active ? 'text-white' : 'text-gray-900'}`}>
                                      {command.name}
                                    </p>
                                    {command.description && (
                                      <p className={`text-xs truncate ${active ? 'text-primary-100' : 'text-gray-500'}`}>
                                        {command.description}
                                      </p>
                                    )}
                                  </div>
                                  <span className={`ml-3 text-xs ${active ? 'text-primary-100' : 'text-gray-400'}`}>
                                    ↵
                                  </span>
                                </>
                              )}
                            </Combobox.Option>
                          ))}
                        </div>
                      )
                    })}
                  </Combobox.Options>
                )}

                {/* Empty state */}
                {query && filteredCommands.length === 0 && (
                  <div className="px-6 py-14 text-center">
                    <p className="text-sm text-gray-500">No results found for "{query}"</p>
                  </div>
                )}

                {/* Footer hint */}
                {!query && (
                  <div className="border-t border-gray-100 px-4 py-3 text-xs text-gray-500 bg-gray-50">
                    <div className="flex items-center justify-between">
                      <span>
                        Type to search • Use <kbd className="px-1.5 py-0.5 bg-white rounded border border-gray-300">↑↓</kbd> to navigate
                      </span>
                      <span>
                        <kbd className="px-1.5 py-0.5 bg-white rounded border border-gray-300">⌘K</kbd> to toggle
                      </span>
                    </div>
                  </div>
                )}
              </Combobox>
            </Dialog.Panel>
          </Transition.Child>
        </div>
      </Dialog>
    </Transition.Root>
  )
}
