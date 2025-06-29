# WebAuthn Passkey Demo - React Frontend

A React 19 web frontend demonstrating WebAuthn passkey authentication with deep link support. Part of a multi-platform passkey demo showcasing cross-platform compatibility.

## 🚨 **REQUIRED**: Domain Setup First

**This frontend uses `passkey-demo.local` for cross-platform WebAuthn compatibility. You MUST configure local domain resolution:**

```bash
# Add to /etc/hosts (requires sudo)
sudo vim /etc/hosts

# Add this line:
127.0.0.1 passkey-demo.local
```

## 🚀 Quick Start

```bash
npm install
npm run dev
```

The custom dev script displays cross-platform configuration info, setup reminders, and starts Vite with helpful guidance.

**Access the demo at**: [http://passkey-demo.local:5173](http://passkey-demo.local:5173)

⚠️ **Important**: Use `passkey-demo.local:5173`, NOT `localhost:5173` for proper cross-platform passkey functionality.

### Expected Output

```bash
$ npm run dev

🌐 Cross-Platform WebAuthn Passkey Demo
========================================
🔐 React Frontend: http://passkey-demo.local:5173
📡 Backend API: http://passkey-demo.local:8080

⚠️  CRITICAL SETUP REQUIRED:
   Add to /etc/hosts (requires sudo):
   127.0.0.1 passkey-demo.local

🔗 For detailed setup: See README.md
🚀 Starting Vite development server...
   Note: Vite shows localhost URLs, but use passkey-demo.local instead

  VITE v7.0.0  ready in 128 ms
  ➜  Local:   http://localhost:5173/  ← Ignore this
  ➜  Network: http://10.0.0.202:5173/ ← Ignore this
  ➜  USE THIS: http://passkey-demo.local:5173/
```

### Alternative Commands

```bash
# Start with domain configuration info (recommended)
npm run dev

# Start Vite directly (skip custom messaging)
npm run dev:direct
```

## 🔧 Features

- **Passwordless Authentication**: True passwordless login using WebAuthn passkeys
- **Discoverable Credentials**: Sign in without entering a username
- **Deep Link Authentication**: Protected routes with automatic redirect after login
- **Real-time Validation**: Client-side username validation with visual feedback
- **Comprehensive UI**: Detailed passkey information and management

## 🔍 Troubleshooting Biometric Authentication

If passkeys are requesting your **system password** instead of **biometrics** (Touch ID, Face ID, fingerprint), the issue is typically system-level configuration, not the demo code.

### **Step 1: Check Browser Console**

Open Developer Tools (F12) → Console tab and look for these logs during registration:

```javascript
// Should show proper WebAuthn configuration
Enhanced WebAuthn Options: {
  authenticatorSelection: {
    authenticatorAttachment: 'platform',
    userVerification: 'required',
    requireResidentKey: true
  }
}

// Critical: Check if biometric authenticator is available
Platform authenticator available: true/false
```

### **Step 2: System Settings Check**

#### **macOS (Chrome/Safari)**
1. **System Settings → Touch ID & Passcode**
   - ✅ Ensure "Use Touch ID for Safari" is enabled
   - ✅ Test Touch ID works in other apps
   
2. **Chrome Settings**
   - Navigate to `chrome://settings/content/securityKeys`
   - ✅ Enable "Allow sites to manage security keys"
   
3. **Safari Alternative**
   - Try the demo in Safari browser
   - Safari often has better Touch ID integration on macOS

#### **iOS (Safari/Chrome)**
1. **Settings → Face ID & Passcode**
   - ✅ Enable "Safari" under "Use Face ID For"
   
2. **Safari Settings → AutoFill**
   - ✅ Enable "Passkeys"

#### **Android (Chrome)**
1. **Settings → Passwords & accounts → Google**
   - ✅ Enable "Passkeys"
   
2. **Chrome Settings → Passwords**
   - ✅ Enable "Offer to save passwords"

#### **Windows (Chrome/Edge)**
1. **Settings → Accounts → Sign-in options**
   - ✅ Set up Windows Hello (fingerprint, face, or PIN)
   
2. **Chrome/Edge Settings**
   - ✅ Enable security key/passkey features

### **Step 3: Debugging Different Scenarios**

#### **Scenario A: `Platform authenticator available: false`**
**Problem**: System not exposing biometric authenticator to browser
**Solutions**:
- Check system biometric settings above
- Restart browser after enabling settings
- Try different browser (Safari on macOS, Edge on Windows)
- Reboot system if recently enabled biometrics

#### **Scenario B: `Platform authenticator available: true` but still uses password**
**Problem**: Browser policy falling back to password
**Solutions**:
- Clear browser data for localhost
- Try incognito/private browsing mode
- Check if Touch ID sensor needs cleaning
- Verify no recent failed biometric attempts

#### **Scenario C: Works sometimes, not others**
**Problem**: Inconsistent biometric availability
**Common causes**:
- Touch ID sensor dirty/wet
- Multiple failed attempts triggered fallback
- System under heavy load
- Browser background/focus issues

### **Step 4: Browser-Specific Testing**

Try the demo in multiple browsers to isolate the issue:

| Browser | Expected Behavior |
|---------|------------------|
| **Safari (macOS)** | Best Touch ID integration |
| **Chrome (macOS)** | Good, but may require settings |
| **Firefox (macOS)** | Limited WebAuthn support |
| **Chrome (Android)** | Good fingerprint integration |
| **Safari (iOS)** | Excellent Face ID integration |
| **Edge (Windows)** | Good Windows Hello integration |

### **Step 5: Alternative Testing**

If biometrics still don't work, test with other WebAuthn demos:
- [webauthn.io](https://webauthn.io)
- [passkeys.dev](https://passkeys.dev)

If those also use system password, the issue is definitely system configuration, not this demo.

### **Step 6: Common System Password Triggers**

The system may request password instead of biometrics when:
- ✋ **Security Policy**: System requires password confirmation for new credentials
- ✋ **Failed Attempts**: Multiple failed biometric attempts triggered fallback
- ✋ **Sensor Issues**: Biometric sensor not working properly
- ✋ **First Setup**: Initial passkey setup requires password confirmation
- ✋ **System Load**: High CPU/memory usage affecting biometric processing

## 🎯 Expected Behavior

When properly configured, you should see:

### **Registration Flow**
1. Click "Create Passkey"
2. Browser shows native biometric prompt (Touch ID, Face ID, etc.)
3. Complete biometric authentication
4. Passkey created and user logged in

### **Login Flow**
1. Click "Sign in with Passkey"
2. Browser shows available passkeys
3. Select your passkey
4. Authenticate with biometrics
5. Signed in successfully

## 🔧 Development

### **Project Structure**
```
src/
├── components/          # React components
│   ├── Dashboard.jsx    # User dashboard with passkey management
│   ├── LoginForm.jsx    # Authentication forms
│   ├── Profile.jsx      # Protected profile page
│   └── RegisterForm.jsx # Registration with validation
├── hooks/
│   └── useWebAuthn.js   # WebAuthn API integration
├── services/
│   └── api.js          # Backend API client
└── App.jsx             # Main app with routing
```

### **Key Files**
- **`useWebAuthn.js`**: Core WebAuthn implementation
- **`api.js`**: Communication with Go backend
- **`Profile.jsx`**: Demonstrates protected routes
- **`Dashboard.jsx`**: Passkey management interface

### **Environment Variables**
The frontend expects the backend at `http://localhost:8080`. Modify `API_BASE` in `src/services/api.js` if using a different backend URL.

## 🐛 Common Issues

### **CORS Errors**
Ensure backend is running on port 8080 and frontend on 5173. The backend has CORS configured for `http://localhost:5173`.

### **WebAuthn Not Supported**
The demo requires a modern browser with WebAuthn support:
- Chrome 67+
- Firefox 60+
- Safari 14+
- Edge 18+

### **Registration Fails**
Check browser console for detailed error messages. Common issues:
- Username validation (client/server mismatch)
- WebAuthn API errors
- Session timeout (registration has 5-minute limit)

### **Profile Page 401 Errors**
This is expected behavior when accessing protected routes without authentication. The demo will redirect to login and back to the requested page.

## 📚 Learn More

- [WebAuthn Guide](https://webauthn.guide/)
- [Passkeys Documentation](https://passkeys.dev/)
- [React 19 Documentation](https://react.dev/)
- [FIDO Alliance](https://fidoalliance.org/)

---

**💡 Tip**: When testing, use the browser's Developer Tools Console to see detailed WebAuthn logs and debug authentication issues.