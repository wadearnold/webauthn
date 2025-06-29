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
          <h3>Welcome, {user.displayName || user.username}!</h3>
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
                  {passkey.userVerified && <span style={{ marginLeft: '0.5rem' }}>✅</span>}
                </h4>
                
                <div style={{ margin: '0.5rem 0', padding: '0.5rem', background: '#f8f9fa', borderRadius: '4px', fontSize: '0.875rem' }}>
                  <p style={{ margin: '0 0 0.25rem 0', fontWeight: '600', color: '#333' }}>
                    👤 User: {passkey.displayName || passkey.username} ({passkey.username})
                  </p>
                  <p style={{ margin: 0, color: '#666' }}>
                    Created: {formatDate(passkey.createdAt)} • 
                    Last used: {formatDate(passkey.lastUsed)}
                  </p>
                </div>

                <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.5rem' }}>
                  <p style={{ margin: '0.125rem 0' }}>
                    <strong>Transport:</strong> {passkey.transports.join(', ')}
                    {passkey.authenticatorAttachment && (
                      <> • <strong>Attachment:</strong> {passkey.authenticatorAttachment}</>
                    )}
                  </p>
                  <p style={{ margin: '0.125rem 0' }}>
                    <strong>Attestation:</strong> {passkey.attestationType || 'none'}
                    {passkey.signCount > 0 && (
                      <> • <strong>Sign Count:</strong> {passkey.signCount}</>
                    )}
                  </p>
                  {passkey.aaguid && (
                    <p style={{ margin: '0.125rem 0' }}>
                      <strong>AAGUID:</strong> {passkey.aaguid}
                    </p>
                  )}
                </div>

                <div style={{ marginTop: '0.5rem', fontSize: '0.75rem' }}>
                  {passkey.backedUp && (
                    <span style={{ color: '#28a745', marginRight: '1rem' }}>
                      ✓ Backed up and synced
                    </span>
                  )}
                  {passkey.backupEligible && !passkey.backedUp && (
                    <span style={{ color: '#ffc107', marginRight: '1rem' }}>
                      ⚠️ Backup eligible but not backed up
                    </span>
                  )}
                  {passkey.userVerified && (
                    <span style={{ color: '#17a2b8', marginRight: '1rem' }}>
                      🔐 User verification enabled
                    </span>
                  )}
                </div>
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
        <strong>Passkey Information Guide:</strong>
        <ul style={{ marginTop: '0.5rem', paddingLeft: '1.5rem' }}>
          <li><strong>User Info:</strong> Shows the display name and username associated with each passkey</li>
          <li><strong>Backup Status (☁️):</strong> Indicates if passkey is synced to cloud keychain (iCloud, Google, etc.)</li>
          <li><strong>User Verification (✅):</strong> Shows if biometric/PIN verification is enabled</li>
          <li><strong>Transport:</strong> How the passkey communicates (internal, USB, NFC, Bluetooth, hybrid)</li>
          <li><strong>Attachment:</strong> Platform (built-in) vs cross-platform (external) authenticator</li>
          <li><strong>Attestation:</strong> Cryptographic proof of authenticator authenticity</li>
          <li><strong>Sign Count:</strong> Counter that helps detect cloned authenticators</li>
          <li><strong>AAGUID:</strong> Authenticator model identifier for device recognition</li>
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