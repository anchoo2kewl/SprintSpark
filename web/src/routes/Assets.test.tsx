import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import Assets from './Assets'

vi.mock('../components/MyAssets', () => ({
  default: () => <div data-testid="my-assets">MyAssets</div>,
}))

describe('Assets', () => {
  it('renders MyAssets component', () => {
    render(<Assets />)
    expect(screen.getByTestId('my-assets')).toBeInTheDocument()
    expect(screen.getByText('MyAssets')).toBeInTheDocument()
  })
})
