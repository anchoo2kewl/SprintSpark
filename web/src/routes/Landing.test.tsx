import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import Landing from './Landing'

vi.mock('react-router-dom', () => ({
  Link: ({ to, children, ...props }: { to: string; children: React.ReactNode; [key: string]: unknown }) => (
    <a href={to} {...props}>{children}</a>
  ),
}))

describe('Landing', () => {
  it('renders the hero heading', () => {
    render(<Landing />)
    expect(screen.getByText('AI-native')).toBeInTheDocument()
    expect(screen.getByText('project management')).toBeInTheDocument()
  })

  it('renders "Sign in" and "Get started" navigation links', () => {
    render(<Landing />)
    const signInLinks = screen.getAllByText('Sign in')
    expect(signInLinks.length).toBeGreaterThanOrEqual(1)

    const getStartedLinks = screen.getAllByText('Get started')
    expect(getStartedLinks.length).toBeGreaterThanOrEqual(1)
  })

  it('renders feature sections', () => {
    render(<Landing />)
    expect(screen.getByText('Built for AI, designed for humans')).toBeInTheDocument()
    expect(screen.getByText('MCP Server')).toBeInTheDocument()
    expect(screen.getByText('API-First Architecture')).toBeInTheDocument()
    expect(screen.getByText('Visual Project Management')).toBeInTheDocument()
  })

  it('renders the "How AI agents use TaskAI" section', () => {
    render(<Landing />)
    expect(screen.getByText('How AI agents use TaskAI')).toBeInTheDocument()
    expect(screen.getByText('Connect via MCP')).toBeInTheDocument()
    expect(screen.getByText('AI discovers available tools')).toBeInTheDocument()
    expect(screen.getByText('LLM works independently')).toBeInTheDocument()
  })

  it('renders the footer', () => {
    render(<Landing />)
    expect(screen.getByText('AI-native project management. Ship with confidence.')).toBeInTheDocument()
  })

  it('has links to /signup', () => {
    render(<Landing />)
    const signupLinks = screen.getAllByRole('link').filter(
      (link) => link.getAttribute('href') === '/signup'
    )
    expect(signupLinks.length).toBeGreaterThanOrEqual(1)
  })

  it('has links to /login', () => {
    render(<Landing />)
    const loginLinks = screen.getAllByRole('link').filter(
      (link) => link.getAttribute('href') === '/login'
    )
    expect(loginLinks.length).toBeGreaterThanOrEqual(1)
  })
})
