# WebAuthn RPID and Origin Security Guide

## The Critical Security Rule

**⚠️ WebAuthn's #1 Security Rule: The browser origin MUST match or be a subdomain of the RPID (Relying Party ID).**

This guide explains why this matters and how to work with it during development.

## Understanding the Problem

### What is RPID?
The RPID (Relying Party ID) is the domain that "owns" the passkey. When you create a passkey, it's permanently tied to this domain.

### The Security Check
When creating a passkey, the browser performs this check:
```
Browser Origin: http://localhost:5173
Requested RPID: abc123.ngrok-free.app

Browser says: "DENIED! localhost cannot create credentials for ngrok domain"
```

### The Error You'll See
```
SecurityError: The relying party ID is not a registrable domain suffix of, 
nor equal to the current domain.
```

## Why This Security Exists

This prevents malicious websites from creating passkeys for other domains:
- `evil-site.com` cannot create passkeys for `yourbank.com`
- `phishing-site.net` cannot impersonate `google.com`

## Valid RPID Scenarios

### ✅ Valid Examples:
- Origin: `https://example.com` → RPID: `example.com` ✅
- Origin: `https://app.example.com` → RPID: `example.com` ✅ (parent domain)
- Origin: `http://localhost:5173` → RPID: `localhost` ✅

### ❌ Invalid Examples:
- Origin: `http://localhost:5173` → RPID: `ngrok.io` ❌
- Origin: `https://site-a.com` → RPID: `site-b.com` ❌
- Origin: `http://example.com` → RPID: `sub.example.com` ❌ (can't use subdomain)

## Development Strategies

### Strategy 1: Local Development Mode
**When to use:** Day-to-day development, UI work, testing basic flows

```bash
# Backend uses RPID: localhost
cd backend
go run .

# Frontend development server
cd frontend-react
npm run dev

# Access at: http://localhost:5173
```

**Pros:**
- Simple setup
- Hot reload works
- Fast development cycle

**Cons:**
- Cannot test cross-platform passkeys
- Different RPID than production

### Strategy 2: Production-like with ngrok
**When to use:** Cross-platform testing, iOS integration, production simulation

```bash
# 1. Start ngrok
./scripts/start-ngrok.sh

# 2. Build React app
cd frontend-react
npm run build

# 3. Backend serves everything
cd backend
source ../.env && go run .

# Access at: https://abc123.ngrok-free.app
```

**Pros:**
- Same RPID everywhere
- Works with iOS/Android
- Production-like setup

**Cons:**
- No hot reload
- Must rebuild React for changes
- Requires ngrok setup

## Common Pitfalls and Solutions

### Pitfall 1: Mixing Origins and RPIDs
**Problem:** Accessing frontend at `localhost:5173` but backend uses ngrok RPID
**Solution:** Choose one strategy and stick with it

### Pitfall 2: Forgetting to Rebuild
**Problem:** Changes not showing in Strategy 2
**Solution:** Remember to `npm run build` after React changes

### Pitfall 3: Cross-Platform Testing Fails
**Problem:** iOS app can't use passkeys created on localhost
**Solution:** Use Strategy 2 with consistent ngrok domain

## Quick Decision Tree

```
Need hot reload for UI development?
  → Use Strategy 1 (localhost)

Testing cross-platform passkeys?
  → Use Strategy 2 (ngrok)

Building production app?
  → Your app domain becomes the RPID

Just learning WebAuthn?
  → Start with Strategy 1
```

## Best Practices

1. **Development:** Use localhost RPID for rapid development
2. **Testing:** Switch to ngrok for cross-platform testing
3. **Production:** Use your actual domain as RPID
4. **Documentation:** Always document which RPID your setup uses

## Debugging Tips

### Check Current RPID
```javascript
// In browser console while on your app
console.log(window.location.hostname); // Your origin
// Check network tab for /register/begin response to see RPID
```

### Verify Setup
1. Browser origin must match or be subdomain of RPID
2. Check backend logs for current RPID
3. Use browser DevTools to inspect WebAuthn API calls

## The Golden Rule

**If the browser URL domain doesn't match the RPID domain, WebAuthn will fail.**

This is not a bug - it's a critical security feature that protects users from credential phishing.