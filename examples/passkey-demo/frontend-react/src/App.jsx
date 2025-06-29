import { useState, useEffect } from 'react';
import RegisterForm from './components/RegisterForm.jsx';
import LoginForm from './components/LoginForm.jsx';
import Dashboard from './components/Dashboard.jsx';
import Profile from './components/Profile.jsx';

function App() {
  const [currentView, setCurrentView] = useState('login'); // 'login', 'register', 'dashboard', 'profile'
  const [user, setUser] = useState(null);
  const [redirectUrl, setRedirectUrl] = useState(null);
  const [profileUsername, setProfileUsername] = useState(null);

  // Check URL and handle routing
  useEffect(() => {
    const path = window.location.pathname;
    
    // Check if this is a profile URL
    const profileMatch = path.match(/^\/user\/([^\/]+)\/profile$/);
    if (profileMatch) {
      const username = profileMatch[1];
      setProfileUsername(username);
      setCurrentView('profile');
    }
    
    // Store the current path for redirect after login
    if (path !== '/' && path !== '/login' && path !== '/register') {
      setRedirectUrl(path);
    }
  }, []);

  const handleRegistrationSuccess = (result) => {
    console.log('Registration successful:', result);
    setUser({ 
      username: result.data.username,
      displayName: result.data.displayName,
      userId: result.data.userId
    });
    
    // Check if we have a redirect URL
    if (redirectUrl) {
      window.history.pushState({}, '', redirectUrl);
      setCurrentView('profile');
    } else {
      setCurrentView('dashboard');
    }
  };

  const handleLoginSuccess = (result) => {
    console.log('Login successful:', result);
    setUser({ 
      username: result.data.username,
      displayName: result.data.displayName,
      userId: result.data.userId
    });
    
    // Check if we have a redirect URL
    if (redirectUrl) {
      window.history.pushState({}, '', redirectUrl);
      setCurrentView('profile');
    } else {
      setCurrentView('dashboard');
    }
  };

  const handleLogout = () => {
    setUser(null);
    setRedirectUrl(null);
    window.history.pushState({}, '', '/');
    setCurrentView('login');
  };

  const handleAuthRequired = (redirectPath) => {
    console.log('Authentication required for:', redirectPath);
    setRedirectUrl(redirectPath);
    setCurrentView('login');
  };

  const showRegister = () => setCurrentView('register');
  const showLogin = () => setCurrentView('login');
  const showDashboard = () => {
    window.history.pushState({}, '', '/');
    setCurrentView('dashboard');
  };

  return (
    <div className="container">
      {currentView === 'register' && (
        <>
          <RegisterForm onSuccess={handleRegistrationSuccess} />
          <div className="card">
            <button 
              onClick={showLogin} 
              className="btn btn-secondary"
              style={{ width: '100%' }}
            >
              ← Back to Sign In
            </button>
          </div>
        </>
      )}
      
      {currentView === 'login' && (
        <>
          {redirectUrl && (
            <div className="card" style={{ marginBottom: '1rem' }}>
              <div style={{ 
                padding: '1rem', 
                background: '#fff3cd', 
                border: '1px solid #ffeaa7',
                borderRadius: '8px'
              }}>
                <h4 style={{ margin: '0 0 0.5rem 0', color: '#856404' }}>
                  🔐 Authentication Required
                </h4>
                <p style={{ margin: 0, color: '#856404' }}>
                  You need to sign in to access: <strong>{redirectUrl}</strong>
                </p>
                <p style={{ margin: '0.5rem 0 0 0', fontSize: '0.875rem', color: '#856404' }}>
                  After signing in, you'll be automatically redirected to your requested page.
                </p>
              </div>
            </div>
          )}
          <LoginForm 
            onSuccess={handleLoginSuccess} 
            onShowRegister={showRegister}
            redirectUrl={redirectUrl}
          />
        </>
      )}
      
      {currentView === 'dashboard' && user && (
        <>
          <Dashboard 
            user={user} 
            onLogout={handleLogout}
          />
          <div className="card" style={{ marginTop: '1rem' }}>
            <h3>Try Deep Link Authentication</h3>
            <p>Visit your profile page to test protected routes:</p>
            <a 
              href={`/user/${user.username}/profile`}
              onClick={(e) => {
                e.preventDefault();
                window.history.pushState({}, '', `/user/${user.username}/profile`);
                setProfileUsername(user.username);
                setCurrentView('profile');
              }}
              style={{ 
                display: 'inline-block',
                marginTop: '0.5rem',
                color: '#1976d2',
                textDecoration: 'underline'
              }}
            >
              View My Profile →
            </a>
          </div>
        </>
      )}
      
      {currentView === 'profile' && profileUsername && (
        <>
          <Profile 
            username={profileUsername} 
            onAuthRequired={handleAuthRequired}
          />
          {user && (
            <div className="card" style={{ marginTop: '1rem' }}>
              <button 
                onClick={showDashboard}
                className="btn btn-secondary"
                style={{ width: '100%' }}
              >
                ← Back to Dashboard
              </button>
            </div>
          )}
        </>
      )}

      {/* Footer with demo info */}
      <div style={{ 
        textAlign: 'center', 
        color: 'rgba(255, 255, 255, 0.8)', 
        fontSize: '0.875rem',
        marginTop: '2rem'
      }}>
        <p>
          🚀 WebAuthn Passkey Demo • Built with React 19 & Go
        </p>
        <p>
          This demo showcases passwordless authentication using WebAuthn passkeys
        </p>
      </div>
    </div>
  );
}

export default App;