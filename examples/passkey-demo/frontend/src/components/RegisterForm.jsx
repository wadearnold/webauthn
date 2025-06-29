import { useState } from 'react';
import { useWebAuthn } from '../hooks/useWebAuthn.js';

export default function RegisterForm({ onSuccess }) {
  const [username, setUsername] = useState('');
  const [displayName, setDisplayName] = useState('');
  const { loading, error, clearError, register, isSupported } = useWebAuthn();

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!username.trim()) {
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
    return (
      <div className="card">
        <div className="error">
          <strong>WebAuthn Not Supported</strong>
          <p>Your browser doesn't support WebAuthn. Please use a modern browser like Chrome, Firefox, Safari, or Edge.</p>
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
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Enter your username"
            required
            autoComplete="username"
            disabled={loading}
          />
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
          disabled={loading || !username.trim()}
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