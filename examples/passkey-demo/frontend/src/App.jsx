import { useState, useEffect } from 'react';
import RegisterForm from './components/RegisterForm.jsx';
import LoginForm from './components/LoginForm.jsx';
import Dashboard from './components/Dashboard.jsx';

function App() {
  const [currentView, setCurrentView] = useState('login'); // 'login', 'register', 'dashboard'
  const [user, setUser] = useState(null);

  // Check if user is already logged in on app start
  useEffect(() => {
    // In a real app, you might check for a valid session here
    // For this demo, we'll start fresh each time
  }, []);

  const handleRegistrationSuccess = (result) => {
    console.log('Registration successful:', result);
    setUser({ username: result.data.username });
    setCurrentView('dashboard');
  };

  const handleLoginSuccess = (result) => {
    console.log('Login successful:', result);
    setUser({ username: result.data.username });
    setCurrentView('dashboard');
  };

  const handleLogout = () => {
    setUser(null);
    setCurrentView('login');
  };

  const showRegister = () => setCurrentView('register');
  const showLogin = () => setCurrentView('login');

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
        <LoginForm 
          onSuccess={handleLoginSuccess} 
          onShowRegister={showRegister}
        />
      )}
      
      {currentView === 'dashboard' && user && (
        <Dashboard 
          user={user} 
          onLogout={handleLogout}
        />
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