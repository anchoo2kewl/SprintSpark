export interface ValidationError {
  field: string
  message: string
}

export interface ValidationResult {
  isValid: boolean
  errors: Record<string, string>
}

// Email validation
export function validateEmail(email: string): string | null {
  if (!email) {
    return 'Email is required'
  }

  // Basic email regex
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(email)) {
    return 'Please enter a valid email address'
  }

  return null
}

// Password validation
export function validatePassword(password: string, minLength = 8): string | null {
  if (!password) {
    return 'Password is required'
  }

  if (password.length < minLength) {
    return `Password must be at least ${minLength} characters`
  }

  // Check for at least one letter and one number
  if (!/[a-zA-Z]/.test(password)) {
    return 'Password must contain at least one letter'
  }

  if (!/[0-9]/.test(password)) {
    return 'Password must contain at least one number'
  }

  return null
}

// Password confirmation validation
export function validatePasswordConfirmation(
  password: string,
  confirmPassword: string
): string | null {
  if (!confirmPassword) {
    return 'Please confirm your password'
  }

  if (password !== confirmPassword) {
    return 'Passwords do not match'
  }

  return null
}

// Validate login form
export function validateLoginForm(email: string, password: string): ValidationResult {
  const errors: Record<string, string> = {}

  const emailError = validateEmail(email)
  if (emailError) {
    errors.email = emailError
  }

  if (!password) {
    errors.password = 'Password is required'
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  }
}

// Validate signup form
export function validateSignupForm(
  email: string,
  password: string,
  confirmPassword: string
): ValidationResult {
  const errors: Record<string, string> = {}

  const emailError = validateEmail(email)
  if (emailError) {
    errors.email = emailError
  }

  const passwordError = validatePassword(password)
  if (passwordError) {
    errors.password = passwordError
  }

  const confirmError = validatePasswordConfirmation(password, confirmPassword)
  if (confirmError) {
    errors.confirmPassword = confirmError
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  }
}
