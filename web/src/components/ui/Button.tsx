import { ButtonHTMLAttributes, ReactNode } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode
  variant?: 'primary' | 'secondary' | 'danger' | 'outline' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  fullWidth?: boolean
}

export default function Button({
  children,
  variant = 'primary',
  size = 'md',
  loading = false,
  fullWidth = false,
  disabled,
  className = '',
  ...props
}: ButtonProps) {
  const baseClasses = 'inline-flex items-center justify-center font-medium rounded-md focus:outline-none focus:ring-1 transition-all duration-150'

  const variantClasses = {
    primary: 'bg-primary-500 text-white hover:bg-primary-600 focus:ring-primary-500/50 disabled:bg-primary-500/50 shadow-linear-sm',
    secondary: 'bg-dark-bg-tertiary text-dark-text-primary hover:bg-dark-bg-elevated focus:ring-dark-border-strong disabled:bg-dark-bg-tertiary/30',
    danger: 'bg-danger-500 text-white hover:bg-danger-600 focus:ring-danger-500/50 disabled:bg-danger-500/50 shadow-linear-sm',
    outline: 'bg-transparent text-dark-text-secondary border border-dark-border-medium hover:bg-dark-bg-tertiary hover:text-dark-text-primary hover:border-dark-border-strong focus:ring-primary-500/30 disabled:opacity-30',
    ghost: 'bg-transparent text-dark-text-tertiary hover:bg-dark-bg-tertiary hover:text-dark-text-primary disabled:opacity-30',
  }

  const sizeClasses = {
    sm: 'px-3 py-2 md:py-1.5 text-sm',
    md: 'px-4 py-2.5 md:py-2 text-sm',
    lg: 'px-6 py-3 text-base',
  }

  const widthClass = fullWidth ? 'w-full' : ''

  return (
    <button
      className={`
        ${baseClasses}
        ${variantClasses[variant]}
        ${sizeClasses[size]}
        ${widthClass}
        ${(disabled || loading) ? 'cursor-not-allowed opacity-50' : ''}
        ${className}
      `}
      disabled={disabled || loading}
      {...props}
    >
      {loading && (
        <svg
          className="animate-spin -ml-1 mr-2 h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      {children}
    </button>
  )
}
