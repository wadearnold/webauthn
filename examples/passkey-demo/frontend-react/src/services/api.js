// ngrok-based WebAuthn cross-platform configuration
// Get API base URL from environment variable or use development default
const getApiBaseUrl = () => {
  // Check for ngrok URL in environment (injected by Vite)
  const ngrokUrl = import.meta.env.VITE_NGROK_URL;
  
  if (ngrokUrl) {
    // Use ngrok tunnel URL
    return `${ngrokUrl}/api`;
  }
  
  // Fallback to localhost for local development
  const isHTTPS = window.location.protocol === 'https:';
  return `${isHTTPS ? 'https' : 'http'}://localhost:8080/api`;
};

const API_BASE = getApiBaseUrl();

// Helper function to handle API responses
async function handleResponse(response) {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Network error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }
  return response.json();
}

// Helper function to make API requests
async function apiRequest(endpoint, options = {}) {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    credentials: 'include', // Include cookies
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });
  
  return handleResponse(response);
}

// Registration API
export const registerBegin = async (username, displayName) => {
  return apiRequest('/register/begin', {
    method: 'POST',
    body: JSON.stringify({ username, displayName }),
  });
};

export const registerFinish = async (credential) => {
  return apiRequest('/register/finish', {
    method: 'POST',
    body: JSON.stringify(credential),
  });
};

// Authentication API
export const loginBegin = async (username = null) => {
  return apiRequest('/login/begin', {
    method: 'POST',
    body: JSON.stringify({ username }),
  });
};

export const loginFinish = async (credential) => {
  return apiRequest('/login/finish', {
    method: 'POST',
    body: JSON.stringify(credential),
  });
};

// User management API
export const getUserPasskeys = async () => {
  return apiRequest('/user/passkeys');
};

export const deletePasskey = async (credentialId) => {
  return apiRequest(`/user/passkeys/${credentialId}`, {
    method: 'DELETE',
  });
};

export const logout = async () => {
  return apiRequest('/logout', {
    method: 'POST',
  });
};

// Protected profile API  
export const getUserProfile = async (username) => {
  try {
    const response = await fetch(`${API_BASE}/user/${username}/profile`, {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (response.status === 401) {
      const data = await response.json();
      // Throw special error for authentication required
      const error = new Error(data.error);
      error.code = data.code;
      error.redirectUrl = data.redirectUrl;
      throw error;
    }

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to fetch profile');
    }

    return response.json();
  } catch (error) {
    throw error;
  }
};