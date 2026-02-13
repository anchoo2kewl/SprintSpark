import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import FormError from './FormError'

describe('FormError', () => {
  it('renders error message', () => {
    render(<FormError message="Something went wrong" />)
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
  })

  it('has alert role for accessibility', () => {
    render(<FormError message="Error occurred" />)
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  it('renders nothing when message is empty', () => {
    const { container } = render(<FormError message="" />)
    expect(container.firstChild).toBeNull()
  })

  it('applies custom className', () => {
    render(<FormError message="Error" className="mt-4" />)
    const alert = screen.getByRole('alert')
    expect(alert.className).toContain('mt-4')
  })

  it('renders error icon SVG', () => {
    render(<FormError message="Error" />)
    const alert = screen.getByRole('alert')
    const svg = alert.querySelector('svg')
    expect(svg).toBeInTheDocument()
    expect(svg).toHaveAttribute('aria-hidden', 'true')
  })

  it('applies default styling classes', () => {
    render(<FormError message="Error" />)
    const alert = screen.getByRole('alert')
    expect(alert.className).toContain('rounded-md')
    expect(alert.className).toContain('bg-danger-500/10')
  })
})
