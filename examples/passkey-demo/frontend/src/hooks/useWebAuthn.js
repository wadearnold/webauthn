import { useState, useCallback } from 'react';
import * as api from '../services/api.js';

// Helper functions for WebAuthn API
function bufferToBase64URLString(buffer) {
  const bytes = new Uint8Array(buffer);
  let str = '';
  for (const charCode of bytes) {
    str += String.fromCharCode(charCode);
  }
  const base64String = btoa(str);
  return base64String.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

function base64URLStringToBuffer(base64URLString) {
  const base64 = base64URLString.replace(/-/g, '+').replace(/_/g, '/');
  const padLength = (4 - (base64.length % 4)) % 4;
  const padded = base64.padEnd(base64.length + padLength, '=');
  const binary = atob(padded);
  const buffer = new ArrayBuffer(binary.length);
  const bytes = new Uint8Array(buffer);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return buffer;
}

// Convert server response to WebAuthn API format
function parseCredentialCreationOptions(options) {
  return {
    ...options,
    challenge: base64URLStringToBuffer(options.challenge),
    user: {
      ...options.user,
      id: base64URLStringToBuffer(options.user.id),
    },
    excludeCredentials: options.excludeCredentials?.map(cred => ({
      ...cred,
      id: base64URLStringToBuffer(cred.id),
    })) || [],
  };
}

function parseCredentialRequestOptions(options) {
  return {
    ...options,
    challenge: base64URLStringToBuffer(options.challenge),
    allowCredentials: options.allowCredentials?.map(cred => ({
      ...cred,
      id: base64URLStringToBuffer(cred.id),
    })) || [],
  };
}

// Convert WebAuthn API response to server format
function formatCredentialCreationResponse(credential) {
  return {
    id: credential.id,
    rawId: bufferToBase64URLString(credential.rawId),
    type: credential.type,
    response: {
      attestationObject: bufferToBase64URLString(credential.response.attestationObject),
      clientDataJSON: bufferToBase64URLString(credential.response.clientDataJSON),
      transports: credential.response.getTransports?.() || [],
    },
  };
}

function formatCredentialGetResponse(credential) {
  return {
    id: credential.id,
    rawId: bufferToBase64URLString(credential.rawId),
    type: credential.type,
    response: {
      authenticatorData: bufferToBase64URLString(credential.response.authenticatorData),
      clientDataJSON: bufferToBase64URLString(credential.response.clientDataJSON),
      signature: bufferToBase64URLString(credential.response.signature),
      userHandle: credential.response.userHandle 
        ? bufferToBase64URLString(credential.response.userHandle) 
        : null,
    },
  };
}

export function useWebAuthn() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const clearError = useCallback(() => setError(null), []);

  // Check if WebAuthn is supported
  const isSupported = useCallback(() => {
    return !!(navigator.credentials && navigator.credentials.create && navigator.credentials.get);
  }, []);

  // Register a new passkey
  const register = useCallback(async (username, displayName) => {
    if (!isSupported()) {
      throw new Error('WebAuthn is not supported in this browser');
    }

    setLoading(true);
    setError(null);

    try {
      // Begin registration
      const options = await api.registerBegin(username, displayName);
      
      // Parse options for WebAuthn API
      const parsedOptions = parseCredentialCreationOptions(options.publicKey);
      
      // Create credential
      const credential = await navigator.credentials.create({
        publicKey: parsedOptions,
      });

      if (!credential) {
        throw new Error('Failed to create credential');
      }

      // Format response for server
      const formattedResponse = formatCredentialCreationResponse(credential);
      
      // Finish registration
      const result = await api.registerFinish(formattedResponse);
      
      return result;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [isSupported]);

  // Authenticate with passkey
  const authenticate = useCallback(async (username = null) => {
    if (!isSupported()) {
      throw new Error('WebAuthn is not supported in this browser');
    }

    setLoading(true);
    setError(null);

    try {
      // Begin authentication
      const options = await api.loginBegin(username);
      
      // Parse options for WebAuthn API
      const parsedOptions = parseCredentialRequestOptions(options.publicKey);
      
      // Get credential
      const credential = await navigator.credentials.get({
        publicKey: parsedOptions,
      });

      if (!credential) {
        throw new Error('Authentication was cancelled or failed');
      }

      // Format response for server
      const formattedResponse = formatCredentialGetResponse(credential);
      
      // Finish authentication
      const result = await api.loginFinish(formattedResponse);
      
      return result;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [isSupported]);

  return {
    loading,
    error,
    clearError,
    isSupported,
    register,
    authenticate,
  };
}