import { describe, it, expect } from 'vitest'
import {
  validateEmail,
  validatePassword,
  validatePasswordConfirmation,
  validateLoginForm,
  validateSignupForm,
} from './validation'

describe('validateEmail', () => {
  it('returns error for empty email', () => {
    expect(validateEmail('')).toBe('Email is required')
  })

  it('returns error for invalid email format', () => {
    expect(validateEmail('notanemail')).toBe('Please enter a valid email address')
    expect(validateEmail('missing@')).toBe('Please enter a valid email address')
    expect(validateEmail('@missing.com')).toBe('Please enter a valid email address')
    expect(validateEmail('spaces in@email.com')).toBe('Please enter a valid email address')
  })

  it('returns null for valid email', () => {
    expect(validateEmail('user@example.com')).toBeNull()
    expect(validateEmail('test.user@domain.co')).toBeNull()
    expect(validateEmail('name+tag@example.org')).toBeNull()
  })
})

describe('validatePassword', () => {
  it('returns error for empty password', () => {
    expect(validatePassword('')).toBe('Password is required')
  })

  it('returns error for short password', () => {
    expect(validatePassword('Ab1')).toBe('Password must be at least 8 characters')
    expect(validatePassword('Short1')).toBe('Password must be at least 8 characters')
  })

  it('returns error when missing letters', () => {
    expect(validatePassword('12345678')).toBe('Password must contain at least one letter')
  })

  it('returns error when missing numbers', () => {
    expect(validatePassword('abcdefgh')).toBe('Password must contain at least one number')
  })

  it('returns null for valid password', () => {
    expect(validatePassword('password1')).toBeNull()
    expect(validatePassword('Str0ngPass!')).toBeNull()
  })

  it('respects custom minLength', () => {
    expect(validatePassword('Ab1', 4)).toBe('Password must be at least 4 characters')
    expect(validatePassword('Ab12', 4)).toBeNull()
  })
})

describe('validatePasswordConfirmation', () => {
  it('returns error for empty confirmation', () => {
    expect(validatePasswordConfirmation('password1', '')).toBe('Please confirm your password')
  })

  it('returns error when passwords do not match', () => {
    expect(validatePasswordConfirmation('password1', 'password2')).toBe('Passwords do not match')
  })

  it('returns null when passwords match', () => {
    expect(validatePasswordConfirmation('password1', 'password1')).toBeNull()
  })
})

describe('validateLoginForm', () => {
  it('returns errors for empty fields', () => {
    const result = validateLoginForm('', '')
    expect(result.isValid).toBe(false)
    expect(result.errors.email).toBe('Email is required')
    expect(result.errors.password).toBe('Password is required')
  })

  it('returns error for invalid email only', () => {
    const result = validateLoginForm('bad', 'password')
    expect(result.isValid).toBe(false)
    expect(result.errors.email).toBe('Please enter a valid email address')
    expect(result.errors.password).toBeUndefined()
  })

  it('returns error for missing password only', () => {
    const result = validateLoginForm('user@example.com', '')
    expect(result.isValid).toBe(false)
    expect(result.errors.email).toBeUndefined()
    expect(result.errors.password).toBe('Password is required')
  })

  it('returns valid for correct input', () => {
    const result = validateLoginForm('user@example.com', 'password')
    expect(result.isValid).toBe(true)
    expect(Object.keys(result.errors)).toHaveLength(0)
  })
})

describe('validateSignupForm', () => {
  it('returns errors for all empty fields', () => {
    const result = validateSignupForm('', '', '')
    expect(result.isValid).toBe(false)
    expect(result.errors.email).toBeDefined()
    expect(result.errors.password).toBeDefined()
    expect(result.errors.confirmPassword).toBeDefined()
  })

  it('returns error for weak password', () => {
    const result = validateSignupForm('user@example.com', 'short', 'short')
    expect(result.isValid).toBe(false)
    expect(result.errors.password).toContain('at least 8 characters')
  })

  it('returns error for mismatched passwords', () => {
    const result = validateSignupForm('user@example.com', 'password1', 'password2')
    expect(result.isValid).toBe(false)
    expect(result.errors.confirmPassword).toBe('Passwords do not match')
  })

  it('returns valid for correct input', () => {
    const result = validateSignupForm('user@example.com', 'password1', 'password1')
    expect(result.isValid).toBe(true)
    expect(Object.keys(result.errors)).toHaveLength(0)
  })

  it('validates password strength in signup', () => {
    const result = validateSignupForm('user@example.com', 'nodigits!', 'nodigits!')
    expect(result.isValid).toBe(false)
    expect(result.errors.password).toContain('at least one number')
  })
})
