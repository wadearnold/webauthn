import { useState, useEffect } from 'react';
import * as api from '../services/api.js';

export default function Profile({ username, onAuthRequired }) {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadProfile();
  }, [username]);

  const loadProfile = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const profileData = await api.getUserProfile(username);
      setProfile(profileData);
    } catch (err) {
      if (err.code === 'AUTH_REQUIRED') {
        // Need authentication - trigger redirect with deep link
        onAuthRequired?.(err.redirectUrl);
      } else {
        setError(err.message);
      }
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="card">
        <div style={{ textAlign: 'center', padding: '2rem' }}>
          <span className="loading" style={{ width: '2rem', height: '2rem' }}></span>
          <p style={{ marginTop: '1rem', color: '#666' }}>Loading profile...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="card">
        <div className="error">
          <strong>Error Loading Profile</strong>
          <p>{error}</p>
        </div>
      </div>
    );
  }

  if (!profile) {
    return null;
  }

  return (
    <div className="card">
      <div className="header">
        <h2>🔐 Protected Profile Page</h2>
        <p style={{ color: '#666', marginTop: '0.5rem' }}>
          This page requires authentication to access
        </p>
      </div>

      <div style={{ 
        padding: '1.5rem', 
        background: '#f8f9fa', 
        borderRadius: '8px',
        marginBottom: '1.5rem'
      }}>
        <h3 style={{ margin: '0 0 1rem 0', color: '#333' }}>
          👤 {profile.displayName || profile.username}'s Profile
        </h3>
        
        <div style={{ display: 'grid', gap: '0.75rem' }}>
          <div>
            <strong style={{ color: '#666' }}>Username:</strong>
            <span style={{ marginLeft: '0.5rem' }}>{profile.username}</span>
          </div>
          
          <div>
            <strong style={{ color: '#666' }}>Display Name:</strong>
            <span style={{ marginLeft: '0.5rem' }}>{profile.displayName || 'Not set'}</span>
          </div>
          
          <div>
            <strong style={{ color: '#666' }}>Member Since:</strong>
            <span style={{ marginLeft: '0.5rem' }}>
              {new Date(profile.createdAt).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
              })}
            </span>
          </div>
          
          <div>
            <strong style={{ color: '#666' }}>Passkeys Registered:</strong>
            <span style={{ marginLeft: '0.5rem' }}>{profile.passkeyCount}</span>
          </div>
          
          <div>
            <strong style={{ color: '#666' }}>Backup-Eligible Passkeys:</strong>
            <span style={{ marginLeft: '0.5rem' }}>
              {profile.hasBackupEligiblePasskeys ? '✅ Yes' : '❌ No'}
            </span>
          </div>
        </div>
      </div>

      <div className="demo-note">
        <strong>🎯 Deep Link Authentication Demo</strong>
        <p style={{ marginTop: '0.5rem' }}>
          This profile page demonstrates protected routes that require authentication.
          When you try to access this page without being logged in, you'll be redirected
          to the login screen. After successful authentication, you'll be automatically
          redirected back to this profile page.
        </p>
        
        <div style={{ marginTop: '1rem' }}>
          <strong>Try this flow:</strong>
          <ol style={{ marginTop: '0.5rem', paddingLeft: '1.5rem' }}>
            <li>Copy this profile URL</li>
            <li>Sign out from the dashboard</li>
            <li>Paste the URL in a new browser tab</li>
            <li>You'll be prompted to authenticate</li>
            <li>After login, you'll return here automatically</li>
          </ol>
        </div>
      </div>

      <div style={{ 
        marginTop: '1.5rem', 
        padding: '1rem', 
        background: '#e3f2fd', 
        borderRadius: '8px',
        border: '1px solid #90caf9'
      }}>
        <h4 style={{ margin: '0 0 0.5rem 0', color: '#1976d2' }}>
          🔗 Share Your Profile Link
        </h4>
        <div style={{ 
          display: 'flex', 
          alignItems: 'center', 
          gap: '0.5rem',
          marginTop: '0.5rem'
        }}>
          <input
            type="text"
            value={`${window.location.origin}/user/${username}/profile`}
            readOnly
            style={{ 
              flex: 1,
              padding: '0.5rem',
              border: '1px solid #ddd',
              borderRadius: '4px',
              background: '#fff'
            }}
          />
          <button
            onClick={() => {
              navigator.clipboard.writeText(`${window.location.origin}/user/${username}/profile`);
              alert('Profile link copied to clipboard!');
            }}
            className="btn btn-secondary btn-small"
          >
            📋 Copy
          </button>
        </div>
      </div>
    </div>
  );
}