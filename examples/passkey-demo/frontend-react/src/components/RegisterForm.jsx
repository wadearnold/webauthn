import { useState } from 'react';
import { useWebAuthn } from '../hooks/useWebAuthn.js';

export default function RegisterForm({ onSuccess }) {
  const [username, setUsername] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [usernameError, setUsernameError] = useState('');
  const { loading, error, clearError, register, isSupported } = useWebAuthn();

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
    const validationError = validateUsername(username);
    setUsernameError(validationError);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validate username before submitting
    const validationError = validateUsername(username);
    if (validationError) {
      setUsernameError(validationError);
      return;
    }

    try {
      const result = await register(username.trim(), displayName.trim() || username.trim());
      onSuccess?.(result);
    } catch (err) {
      // Error is handled by the hook
      console.error('Registration failed:', err);
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
            <p>Your browser doesn't support WebAuthn. Please use a modern browser like Chrome, Firefox, Safari, or Edge.</p>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="header">
        <h1>🔐 Create Your Passkey</h1>
        <p>Register with a username to create your first passkey</p>
      </div>

      <div className="demo-note">
        <strong>Demo Note:</strong> This is a demonstration of WebAuthn passkeys. 
        Your credentials are stored only in memory and will be lost when the server restarts.
      </div>

      {error && (
        <div className="error">
          <strong>Registration Failed</strong>
          <p>{error}</p>
          <button onClick={clearError} className="btn btn-secondary btn-small">
            Try Again
          </button>
        </div>
      )}

      <form onSubmit={handleSubmit} className="form">
        <div className="form-group">
          <label htmlFor="username">Username *</label>
          <input
            id="username"
            type="text"
            value={username}
            onChange={handleUsernameChange}
            onBlur={handleUsernameBlur}
            placeholder="3-30 chars: letters, numbers, . _ -"
            required
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
          {username && !usernameError && (
            <div style={{ 
              color: '#28a745', 
              fontSize: '0.875rem', 
              marginTop: '0.25rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.25rem'
            }}>
              ✅ Username looks good!
            </div>
          )}
        </div>

        <div className="form-group">
          <label htmlFor="displayName">Display Name</label>
          <input
            id="displayName"
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="Your full name (optional)"
            autoComplete="name"
            disabled={loading}
          />
        </div>

        <button
          type="submit"
          className="btn btn-primary"
          disabled={loading || !username.trim() || !!usernameError}
        >
          {loading ? (
            <>
              <span className="loading"></span>
              Creating Passkey...
            </>
          ) : (
            <>
              <span className="fingerprint-icon">👤</span>
              Create Passkey
            </>
          )}
        </button>
      </form>

      <div className="demo-note">
        <strong>Username Requirements:</strong>
        <ul style={{ marginTop: '0.5rem', paddingLeft: '1.5rem', marginBottom: '1rem' }}>
          <li>3-30 characters long</li>
          <li>Letters, numbers, dots (.), hyphens (-), and underscores (_) only</li>
          <li>Cannot start or end with dots, hyphens, or underscores</li>
          <li>Examples: <code>john.doe</code>, <code>user_123</code>, <code>test-user</code></li>
        </ul>
        
        <strong>What happens next?</strong>
        <ul style={{ marginTop: '0.5rem', paddingLeft: '1.5rem' }}>
          <li>Your browser will prompt you to create a passkey</li>
          <li>Use your device's biometrics, PIN, or security key</li>
          <li>Your passkey will be saved for future logins</li>
        </ul>
      </div>
    </div>
  );
}