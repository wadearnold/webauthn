import { useState } from 'react';
import { useWebAuthn } from '../hooks/useWebAuthn.js';

export default function LoginForm({ onSuccess, onShowRegister }) {
  const [username, setUsername] = useState('');
  const [loginMode, setLoginMode] = useState('discoverable'); // 'discoverable' or 'username'
  const { loading, error, clearError, authenticate, isSupported } = useWebAuthn();

  const handleUsernameLogin = async (e) => {
    e.preventDefault();
    
    if (!username.trim()) {
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
    return (
      <div className="card">
        <div className="error">
          <strong>WebAuthn Not Supported</strong>
          <p>Your browser doesn't support WebAuthn. Please use a modern browser.</p>
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
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Enter your username"
            autoComplete="username"
            disabled={loading}
          />
        </div>

        <button
          type="submit"
          className="btn btn-secondary"
          disabled={loading || !username.trim()}
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