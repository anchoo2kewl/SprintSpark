import { useState, FormEvent, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../state/AuthContext'
import { validateSignupForm } from '../lib/validation'
import Card, { CardHeader, CardBody } from '../components/ui/Card'
import TextInput from '../components/ui/TextInput'
import Button from '../components/ui/Button'
import FormError from '../components/ui/FormError'

export default function Signup() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [touched, setTouched] = useState<Record<string, boolean>>({})
  const { signup, error, loading, clearError, user } = useAuth()
  const navigate = useNavigate()

  // Redirect if already logged in
  useEffect(() => {
    if (user) {
      navigate('/app', { replace: true })
    }
  }, [user, navigate])

  const handleBlur = (field: string) => {
    setTouched({ ...touched, [field]: true })

    // Validate on blur
    const validation = validateSignupForm(email, password, confirmPassword)
    setFieldErrors(validation.errors)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    clearError()

    // Mark all fields as touched
    setTouched({ email: true, password: true, confirmPassword: true })

    // Validate form
    const validation = validateSignupForm(email, password, confirmPassword)
    setFieldErrors(validation.errors)

    if (!validation.isValid) {
      return
    }

    try {
      await signup({ email, password })
      // AuthContext will update user, useEffect will redirect
    } catch (err) {
      // Error is handled by AuthContext
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-secondary-50 via-secondary-100 to-purple-100 px-4">
      <Card className="max-w-md w-full">
        <CardHeader>
          <div className="text-center">
            <img
              src="/logo.svg"
              alt="SprintSpark"
              className="mx-auto h-20 w-20 mb-4"
            />
            <h2 className="text-3xl font-bold text-gray-900">
              SprintSpark
            </h2>
            <p className="mt-2 text-sm text-gray-600">
              Create your account
            </p>
          </div>
        </CardHeader>

        <CardBody>
          <form className="space-y-6" onSubmit={handleSubmit}>
            <FormError message={error || ''} />

            <div className="space-y-4">
              <TextInput
                id="email"
                name="email"
                type="email"
                label="Email address"
                autoComplete="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onBlur={() => handleBlur('email')}
                error={touched.email ? fieldErrors.email : undefined}
                placeholder="you@example.com"
                disabled={loading}
              />

              <TextInput
                id="password"
                name="password"
                type="password"
                label="Password"
                autoComplete="new-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onBlur={() => handleBlur('password')}
                error={touched.password ? fieldErrors.password : undefined}
                helpText="Must be at least 8 characters with a letter and number"
                placeholder="••••••••"
                disabled={loading}
              />

              <TextInput
                id="confirm-password"
                name="confirm-password"
                type="password"
                label="Confirm Password"
                autoComplete="new-password"
                required
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                onBlur={() => handleBlur('confirmPassword')}
                error={touched.confirmPassword ? fieldErrors.confirmPassword : undefined}
                placeholder="••••••••"
                disabled={loading}
              />
            </div>

            <Button
              type="submit"
              variant="secondary"
              fullWidth
              loading={loading}
            >
              Create account
            </Button>

            <div className="text-sm text-center">
              <span className="text-gray-600">Already have an account? </span>
              <Link to="/login" className="font-medium text-secondary-600 hover:text-secondary-500">
                Sign in
              </Link>
            </div>
          </form>
        </CardBody>
      </Card>
    </div>
  )
}
