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
        <label htmlFor={props.id} className="block text-sm font-medium text-gray-700">
          {label}
          {props.required && <span className="text-danger-500 ml-1">*</span>}
        </label>
        <input
          ref={ref}
          className={`
            mt-1 appearance-none block w-full px-3 py-2 border rounded-md shadow-sm
            placeholder-gray-400 focus:outline-none sm:text-sm
            ${error
              ? 'border-danger-300 focus:ring-danger-500 focus:border-danger-500'
              : 'border-gray-300 focus:ring-primary-500 focus:border-primary-500'
            }
            ${props.disabled ? 'bg-gray-100 cursor-not-allowed' : ''}
            ${className}
          `}
          aria-invalid={error ? 'true' : 'false'}
          aria-describedby={error ? `${props.id}-error` : helpText ? `${props.id}-help` : undefined}
          {...props}
        />
        {helpText && !error && (
          <p id={`${props.id}-help`} className="mt-1 text-xs text-gray-500">
            {helpText}
          </p>
        )}
        {error && (
          <p id={`${props.id}-error`} className="mt-1 text-xs text-danger-600" role="alert">
            {error}
          </p>
        )}
      </div>
    )
  }
)

TextInput.displayName = 'TextInput'

export default TextInput
