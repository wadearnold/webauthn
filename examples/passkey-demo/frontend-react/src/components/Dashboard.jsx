import { useState, useEffect } from 'react';
import * as api from '../services/api.js';

export default function Dashboard({ user, onLogout }) {
  const [passkeys, setPasskeys] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const serverName = window.location.hostname;

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

  const handleDeletePasskey = async (credentialId, passkeyName) => {
    const confirmMessage = `Delete "${passkeyName}"?\n\n⚠️ IMPORTANT: This only removes the passkey from this demo server.\n\nTo fully remove it from your device keychain, please consult your system's documentation for managing saved passwords and passkeys.\n\nContinue with deletion?`;
    
    if (!confirm(confirmMessage)) {
      return;
    }

    // Optimistically remove from UI first
    const originalPasskeys = [...passkeys];
    setPasskeys(passkeys.filter(pk => pk.id !== credentialId));
    setError(null);

    try {
      // Encode the credential ID for URL safety
      const encodedCredentialId = encodeURIComponent(credentialId);
      await api.deletePasskey(encodedCredentialId);
      
      // Success - the optimistic update already happened
      console.log(`Successfully deleted passkey: ${passkeyName}`);
      
    } catch (err) {
      // Revert the optimistic update on error
      setPasskeys(originalPasskeys);
      console.error('Failed to delete passkey:', err);
      
      // Show user-friendly error message
      if (err.message.includes('Not authenticated')) {
        setError('Session expired. Please sign in again to manage passkeys.');
      } else if (err.message.includes('not found')) {
        setError('Passkey not found. It may have already been deleted.');
        // Don't revert in this case - it's already gone
        setPasskeys(passkeys.filter(pk => pk.id !== credentialId));
      } else {
        setError(`Failed to delete passkey: ${err.message}`);
      }
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
            Signed in with passkey authentication • Server: <code style={{ background: '#e9ecef', padding: '0.125rem 0.25rem', borderRadius: '3px' }}>{serverName}</code>
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
          <strong>🔑 No passkeys registered</strong>
          <p>You don't have any passkeys associated with this account yet.</p>
          <p style={{ marginTop: '0.5rem' }}>
            <strong>For demo purposes:</strong> Since you're already signed in, you likely have passkeys 
            in your device keychain. Try signing out and signing back in with the "Sign in with Passkey" 
            button to re-register, or create a new account with a different username.
          </p>
          <div style={{ marginTop: '1rem' }}>
            <button onClick={handleLogout} className="btn btn-primary">
              Sign Out and Try Again
            </button>
          </div>
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
                  onClick={() => handleDeletePasskey(passkey.id, passkey.name)}
                  className="btn btn-danger btn-small"
                  title="Delete this passkey from server (will remain in device keychain)"
                >
                  🗑️ Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="demo-note">
        <strong>⚠️ Important: Passkey Deletion Behavior</strong>
        <p style={{ margin: '0.5rem 0', color: '#d32f2f', fontWeight: '600' }}>
          Deleting a passkey here only removes it from this demo server. The passkey remains in your device's keychain and may still appear during authentication prompts.
        </p>
        
        <p style={{ margin: '0.5rem 0' }}>
          <strong>To fully remove passkeys from your device:</strong> Please consult your operating system's documentation for managing saved passwords and passkeys. The exact steps vary by system version and may change over time.
        </p>

        <details style={{ marginTop: '1rem' }}>
          <summary style={{ cursor: 'pointer', fontWeight: '600' }}>📖 Passkey Technical Information</summary>
          <ul style={{ marginTop: '0.5rem', paddingLeft: '1.5rem', fontSize: '0.875rem' }}>
            <li><strong>User Info:</strong> Display name and username associated with each passkey</li>
            <li><strong>Backup Status (☁️):</strong> Indicates if passkey is synced to cloud keychain</li>
            <li><strong>User Verification (✅):</strong> Shows if biometric/PIN verification is enabled</li>
            <li><strong>Transport:</strong> How the passkey communicates (internal, USB, NFC, Bluetooth, hybrid)</li>
            <li><strong>Attachment:</strong> Platform (built-in) vs cross-platform (external) authenticator</li>
            <li><strong>Attestation:</strong> Cryptographic proof of authenticator authenticity</li>
            <li><strong>Sign Count:</strong> Counter that helps detect cloned authenticators</li>
            <li><strong>AAGUID:</strong> Authenticator model identifier for device recognition</li>
          </ul>
        </details>
      </div>

      <div style={{ marginTop: '1.5rem', padding: '1rem', background: '#f8f9fa', borderRadius: '8px' }}>
        <h4 style={{ margin: '0 0 0.5rem 0' }}>Try These Demo Features:</h4>
        <ul style={{ margin: 0, paddingLeft: '1.5rem' }}>
          <li>Sign out and sign back in with the "Sign in with Passkey" button</li>
          <li>Try creating additional passkeys by registering again</li>
          <li>Test the username-based sign in flow</li>
          <li>Delete passkeys and see them removed from the list</li>
          <li>
            <strong>Test Deep Link Authentication:</strong> Copy your profile URL, sign out, 
            then paste it in a new tab to see authentication redirect
          </li>
        </ul>
      </div>
    </div>
  );
}