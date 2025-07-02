import { useState } from 'react';
import { useWebAuthn } from '../hooks/useWebAuthn.js';

export default function LoginForm({ onSuccess, onShowRegister }) {
  const [username, setUsername] = useState('');
  const [usernameError, setUsernameError] = useState('');
  const [loginMode, setLoginMode] = useState('discoverable'); // 'discoverable' or 'username'
  const { loading, error, clearError, authenticate, isSupported } = useWebAuthn();

  // Username validation regex matching backend
  const usernameRegex = /^[a-zA-Z0-9._-]{3,30}$/;

  const validateUsername = (value) => {
    if (!value) {
      return 'Username is required';
    }
    
    if (value.length < 3) {
      return 'Username must be at least 3 characters long';
    }
    
    if (value.length > 30) {
      return 'Username must be no more than 30 characters long';
    }
    
    if (!usernameRegex.test(value)) {
      return 'Username can only contain letters, numbers, dots, hyphens, and underscores';
    }
    
    if (value.startsWith('.') || value.startsWith('-') || value.startsWith('_') ||
        value.endsWith('.') || value.endsWith('-') || value.endsWith('_')) {
      return 'Username cannot start or end with dots, hyphens, or underscores';
    }
    
    return '';
  };

  const handleUsernameChange = (e) => {
    const value = e.target.value;
    setUsername(value);
    setLoginMode('username');
    
    // Clear previous errors
    if (usernameError) {
      setUsernameError('');
    }
    if (error) {
      clearError();
    }
    
    // Validate on change
    if (value) {
      const validationError = validateUsername(value);
      setUsernameError(validationError);
    }
  };

  const handleUsernameBlur = () => {
    if (username) {
      const validationError = validateUsername(username);
      setUsernameError(validationError);
    }
  };

  const handleUsernameLogin = async (e) => {
    e.preventDefault();
    
    // Validate username before submitting
    const validationError = validateUsername(username);
    if (validationError) {
      setUsernameError(validationError);
      return;
    }

    try {
      const result = await authenticate(username.trim());
      onSuccess?.(result);
    } catch (err) {
      console.error('Login failed:', err);
    }
  };

  const handleDiscoverableLogin = async () => {
    try {
      const result = await authenticate(); // No username = discoverable login
      onSuccess?.(result);
    } catch (err) {
      console.error('Discoverable login failed:', err);
    }
  };

  if (!isSupported()) {
    const isCustomDomain = window.location.hostname.endsWith('.local');
    const needsHTTPS = isCustomDomain && window.location.protocol === 'http:';
    
    return (
      <div className="card">
        <div className="error">
          <strong>WebAuthn Not Available</strong>
          {needsHTTPS ? (
            <div>
              <p><strong>HTTPS Required:</strong> WebAuthn requires HTTPS for custom domains.</p>
              <p><strong>Quick Fix:</strong> Use <code>http://localhost:5173</code> instead of <code>{window.location.origin}</code></p>
              <p><strong>Or:</strong> Set up HTTPS for {window.location.hostname}</p>
              <button 
                onClick={() => window.location.href = 'http://localhost:5173'} 
                className="btn btn-primary"
                style={{ marginTop: '1rem' }}
              >
                Switch to localhost
              </button>
            </div>
          ) : (
            <p>Your browser doesn't support WebAuthn. Please use a modern browser.</p>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="header">
        <h1>🔑 Sign In with Passkey</h1>
        <p>Use your passkey to sign in securely</p>
      </div>

      {error && (
        <div className="error">
          <strong>Sign In Failed</strong>
          <p>{error}</p>
          <button onClick={clearError} className="btn btn-secondary btn-small">
            Try Again
          </button>
        </div>
      )}

      {/* Discoverable Login (Passwordless) */}
      <div style={{ marginBottom: '1.5rem' }}>
        <button
          onClick={handleDiscoverableLogin}
          className="btn btn-primary"
          disabled={loading}
          style={{ width: '100%', marginBottom: '1rem' }}
        >
          {loading && loginMode === 'discoverable' ? (
            <>
              <span className="loading"></span>
              Signing in...
            </>
          ) : (
            <>
              <span className="fingerprint-icon">🔐</span>
              Sign in with Passkey
            </>
          )}
        </button>
        <p style={{ fontSize: '0.875rem', color: '#666', textAlign: 'center' }}>
          No username required - your device will show available passkeys
        </p>
      </div>

      <div className="divider">
        <span>or sign in with username</span>
      </div>

      {/* Username-based Login */}
      <form onSubmit={handleUsernameLogin} className="form">
        <div className="form-group">
          <label htmlFor="loginUsername">Username</label>
          <input
            id="loginUsername"
            type="text"
            value={username}
            onChange={handleUsernameChange}
            onBlur={handleUsernameBlur}
            placeholder="Enter your username"
            autoComplete="username"
            disabled={loading}
            pattern="[a-zA-Z0-9.\\_\\-]{3,30}"
            title="Username must be 3-30 characters and contain only letters, numbers, dots, hyphens, and underscores"
            // iOS-specific attributes to prevent auto-capitalization and correction
            autoCapitalize="none"
            autoCorrect="off"
            spellCheck="false"
            inputMode="text"
            style={{
              borderColor: usernameError ? '#dc3545' : (username && !usernameError ? '#28a745' : '#e1e5e9')
            }}
          />
          {usernameError && (
            <div style={{ 
              color: '#dc3545', 
              fontSize: '0.875rem', 
              marginTop: '0.25rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.25rem'
            }}>
              ⚠️ {usernameError}
            </div>
          )}
        </div>

        <button
          type="submit"
          className="btn btn-secondary"
          disabled={loading || !username.trim() || !!usernameError}
          style={{ width: '100%' }}
        >
          {loading && loginMode === 'username' ? (
            <>
              <span className="loading"></span>
              Signing in...
            </>
          ) : (
            <>
              <span className="fingerprint-icon">👤</span>
              Sign in with Username
            </>
          )}
        </button>
      </form>

      <div className="divider">
        <span>don't have a passkey?</span>
      </div>

      <button
        onClick={onShowRegister}
        className="btn btn-secondary"
        style={{ width: '100%' }}
        disabled={loading}
      >
        Create New Passkey
      </button>

      <div className="demo-note">
        <strong>Demo Features:</strong>
        <ul style={{ marginTop: '0.5rem', paddingLeft: '1.5rem' }}>
          <li><strong>Passwordless:</strong> Click "Sign in with Passkey" for true passwordless authentication</li>
          <li><strong>Username-based:</strong> Enter username first, then authenticate with passkey</li>
          <li><strong>Multi-device:</strong> Your passkeys work across devices when synced</li>
        </ul>
      </div>
    </div>
  );
}