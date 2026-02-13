import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import Card, { CardHeader, CardBody } from './Card'

describe('Card', () => {
  it('renders children', () => {
    render(<Card><p>Card content</p></Card>)
    expect(screen.getByText('Card content')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(<Card className="my-custom-class">Content</Card>)
    const card = screen.getByText('Content').closest('div')!
    expect(card.className).toContain('my-custom-class')
  })

  it('applies default classes', () => {
    render(<Card>Content</Card>)
    const card = screen.getByText('Content').closest('div')!
    expect(card.className).toContain('bg-dark-bg-elevated')
    expect(card.className).toContain('border')
    expect(card.className).toContain('rounded-xl')
    expect(card.className).toContain('shadow-linear')
  })

  it('passes id to the root element', () => {
    render(<Card id="test-card">Content</Card>)
    expect(document.getElementById('test-card')).toBeInTheDocument()
  })
})

describe('CardHeader', () => {
  it('renders children', () => {
    render(<CardHeader><h2>Header Title</h2></CardHeader>)
    expect(screen.getByText('Header Title')).toBeInTheDocument()
  })

  it('applies default padding classes', () => {
    render(<CardHeader>Header</CardHeader>)
    const header = screen.getByText('Header').closest('div')!
    expect(header.className).toContain('px-8')
    expect(header.className).toContain('pt-8')
    expect(header.className).toContain('pb-4')
  })

  it('applies custom className', () => {
    render(<CardHeader className="extra-class">Header</CardHeader>)
    const header = screen.getByText('Header').closest('div')!
    expect(header.className).toContain('extra-class')
  })
})

describe('CardBody', () => {
  it('renders children', () => {
    render(<CardBody><span>Body content</span></CardBody>)
    expect(screen.getByText('Body content')).toBeInTheDocument()
  })

  it('applies default padding classes', () => {
    render(<CardBody>Body</CardBody>)
    const body = screen.getByText('Body').closest('div')!
    expect(body.className).toContain('px-8')
    expect(body.className).toContain('pb-8')
  })

  it('applies custom className', () => {
    render(<CardBody className="body-extra">Body</CardBody>)
    const body = screen.getByText('Body').closest('div')!
    expect(body.className).toContain('body-extra')
  })
})
