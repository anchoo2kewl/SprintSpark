import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import TextInput from './TextInput'

describe('TextInput', () => {
  it('renders with label', () => {
    render(<TextInput label="Email" id="email" />)
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
  })

  it('accepts user input', async () => {
    const user = userEvent.setup()
    render(<TextInput label="Name" id="name" />)

    const input = screen.getByLabelText('Name')
    await user.type(input, 'Hello')
    expect(input).toHaveValue('Hello')
  })

  it('shows error message', () => {
    render(<TextInput label="Email" id="email" error="Invalid email" />)
    expect(screen.getByRole('alert')).toHaveTextContent('Invalid email')
  })

  it('sets aria-invalid when error is present', () => {
    render(<TextInput label="Email" id="email" error="Required" />)
    expect(screen.getByLabelText('Email')).toHaveAttribute('aria-invalid', 'true')
  })

  it('sets aria-invalid to false when no error', () => {
    render(<TextInput label="Email" id="email" />)
    expect(screen.getByLabelText('Email')).toHaveAttribute('aria-invalid', 'false')
  })

  it('shows help text when no error', () => {
    render(<TextInput label="Name" id="name" helpText="Enter your full name" />)
    expect(screen.getByText('Enter your full name')).toBeInTheDocument()
  })

  it('hides help text when error is present', () => {
    render(
      <TextInput label="Name" id="name" helpText="Enter your full name" error="Required" />
    )
    expect(screen.queryByText('Enter your full name')).not.toBeInTheDocument()
    expect(screen.getByRole('alert')).toHaveTextContent('Required')
  })

  it('shows required indicator', () => {
    render(<TextInput label="Email" id="email" required />)
    expect(screen.getByText('*')).toBeInTheDocument()
  })

  it('is disabled when disabled prop is set', () => {
    render(<TextInput label="Email" id="email" disabled />)
    expect(screen.getByLabelText('Email')).toBeDisabled()
  })

  it('passes placeholder through', () => {
    render(<TextInput label="Search" id="search" placeholder="Search..." />)
    expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument()
  })

  it('calls onChange handler', async () => {
    const user = userEvent.setup()
    const handleChange = vi.fn()
    render(<TextInput label="Test" id="test" onChange={handleChange} />)

    await user.type(screen.getByLabelText('Test'), 'a')
    expect(handleChange).toHaveBeenCalled()
  })
})
