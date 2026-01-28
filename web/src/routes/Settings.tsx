import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import TextInput from '../components/ui/TextInput'
import FormError from '../components/ui/FormError'
import { apiClient } from '../lib/api'

export default function Settings() {
  const navigate = useNavigate()

  // Password change state
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordError, setPasswordError] = useState('')
  const [passwordSuccess, setPasswordSuccess] = useState('')
  const [isChangingPassword, setIsChangingPassword] = useState(false)

  // 2FA state
  const [twoFAEnabled, setTwoFAEnabled] = useState(false)
  const [twoFASecret, setTwoFASecret] = useState('')
  const [qrCodeURL, setQrCodeURL] = useState('')
  const [verificationCode, setVerificationCode] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [twoFAError, setTwoFAError] = useState('')
  const [twoFASuccess, setTwoFASuccess] = useState('')
  const [isSettingUp2FA, setIsSettingUp2FA] = useState(false)
  const [isEnabling2FA, setIsEnabling2FA] = useState(false)
  const [showBackupCodes, setShowBackupCodes] = useState(false)
  const [disablePassword, setDisablePassword] = useState('')
  const [isDisabling2FA, setIsDisabling2FA] = useState(false)

  // API Keys state
  const [apiKeys, setApiKeys] = useState<any[]>([])
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyExpires, setNewKeyExpires] = useState<number | undefined>(90)
  const [createdKey, setCreatedKey] = useState<{ key: string; name: string } | null>(null)
  const [apiKeyError, setApiKeyError] = useState('')
  const [apiKeySuccess, setApiKeySuccess] = useState('')
  const [isCreatingKey, setIsCreatingKey] = useState(false)
  const [isDeletingKey, setIsDeletingKey] = useState<number | null>(null)

  useEffect(() => {
    load2FAStatus()
    loadAPIKeys()
  }, [])

  const load2FAStatus = async () => {
    try {
      const status = await apiClient.get2FAStatus()
      setTwoFAEnabled(status.enabled)
    } catch (error) {
      console.error('Failed to load 2FA status:', error)
    }
  }

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault()
    setPasswordError('')
    setPasswordSuccess('')

    if (newPassword !== confirmPassword) {
      setPasswordError('Passwords do not match')
      return
    }

    if (newPassword.length < 8) {
      setPasswordError('Password must be at least 8 characters')
      return
    }

    setIsChangingPassword(true)

    try {
      await apiClient.changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      })

      setPasswordSuccess('Password changed successfully')
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (error: any) {
      setPasswordError(error.message || 'Failed to change password')
    } finally {
      setIsChangingPassword(false)
    }
  }

  const handleSetup2FA = async () => {
    setTwoFAError('')
    setIsSettingUp2FA(true)

    try {
      const response = await apiClient.setup2FA()
      setTwoFASecret(response.secret)
      setQrCodeURL(response.qr_code_url)
      setShowBackupCodes(false)
    } catch (error: any) {
      setTwoFAError(error.message || 'Failed to setup 2FA')
    } finally {
      setIsSettingUp2FA(false)
    }
  }

  const handleEnable2FA = async (e: React.FormEvent) => {
    e.preventDefault()
    setTwoFAError('')
    setTwoFASuccess('')

    if (!verificationCode || verificationCode.length !== 6) {
      setTwoFAError('Please enter a 6-digit verification code')
      return
    }

    setIsEnabling2FA(true)

    try {
      const response = await apiClient.enable2FA({ code: verificationCode })
      setBackupCodes(response.backup_codes)
      setShowBackupCodes(true)
      setTwoFAEnabled(true)
      setTwoFASuccess('2FA enabled successfully! Save your backup codes.')
      setVerificationCode('')
      setQrCodeURL('')
      setTwoFASecret('')
    } catch (error: any) {
      setTwoFAError(error.message || 'Invalid verification code')
    } finally {
      setIsEnabling2FA(false)
    }
  }

  const handleDisable2FA = async (e: React.FormEvent) => {
    e.preventDefault()
    setTwoFAError('')
    setTwoFASuccess('')

    if (!disablePassword) {
      setTwoFAError('Password is required to disable 2FA')
      return
    }

    setIsDisabling2FA(true)

    try {
      await apiClient.disable2FA({ password: disablePassword })
      setTwoFAEnabled(false)
      setTwoFASuccess('2FA disabled successfully')
      setDisablePassword('')
      setBackupCodes([])
      setShowBackupCodes(false)
    } catch (error: any) {
      setTwoFAError(error.message || 'Failed to disable 2FA')
    } finally {
      setIsDisabling2FA(false)
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  const copyAllBackupCodes = () => {
    const allCodes = backupCodes.join('\n')
    copyToClipboard(allCodes)
    setTwoFASuccess('All backup codes copied to clipboard')
  }

  const copySecret = () => {
    copyToClipboard(twoFASecret)
    setTwoFASuccess('Secret key copied to clipboard')
  }

  const loadAPIKeys = async () => {
    try {
      const keys = await apiClient.getAPIKeys()
      setApiKeys(keys)
    } catch (error) {
      console.error('Failed to load API keys:', error)
    }
  }

  const handleCreateAPIKey = async (e: React.FormEvent) => {
    e.preventDefault()
    setApiKeyError('')
    setApiKeySuccess('')

    if (!newKeyName.trim()) {
      setApiKeyError('Key name is required')
      return
    }

    setIsCreatingKey(true)

    try {
      const response = await apiClient.createAPIKey({
        name: newKeyName,
        expires_in: newKeyExpires,
      })

      setCreatedKey({ key: response.key, name: response.name })
      setApiKeySuccess('API key created successfully')
      setNewKeyName('')
      setNewKeyExpires(90)
      await loadAPIKeys()
    } catch (error: any) {
      setApiKeyError(error.message || 'Failed to create API key')
    } finally {
      setIsCreatingKey(false)
    }
  }

  const handleDeleteAPIKey = async (id: number, name: string) => {
    if (!confirm(`Are you sure you want to delete the API key "${name}"?`)) {
      return
    }

    setIsDeletingKey(id)
    setApiKeyError('')
    setApiKeySuccess('')

    try {
      await apiClient.deleteAPIKey(id)
      setApiKeySuccess('API key deleted successfully')
      await loadAPIKeys()
    } catch (error: any) {
      setApiKeyError(error.message || 'Failed to delete API key')
    } finally {
      setIsDeletingKey(null)
    }
  }

  const copyAPIKey = (key: string) => {
    copyToClipboard(key)
    setApiKeySuccess('API key copied to clipboard')
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString()
  }

  return (
    <div className="min-h-screen bg-dark-bg-primary py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-dark-text-primary">Account Settings</h1>
              <p className="text-dark-text-secondary mt-1">Manage your security and authentication preferences</p>
            </div>
            <Button onClick={() => navigate('/app')} variant="secondary">
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back
            </Button>
          </div>
        </div>

        <div className="space-y-6">
          {/* Password Change Section */}
          <Card className="shadow-md">
            <div className="p-6 sm:p-8 flex items-start gap-4">
              <div className="flex-shrink-0 w-10 h-10 bg-primary-500/10 rounded-lg flex items-center justify-center">
                <svg className="w-6 h-6 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <div className="flex-1">
                <h2 className="text-xl font-semibold text-dark-text-primary mb-1">Change Password</h2>
                <p className="text-sm text-dark-text-secondary mb-6">Update your password to keep your account secure</p>

                {passwordSuccess && (
                  <div className="mb-4 p-4 bg-success-500/10 border-l-4 border-success-400 rounded-r-lg">
                    <div className="flex items-center">
                      <svg className="w-5 h-5 text-success-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                      <span className="text-success-300 font-medium">{passwordSuccess}</span>
                    </div>
                  </div>
                )}

                <form onSubmit={handlePasswordChange} className="space-y-4">
                  <TextInput
                    label="Current Password"
                    type="password"
                    value={currentPassword}
                    onChange={(e) => setCurrentPassword(e.target.value)}
                    required
                    autoComplete="current-password"
                  />

                  <TextInput
                    label="New Password"
                    type="password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    required
                    autoComplete="new-password"
                    helpText="Must be at least 8 characters with a letter and number"
                  />

                  <TextInput
                    label="Confirm New Password"
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    required
                    autoComplete="new-password"
                  />

                  {passwordError && <FormError message={passwordError} />}

                  <Button
                    type="submit"
                    disabled={isChangingPassword}
                    className="w-full sm:w-auto"
                  >
                    {isChangingPassword ? 'Changing Password...' : 'Change Password'}
                  </Button>
                </form>
              </div>
            </div>
          </Card>

          {/* Two-Factor Authentication Section */}
          <Card className="shadow-md">
            <div className="p-6 sm:p-8 flex items-start gap-4">
              <div className="flex-shrink-0 w-10 h-10 bg-purple-500/10 rounded-lg flex items-center justify-center">
                <svg className="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <div className="flex-1">
                <h2 className="text-xl font-semibold text-dark-text-primary mb-1">Two-Factor Authentication</h2>
                <p className="text-sm text-dark-text-secondary mb-6">Add an extra layer of security to your account</p>

                {/* Status Badge */}
                <div className="mb-6 p-4 bg-dark-bg-primary border border-dark-bg-tertiary/30 rounded-lg flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={`w-3 h-3 rounded-full ${twoFAEnabled ? 'bg-success-500' : 'bg-dark-text-tertiary'}`}></div>
                    <div>
                      <p className="font-medium text-dark-text-primary">Status</p>
                      <p className="text-sm text-dark-text-secondary">{twoFAEnabled ? 'Active' : 'Not configured'}</p>
                    </div>
                  </div>
                  <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                    twoFAEnabled
                      ? 'bg-success-500/10 text-success-400'
                      : 'bg-dark-bg-tertiary/30 text-dark-text-tertiary'
                  }`}>
                    {twoFAEnabled ? '✓ Enabled' : 'Disabled'}
                  </span>
                </div>

                {twoFASuccess && (
                  <div className="mb-4 p-4 bg-success-500/10 border-l-4 border-success-400 rounded-r-lg">
                    <div className="flex items-center">
                      <svg className="w-5 h-5 text-success-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                      <span className="text-success-300 font-medium">{twoFASuccess}</span>
                    </div>
                  </div>
                )}

                {twoFAError && <FormError message={twoFAError} className="mb-4" />}

                {/* Enable 2FA Flow */}
                {!twoFAEnabled && !qrCodeURL && (
                  <div>
                    <div className="bg-primary-500/10 border border-primary-500/30 rounded-lg p-4 mb-4">
                      <div className="flex gap-3">
                        <svg className="w-5 h-5 text-primary-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                        </svg>
                        <div className="text-sm text-dark-text-secondary">
                          <p className="font-medium mb-1 text-dark-text-primary">How it works</p>
                          <p>Two-factor authentication adds an extra security layer. You'll need your password and a verification code from your phone to sign in.</p>
                        </div>
                      </div>
                    </div>
                    <Button onClick={handleSetup2FA} disabled={isSettingUp2FA}>
                      {isSettingUp2FA ? 'Setting up...' : 'Enable 2FA'}
                    </Button>
                  </div>
                )}

                {/* QR Code Display */}
                {qrCodeURL && !twoFAEnabled && (
                  <div className="space-y-4">
                    <div className="bg-dark-bg-primary border-2 border-dark-bg-tertiary/30 rounded-xl p-6">
                      <div className="flex items-center gap-2 mb-4">
                        <div className="w-8 h-8 bg-purple-500/10 rounded-full flex items-center justify-center text-purple-400 font-bold">1</div>
                        <h3 className="font-semibold text-dark-text-primary">Scan QR Code</h3>
                      </div>
                      <p className="text-sm text-dark-text-secondary mb-4">
                        Open your authenticator app (Google Authenticator, Authy, 1Password, etc.) and scan this code
                      </p>
                      <div className="flex flex-col items-center gap-4">
                        <div className="bg-white p-4 rounded-lg border-2 border-dark-bg-tertiary/30 shadow-sm">
                          <img src={qrCodeURL} alt="2FA QR Code" className="w-48 h-48" />
                        </div>

                        <div className="w-full bg-dark-bg-secondary p-4 rounded-lg border border-dark-bg-tertiary/30">
                          <p className="text-xs font-medium text-dark-text-secondary mb-2">Manual Entry Key:</p>
                          <div className="flex items-center gap-2">
                            <code className="flex-1 text-sm font-mono bg-dark-bg-primary text-dark-text-primary px-3 py-2 rounded border border-dark-bg-tertiary/30 break-all">
                              {twoFASecret}
                            </code>
                            <Button size="sm" variant="secondary" onClick={copySecret}>
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                              </svg>
                            </Button>
                          </div>
                        </div>
                      </div>
                    </div>

                    <form onSubmit={handleEnable2FA} className="bg-dark-bg-primary border-2 border-dark-bg-tertiary/30 rounded-xl p-6">
                      <div className="flex items-center gap-2 mb-4">
                        <div className="w-8 h-8 bg-purple-500/10 rounded-full flex items-center justify-center text-purple-400 font-bold">2</div>
                        <h3 className="font-semibold text-dark-text-primary">Enter Verification Code</h3>
                      </div>
                      <p className="text-sm text-dark-text-secondary mb-4">
                        Enter the 6-digit code shown in your authenticator app
                      </p>
                      <div className="flex gap-3">
                        <input
                          type="text"
                          value={verificationCode}
                          onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                          placeholder="000000"
                          maxLength={6}
                          pattern="\d{6}"
                          required
                          className="flex-1 text-center text-3xl font-mono tracking-widest px-4 py-3 border-2 border-dark-bg-tertiary/30 bg-dark-bg-secondary text-dark-text-primary rounded-lg focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-colors placeholder-dark-text-tertiary"
                        />
                        <Button type="submit" disabled={isEnabling2FA || verificationCode.length !== 6}>
                          {isEnabling2FA ? 'Verifying...' : 'Verify & Enable'}
                        </Button>
                      </div>
                    </form>
                  </div>
                )}

                {/* Backup Codes Display */}
                {showBackupCodes && backupCodes.length > 0 && (
                  <div className="border-2 border-yellow-500/30 bg-yellow-500/10 rounded-xl p-6">
                    <div className="flex items-start gap-3 mb-4">
                      <svg className="w-6 h-6 text-yellow-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                      </svg>
                      <div>
                        <h3 className="font-bold text-yellow-300 mb-1">Save Your Backup Codes</h3>
                        <p className="text-sm text-yellow-400/90">
                          Store these codes in a safe place. You can use them to access your account if you lose your device. Each code can only be used once.
                        </p>
                      </div>
                    </div>

                    <div className="bg-dark-bg-primary p-5 rounded-lg border-2 border-yellow-500/30 mb-4">
                      <div className="grid grid-cols-2 gap-3">
                        {backupCodes.map((code, index) => (
                          <div key={index} className="flex items-center justify-between p-3 bg-dark-bg-secondary rounded-lg border border-dark-bg-tertiary/30">
                            <span className="font-mono text-sm font-medium text-dark-text-primary">{code}</span>
                            <button
                              onClick={() => {
                                copyToClipboard(code)
                                setTwoFASuccess(`Code ${index + 1} copied`)
                              }}
                              className="text-primary-400 hover:text-primary-300 text-xs font-medium px-2 py-1 rounded hover:bg-primary-500/10 transition-colors"
                            >
                              Copy
                            </button>
                          </div>
                        ))}
                      </div>
                    </div>

                    <Button onClick={copyAllBackupCodes} variant="secondary" className="w-full">
                      <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                      Copy All Codes
                    </Button>
                  </div>
                )}

                {/* Disable 2FA */}
                {twoFAEnabled && !showBackupCodes && (
                  <form onSubmit={handleDisable2FA} className="space-y-4">
                    <div className="bg-danger-500/10 border-l-4 border-danger-400 rounded-r-lg p-4">
                      <div className="flex items-start gap-3">
                        <svg className="w-5 h-5 text-danger-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                        </svg>
                        <div className="flex-1">
                          <h3 className="font-semibold text-danger-300 mb-1">Disable Two-Factor Authentication</h3>
                          <p className="text-sm text-danger-400/90 mb-4">
                            This will make your account less secure. Enter your password to confirm.
                          </p>

                          <TextInput
                            label="Password"
                            type="password"
                            value={disablePassword}
                            onChange={(e) => setDisablePassword(e.target.value)}
                            required
                            autoComplete="current-password"
                          />

                          <Button
                            type="submit"
                            variant="danger"
                            disabled={isDisabling2FA}
                            className="w-full mt-4"
                          >
                            {isDisabling2FA ? 'Disabling...' : 'Disable 2FA'}
                          </Button>
                        </div>
                      </div>
                    </div>
                  </form>
                )}
              </div>
            </div>
          </Card>

          {/* API Keys Section */}
          <Card className="shadow-md">
            <div className="p-6 sm:p-8 flex items-start gap-4">
              <div className="flex-shrink-0 w-10 h-10 bg-blue-500/10 rounded-lg flex items-center justify-center">
                <svg className="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                </svg>
              </div>
              <div className="flex-1">
                <h2 className="text-xl font-semibold text-dark-text-primary mb-1">API Keys</h2>
                <p className="text-sm text-dark-text-secondary mb-6">Create and manage API keys for programmatic access</p>

                {apiKeySuccess && (
                  <div className="mb-4 p-4 bg-success-500/10 border-l-4 border-success-400 rounded-r-lg">
                    <div className="flex items-center">
                      <svg className="w-5 h-5 text-success-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                      <span className="text-success-300 font-medium">{apiKeySuccess}</span>
                    </div>
                  </div>
                )}

                {apiKeyError && <FormError message={apiKeyError} className="mb-4" />}

                {/* Newly Created Key Display */}
                {createdKey && (
                  <div className="mb-6 border-2 border-yellow-500/30 bg-yellow-500/10 rounded-xl p-6">
                    <div className="flex items-start gap-3 mb-4">
                      <svg className="w-6 h-6 text-yellow-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                      </svg>
                      <div>
                        <h3 className="font-bold text-yellow-300 mb-1">Save Your API Key</h3>
                        <p className="text-sm text-yellow-400/90">
                          This is the only time you'll see the full key. Copy it now and store it securely.
                        </p>
                      </div>
                    </div>

                    <div className="bg-dark-bg-primary p-5 rounded-lg border-2 border-yellow-500/30 mb-4">
                      <p className="text-xs font-medium text-dark-text-secondary mb-2">Key Name: {createdKey.name}</p>
                      <div className="flex items-center gap-2">
                        <code className="flex-1 text-sm font-mono bg-dark-bg-secondary text-dark-text-primary px-3 py-2 rounded border border-dark-bg-tertiary/30 break-all">
                          {createdKey.key}
                        </code>
                        <Button size="sm" variant="secondary" onClick={() => copyAPIKey(createdKey.key)}>
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                          </svg>
                        </Button>
                      </div>
                    </div>

                    <Button onClick={() => setCreatedKey(null)} variant="secondary" className="w-full">
                      I've saved my key
                    </Button>
                  </div>
                )}

                {/* Create New Key Form */}
                <form onSubmit={handleCreateAPIKey} className="mb-6 space-y-4">
                  <TextInput
                    label="Key Name"
                    value={newKeyName}
                    onChange={(e) => setNewKeyName(e.target.value)}
                    placeholder="e.g., CI/CD Pipeline, Mobile App"
                    required
                  />

                  <div>
                    <label className="block text-sm font-medium text-dark-text-primary mb-2">
                      Expiration
                    </label>
                    <select
                      value={newKeyExpires || ''}
                      onChange={(e) => setNewKeyExpires(e.target.value ? parseInt(e.target.value) : undefined)}
                      className="w-full px-3 py-2 border border-dark-bg-tertiary/30 bg-dark-bg-secondary text-dark-text-primary rounded-lg focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-colors"
                    >
                      <option value="30">30 days</option>
                      <option value="90">90 days</option>
                      <option value="180">180 days</option>
                      <option value="365">1 year</option>
                      <option value="">Never expires</option>
                    </select>
                  </div>

                  <Button type="submit" disabled={isCreatingKey}>
                    {isCreatingKey ? 'Creating...' : 'Create API Key'}
                  </Button>
                </form>

                {/* Existing Keys List */}
                <div>
                  <h3 className="text-sm font-semibold text-dark-text-primary mb-3">Active API Keys</h3>
                  {apiKeys.length === 0 ? (
                    <div className="text-center py-8 text-dark-text-tertiary">
                      <svg className="w-12 h-12 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                      </svg>
                      <p className="text-sm">No API keys created yet</p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {apiKeys.map((key) => (
                        <div key={key.id} className="p-4 bg-dark-bg-secondary border border-dark-bg-tertiary/30 rounded-lg">
                          <div className="flex items-center justify-between">
                            <div className="flex-1">
                              <div className="flex items-center gap-3">
                                <h4 className="font-medium text-dark-text-primary">{key.name}</h4>
                                <code className="text-xs font-mono bg-dark-bg-primary text-dark-text-secondary px-2 py-1 rounded border border-dark-bg-tertiary/30">
                                  {key.key_prefix}...
                                </code>
                              </div>
                              <div className="mt-1 flex items-center gap-4 text-xs text-dark-text-tertiary">
                                <span>Created: {formatDate(key.created_at)}</span>
                                {key.last_used_at && <span>Last used: {formatDate(key.last_used_at)}</span>}
                                {key.expires_at && <span>Expires: {formatDate(key.expires_at)}</span>}
                              </div>
                            </div>
                            <Button
                              size="sm"
                              variant="danger"
                              onClick={() => handleDeleteAPIKey(key.id, key.name)}
                              disabled={isDeletingKey === key.id}
                            >
                              {isDeletingKey === key.id ? 'Deleting...' : 'Delete'}
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="mt-6 bg-blue-500/10 border border-blue-500/30 rounded-lg p-4">
                  <div className="flex gap-3">
                    <svg className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                    </svg>
                    <div className="text-sm text-dark-text-secondary">
                      <p className="font-medium mb-1 text-dark-text-primary">Using API keys</p>
                      <p className="mb-2">Include your API key in the Authorization header:</p>
                      <code className="block bg-dark-bg-secondary text-dark-text-primary px-3 py-2 rounded border border-dark-bg-tertiary/30 font-mono text-xs">
                        Authorization: ApiKey YOUR_API_KEY
                      </code>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
