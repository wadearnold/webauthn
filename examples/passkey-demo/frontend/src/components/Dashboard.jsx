import { useState, useEffect } from 'react';
import * as api from '../services/api.js';

export default function Dashboard({ user, onLogout }) {
  const [passkeys, setPasskeys] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadPasskeys();
  }, []);

  const loadPasskeys = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const passkeyList = await api.getUserPasskeys();
      setPasskeys(passkeyList);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeletePasskey = async (credentialId) => {
    if (!confirm('Are you sure you want to delete this passkey?')) {
      return;
    }

    try {
      await api.deletePasskey(credentialId);
      await loadPasskeys(); // Reload the list
    } catch (err) {
      setError(err.message);
    }
  };

  const handleLogout = async () => {
    try {
      await api.logout();
      onLogout?.();
    } catch (err) {
      // Even if logout fails, clear the session locally
      onLogout?.();
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getTransportIcon = (transport) => {
    switch (transport) {
      case 'internal': return '📱';
      case 'usb': return '🔌';
      case 'nfc': return '📡';
      case 'ble': return '🔵';
      case 'hybrid': return '📲';
      default: return '🔑';
    }
  };

  return (
    <div className="card">
      <div className="user-info">
        <div>
          <h3>Welcome, {user.username}!</h3>
          <p style={{ margin: 0, color: '#666', fontSize: '0.875rem' }}>
            Signed in with passkey authentication
          </p>
        </div>
        <button onClick={handleLogout} className="btn btn-secondary btn-small">
          Sign Out
        </button>
      </div>

      <div className="header">
        <h2>Your Passkeys</h2>
        <p>Manage your registered passkeys below</p>
      </div>

      {error && (
        <div className="error">
          <strong>Error</strong>
          <p>{error}</p>
          <button onClick={() => setError(null)} className="btn btn-secondary btn-small">
            Dismiss
          </button>
        </div>
      )}

      {loading ? (
        <div style={{ textAlign: 'center', padding: '2rem' }}>
          <span className="loading" style={{ width: '2rem', height: '2rem' }}></span>
          <p style={{ marginTop: '1rem', color: '#666' }}>Loading passkeys...</p>
        </div>
      ) : passkeys.length === 0 ? (
        <div className="demo-note">
          <strong>No passkeys found</strong>
          <p>This shouldn't happen since you just signed in with a passkey. 
          Try refreshing the page or contact support.</p>
        </div>
      ) : (
        <div className="passkey-list">
          {passkeys.map((passkey) => (
            <div key={passkey.id} className="passkey-item">
              <div className="passkey-info">
                <h4>
                  {getTransportIcon(passkey.transports[0])} {passkey.name}
                  {passkey.backedUp && <span style={{ marginLeft: '0.5rem' }}>☁️</span>}
                </h4>
                <p>
                  Created: {formatDate(passkey.createdAt)} • 
                  Last used: {formatDate(passkey.lastUsed)}
                  {passkey.transports.length > 0 && (
                    <> • Transport: {passkey.transports.join(', ')}</>
                  )}
                </p>
                {passkey.backedUp && (
                  <p style={{ color: '#28a745', fontSize: '0.75rem', margin: '0.25rem 0 0 0' }}>
                    ✓ Backed up and synced across devices
                  </p>
                )}
              </div>
              <div className="passkey-actions">
                <button
                  onClick={() => handleDeletePasskey(passkey.id)}
                  className="btn btn-danger btn-small"
                  title="Delete this passkey"
                >
                  🗑️ Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="demo-note">
        <strong>Demo Information:</strong>
        <ul style={{ marginTop: '0.5rem', paddingLeft: '1.5rem' }}>
          <li><strong>Backed up passkeys (☁️):</strong> Synced to your cloud keychain (iCloud, Google, etc.)</li>
          <li><strong>Transport types:</strong> How the passkey can be used (device biometrics, USB, NFC, etc.)</li>
          <li><strong>Security:</strong> Each passkey is cryptographically unique and cannot be phished</li>
          <li><strong>Deletion:</strong> Removing a passkey here only affects this demo server</li>
        </ul>
      </div>

      <div style={{ marginTop: '1.5rem', padding: '1rem', background: '#f8f9fa', borderRadius: '8px' }}>
        <h4 style={{ margin: '0 0 0.5rem 0' }}>Try These Demo Features:</h4>
        <ul style={{ margin: 0, paddingLeft: '1.5rem' }}>
          <li>Sign out and sign back in with the "Sign in with Passkey" button</li>
          <li>Try creating additional passkeys by registering again</li>
          <li>Test the username-based sign in flow</li>
          <li>Delete passkeys and see them removed from the list</li>
        </ul>
      </div>
    </div>
  );
}