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
    // Basic WebAuthn API check
    const hasWebAuthnAPI = !!(navigator.credentials && navigator.credentials.create && navigator.credentials.get);
    
    if (!hasWebAuthnAPI) {
      console.error('WebAuthn API not available in this browser');
      return false;
    }
    
    // Check for secure context requirement
    const isSecureContext = window.isSecureContext;
    const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    const isCustomLocalDomain = window.location.hostname.endsWith('.local');
    
    console.log('WebAuthn Support Check:', {
      hasWebAuthnAPI,
      isSecureContext,
      isLocalhost,
      isCustomLocalDomain,
      hostname: window.location.hostname,
      protocol: window.location.protocol
    });
    
    // WebAuthn requires secure context (HTTPS) except for localhost
    if (!isSecureContext && !isLocalhost) {
      console.error('WebAuthn requires HTTPS or localhost. Current origin:', window.location.origin);
      return false;
    }
    
    return true;
  }, []);

  // Register a new passkey
  const register = useCallback(async (username, displayName) => {
    if (!isSupported()) {
      throw new Error('WebAuthn is not supported in this browser');
    }

    setLoading(true);
    setError(null);

    try {
      console.log('=== WEBAUTHN REGISTRATION DEBUG START ===');
      console.log('Username:', username, 'DisplayName:', displayName);
      
      // Begin registration
      const options = await api.registerBegin(username, displayName);
      console.log('Server options received:', options);
      
      // Parse options for WebAuthn API
      const parsedOptions = parseCredentialCreationOptions(options.publicKey);
      console.log('Parsed options for browser:', parsedOptions);
      
      // Comprehensive biometric availability check
      console.log('=== BIOMETRIC AVAILABILITY CHECK ===');
      if (window.PublicKeyCredential) {
        console.log('PublicKeyCredential available:', true);
        
        // Check platform authenticator availability
        const platformAvailable = await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
        console.log('Platform authenticator available:', platformAvailable);
        
        // Check conditional UI support (Level 3)
        if (PublicKeyCredential.isConditionalMediationAvailable) {
          const conditionalAvailable = await PublicKeyCredential.isConditionalMediationAvailable();
          console.log('Conditional mediation available:', conditionalAvailable);
        }
      } else {
        console.log('PublicKeyCredential NOT available');
      }
      
      // Log user agent for debugging
      console.log('User Agent:', navigator.userAgent);
      console.log('Device info:', {
        platform: navigator.platform,
        maxTouchPoints: navigator.maxTouchPoints,
        cookieEnabled: navigator.cookieEnabled
      });
      
      console.log('=== FINAL WEBAUTHN OPTIONS ===');
      console.log('Challenge length:', parsedOptions.challenge.byteLength);
      console.log('RPID:', parsedOptions.rp.id);
      console.log('RP Name:', parsedOptions.rp.name);
      console.log('User ID length:', parsedOptions.user.id.byteLength);
      console.log('User name:', parsedOptions.user.name);
      console.log('User display name:', parsedOptions.user.displayName);
      console.log('Authenticator Selection:', parsedOptions.authenticatorSelection);
      console.log('Attestation:', parsedOptions.attestation);
      console.log('Timeout:', parsedOptions.timeout);
      console.log('Credential Parameters:', parsedOptions.pubKeyCredParams);
      
      console.log('=== CALLING navigator.credentials.create() ===');
      const startTime = performance.now();
      
      const credential = await navigator.credentials.create({
        publicKey: parsedOptions,
      });
      
      const endTime = performance.now();
      console.log(`Credential creation took ${endTime - startTime} milliseconds`);

      if (!credential) {
        console.error('Credential creation returned null');
        throw new Error('Failed to create credential');
      }
      
      console.log('=== CREDENTIAL CREATED SUCCESSFULLY ===');
      console.log('Credential ID:', credential.id);
      console.log('Credential type:', credential.type);
      console.log('Response type:', credential.response.constructor.name);
      console.log('Client data JSON:', new TextDecoder().decode(credential.response.clientDataJSON));
      
      // Log authenticator data details
      const attestationObject = credential.response.attestationObject;
      console.log('Attestation object length:', attestationObject.byteLength);
      
      // Check for transports
      if (credential.response.getTransports) {
        const transports = credential.response.getTransports();
        console.log('Available transports:', transports);
      }

      // Format response for server
      const formattedResponse = formatCredentialCreationResponse(credential);
      console.log('Formatted response for server:', formattedResponse);
      
      // Finish registration
      console.log('=== SENDING TO SERVER ===');
      const result = await api.registerFinish(formattedResponse);
      console.log('Registration result:', result);
      
      console.log('=== WEBAUTHN REGISTRATION DEBUG END ===');
      return result;
    } catch (err) {
      console.error('=== WEBAUTHN REGISTRATION ERROR ===');
      console.error('Error name:', err.name);
      console.error('Error message:', err.message);
      console.error('Error stack:', err.stack);
      console.error('Full error object:', err);
      
      // Log specific WebAuthn error types
      if (err.name === 'NotAllowedError') {
        console.error('BIOMETRIC ISSUE: User denied permission or biometric authentication failed');
        console.error('Common causes:');
        console.error('- User cancelled the biometric prompt');
        console.error('- Biometric sensor not working');
        console.error('- System falling back to password');
        console.error('- Browser policy restrictions');
      } else if (err.name === 'InvalidStateError') {
        console.error('CREDENTIAL ISSUE: Credential may already exist');
      } else if (err.name === 'SecurityError') {
        console.error('SECURITY ISSUE: Origin or RPID mismatch');
      } else if (err.name === 'NotSupportedError') {
        console.error('SUPPORT ISSUE: WebAuthn feature not supported');
      }
      
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
      console.log('=== WEBAUTHN AUTHENTICATION DEBUG START ===');
      console.log('Authentication type:', username ? 'Traditional (with username)' : 'Discoverable (passwordless)');
      console.log('Username:', username);
      
      // Begin authentication
      const options = await api.loginBegin(username);
      console.log('Server options received:', options);
      
      // Parse options for WebAuthn API
      const parsedOptions = parseCredentialRequestOptions(options.publicKey);
      console.log('Parsed options for browser:', parsedOptions);
      
      console.log('=== AUTHENTICATION OPTIONS ===');
      console.log('Challenge length:', parsedOptions.challenge.byteLength);
      console.log('RPID:', parsedOptions.rpId);
      console.log('User Verification:', parsedOptions.userVerification);
      console.log('Timeout:', parsedOptions.timeout);
      console.log('Allow Credentials count:', parsedOptions.allowCredentials ? parsedOptions.allowCredentials.length : 0);
      
      if (parsedOptions.allowCredentials) {
        parsedOptions.allowCredentials.forEach((cred, index) => {
          console.log(`Credential ${index + 1}:`, {
            id: bufferToBase64URLString(cred.id),
            type: cred.type,
            transports: cred.transports
          });
        });
      }
      
      console.log('=== CALLING navigator.credentials.get() ===');
      const startTime = performance.now();
      
      const credential = await navigator.credentials.get({
        publicKey: parsedOptions,
      });
      
      const endTime = performance.now();
      console.log(`Authentication took ${endTime - startTime} milliseconds`);

      if (!credential) {
        console.error('Authentication returned null');
        throw new Error('Authentication was cancelled or failed');
      }
      
      console.log('=== AUTHENTICATION SUCCESSFUL ===');
      console.log('Credential ID:', credential.id);
      console.log('Credential type:', credential.type);
      console.log('Response type:', credential.response.constructor.name);
      console.log('Client data JSON:', new TextDecoder().decode(credential.response.clientDataJSON));
      console.log('User handle:', credential.response.userHandle ? bufferToBase64URLString(credential.response.userHandle) : 'null');
      console.log('Authenticator data length:', credential.response.authenticatorData.byteLength);
      console.log('Signature length:', credential.response.signature.byteLength);

      // Format response for server
      const formattedResponse = formatCredentialGetResponse(credential);
      console.log('Formatted response for server:', formattedResponse);
      
      // Finish authentication
      console.log('=== SENDING TO SERVER ===');
      const result = await api.loginFinish(formattedResponse);
      console.log('Authentication result:', result);
      
      console.log('=== WEBAUTHN AUTHENTICATION DEBUG END ===');
      return result;
    } catch (err) {
      console.error('=== WEBAUTHN AUTHENTICATION ERROR ===');
      console.error('Error name:', err.name);
      console.error('Error message:', err.message);
      console.error('Error stack:', err.stack);
      console.error('Full error object:', err);
      
      // Log specific WebAuthn error types
      if (err.name === 'NotAllowedError') {
        console.error('BIOMETRIC ISSUE: User denied permission or biometric authentication failed');
        console.error('Common causes:');
        console.error('- User cancelled the biometric prompt');
        console.error('- Biometric sensor not working');
        console.error('- No matching credentials found');
        console.error('- System falling back to password');
      } else if (err.name === 'InvalidStateError') {
        console.error('STATE ISSUE: Invalid authentication state');
      } else if (err.name === 'SecurityError') {
        console.error('SECURITY ISSUE: Origin or RPID mismatch');
      } else if (err.name === 'NotSupportedError') {
        console.error('SUPPORT ISSUE: WebAuthn feature not supported');
      }
      
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