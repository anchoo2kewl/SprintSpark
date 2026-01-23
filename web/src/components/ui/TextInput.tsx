import { InputHTMLAttributes, forwardRef } from 'react'

interface TextInputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
  helpText?: string
}

const TextInput = forwardRef<HTMLInputElement, TextInputProps>(
  ({ label, error, helpText, className = '', ...props }, ref) => {
    return (
      <div className="w-full">
        <label htmlFor={props.id} className="block text-xs font-medium text-dark-text-primary mb-1">
          {label}
          {props.required && <span className="text-danger-400 ml-1">*</span>}
        </label>
        <input
          ref={ref}
          className={`
            appearance-none block w-full px-3 py-2 border rounded-md shadow-sm
            placeholder-dark-text-tertiary focus:outline-none text-sm
            bg-dark-bg-primary text-dark-text-primary
            ${error
              ? 'border-danger-500/50 focus:ring-1 focus:ring-danger-500 focus:border-danger-500'
              : 'border-dark-bg-tertiary/30 focus:ring-1 focus:ring-primary-500 focus:border-primary-500'
            }
            ${props.disabled ? 'bg-dark-bg-tertiary/20 cursor-not-allowed opacity-50' : ''}
            ${className}
          `}
          aria-invalid={error ? 'true' : 'false'}
          aria-describedby={error ? `${props.id}-error` : helpText ? `${props.id}-help` : undefined}
          {...props}
        />
        {helpText && !error && (
          <p id={`${props.id}-help`} className="mt-1 text-xs text-dark-text-secondary">
            {helpText}
          </p>
        )}
        {error && (
          <p id={`${props.id}-error`} className="mt-1 text-xs text-danger-400" role="alert">
            {error}
          </p>
        )}
      </div>
    )
  }
)

TextInput.displayName = 'TextInput'

export default TextInput
