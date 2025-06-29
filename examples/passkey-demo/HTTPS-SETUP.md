# HTTPS Setup for WebAuthn Passkey Demo

This guide explains how to set up HTTPS certificates for local development to enable full WebAuthn functionality with custom domains.

## 🔒 Why HTTPS is Required

**WebAuthn Security Requirement**: WebAuthn only works in secure contexts. While `localhost` gets a special exemption, custom domains like `passkey-demo.local` require HTTPS for WebAuthn to function properly.

### Without HTTPS (`http://passkey-demo.local:5173`)
- ❌ WebAuthn API unavailable
- ❌ "WebAuthn Not Supported" error
- ❌ No passkey functionality

### With HTTPS (`https://passkey-demo.local:5173`)
- ✅ Full WebAuthn API access
- ✅ Cross-platform passkey sharing
- ✅ Production-like testing environment

## 🚀 Quick Setup (Automated)

The easiest way to set up HTTPS is using the automated setup script:

```bash
# From the passkey-demo directory
./setup-https.sh
```

This script will:
1. Install `mkcert` via Homebrew (macOS)
2. Install the local CA certificate
3. Generate HTTPS certificates for `passkey-demo.local`
4. Configure `/etc/hosts` entries
5. Provide next steps

## 📋 Manual Setup (Step-by-Step)

If you prefer manual setup or need to understand each step:

### 1. Install mkcert

**macOS (Homebrew):**
```bash
brew install mkcert
```

**Linux:**
```bash
# Ubuntu/Debian
curl -JLO "https://dl.filippo.io/mkcert/latest?for=linux/amd64"
chmod +x mkcert-v*-linux-amd64
sudo cp mkcert-v*-linux-amd64 /usr/local/bin/mkcert

# Arch Linux
sudo pacman -S mkcert
```

**Windows:**
```powershell
# Using Chocolatey
choco install mkcert

# Using Scoop
scoop bucket add extras
scoop install mkcert
```

### 2. Install CA Certificate

```bash
mkcert -install
```

This installs the local Certificate Authority in your system's trust store. You'll see a message like:
```
Created a new local CA 💥
The local CA is now installed in the system trust store! ⚡️
```

### 3. Generate Certificates

```bash
# Navigate to the passkey-demo directory
cd examples/passkey-demo

# Create certificates directory
mkdir -p certs

# Generate certificates
cd certs
mkcert passkey-demo.local api.passkey-demo.local localhost 127.0.0.1 ::1
```

This creates two files:
- `passkey-demo.local+4.pem` (certificate)
- `passkey-demo.local+4-key.pem` (private key)

### 4. Configure Hosts File

Add these entries to your `/etc/hosts` file:

```bash
sudo vim /etc/hosts
```

Add these lines:
```
# WebAuthn Passkey Demo - Cross-Platform Configuration
127.0.0.1 passkey-demo.local
127.0.0.1 api.passkey-demo.local
```

### 5. Verify Setup

Test domain resolution:
```bash
ping passkey-demo.local
# Should respond from 127.0.0.1
```

## 🎯 Usage

### Starting the Backend

The backend automatically detects HTTPS certificates:

```bash
cd backend
go run .
```

**With HTTPS certificates:**
```
🚀 WebAuthn Passkey Demo Server
===============================
🌐 Cross-Platform Configuration:
🔐 WebAuthn RPID: passkey-demo.local
🔒 HTTPS Mode: ENABLED
📱 React Frontend: https://passkey-demo.local:5173
📡 Backend API: https://passkey-demo.local:8080
📜 Certificate: ../certs/passkey-demo.local+4.pem
```

**Without HTTPS certificates:**
```
🔓 HTTP Mode: Fallback (HTTPS certificates not found)
💡 Run './setup-https.sh' to enable HTTPS for full WebAuthn support
```

### Starting the Frontend

The frontend also automatically detects and uses HTTPS certificates:

```bash
cd frontend-react
npm run dev
```

**With HTTPS:**
```
VITE v5.0.0  ready in 500 ms

➜  Local:   https://localhost:5173/
➜  Network: https://passkey-demo.local:5173/
```

**Without HTTPS:**
```
VITE v5.0.0  ready in 500 ms

➜  Local:   http://localhost:5173/
➜  Network: http://passkey-demo.local:5173/
```

### Accessing the Demo

**Preferred (HTTPS):**
- Frontend: https://passkey-demo.local:5173
- Backend API: https://passkey-demo.local:8080

**Fallback (HTTP localhost):**
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080

## 🔧 Advanced Configuration

### Certificate Renewal

mkcert certificates are valid for 90 days. To renew:

```bash
cd examples/passkey-demo/certs
rm passkey-demo.local+4*
mkcert passkey-demo.local api.passkey-demo.local localhost 127.0.0.1 ::1
```

Restart both frontend and backend to pick up new certificates.

### Custom Certificate Paths

If you want to use different certificate paths, modify the file paths in:

**Backend (`backend/main.go`):**
```go
certFile := filepath.Join("path/to/your", "cert.pem")
keyFile := filepath.Join("path/to/your", "key.pem")
```

**Frontend (`frontend-react/vite.config.js`):**
```javascript
const certFile = path.join('/path/to/your', 'cert.pem')
const keyFile = path.join('/path/to/your', 'key.pem')
```

### Additional Domains

To add more domains to your certificate:

```bash
cd certs
mkcert passkey-demo.local api.passkey-demo.local my-custom-domain.local localhost 127.0.0.1 ::1
```

Update `/etc/hosts`:
```
127.0.0.1 passkey-demo.local
127.0.0.1 api.passkey-demo.local
127.0.0.1 my-custom-domain.local
```

## 🐛 Troubleshooting

### Certificate Not Trusted

**Symptoms:**
- Browser shows "Not Secure" warning
- NET::ERR_CERT_AUTHORITY_INVALID error

**Solutions:**
```bash
# Reinstall CA certificate
mkcert -uninstall
mkcert -install

# Regenerate certificates
cd certs
rm passkey-demo.local+4*
mkcert passkey-demo.local localhost 127.0.0.1
```

### Port Already in Use

**Symptoms:**
- "EADDRINUSE: address already in use :::8080"
- "Port 5173 is already in use"

**Solutions:**
```bash
# Find and kill processes using ports
lsof -ti:8080 | xargs kill
lsof -ti:5173 | xargs kill

# Or use different ports
PORT=8081 go run .  # Backend
npm run dev -- --port 5174  # Frontend
```

### Domain Not Resolving

**Symptoms:**
- Cannot connect to passkey-demo.local
- DNS_PROBE_FINISHED_NXDOMAIN

**Solutions:**
```bash
# Verify /etc/hosts entry
grep passkey-demo /etc/hosts

# Should show:
# 127.0.0.1 passkey-demo.local

# Test resolution
nslookup passkey-demo.local
# Should return 127.0.0.1
```

### WebAuthn Still Not Working

**Symptoms:**
- "WebAuthn Not Supported" despite HTTPS setup

**Checklist:**
1. ✅ Accessing via `https://passkey-demo.local:5173` (not http://)
2. ✅ Certificate trusted (no browser warnings)
3. ✅ Backend running in HTTPS mode
4. ✅ Modern browser (Chrome 67+, Firefox 60+, Safari 14+)

### Mixed Content Errors

**Symptoms:**
- "Mixed Content" warnings in browser console
- API calls failing from HTTPS frontend

**Solutions:**
- Ensure backend is also running HTTPS
- Check API_BASE in `frontend-react/src/services/api.js`
- Verify both frontend and backend use same protocol

## 🏠 Platform-Specific Notes

### macOS
- mkcert automatically integrates with Keychain Access
- Certificates appear in "System" keychain
- No additional configuration needed

### Linux
- May require additional packages: `libnss3-tools`
- Some distributions need manual certificate installation
- Check browser-specific certificate stores

### Windows
- Certificates install in Windows Certificate Store
- May require running as Administrator
- Internet Explorer/Edge use system store, others may need manual import

### Docker/WSL
- Certificates must be mounted into containers
- Pay attention to file paths and permissions
- Consider using host networking mode

## 🧹 Cleanup

To completely remove the HTTPS setup:

```bash
# Remove certificates
rm -rf examples/passkey-demo/certs

# Remove CA from system
mkcert -uninstall

# Remove hosts entries
sudo vim /etc/hosts
# Remove lines containing passkey-demo.local

# Restart applications
# Frontend and backend will fall back to HTTP mode
```

## 📚 Additional Resources

- [mkcert Documentation](https://github.com/FiloSottile/mkcert)
- [WebAuthn Secure Contexts](https://w3c.github.io/webauthn/#sctn-security-considerations)
- [MDN: Secure Contexts](https://developer.mozilla.org/en-US/docs/Web/Security/Secure_Contexts)
- [Chrome DevTools: Security Panel](https://developers.google.com/web/tools/chrome-devtools/security)