const API_BASE = 'http://localhost:8080/api';

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