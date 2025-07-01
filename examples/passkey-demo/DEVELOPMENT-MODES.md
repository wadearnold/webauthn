# Passkey Demo Development Modes - Quick Reference

## 🚀 Quick Decision Guide

**Question: What are you working on?**

### "I'm working on the React UI or backend logic"
→ Use **Local Development Mode**

### "I'm testing iOS/Android apps or cross-platform features"
→ Use **Cross-Platform Mode with ngrok**

---

## 📋 Mode Comparison

| Feature | Local Mode | Cross-Platform Mode |
|---------|------------|-------------------|
| **Setup Time** | Instant | ~2 minutes |
| **React Hot Reload** | ✅ Yes | ❌ No (requires build) |
| **Cross-Platform Passkeys** | ❌ No | ✅ Yes |
| **iOS/Android Testing** | ❌ No | ✅ Yes |
| **Real Device Testing** | ❌ No | ✅ Yes |
| **RPID Domain** | localhost | your-tunnel.ngrok.io |
| **Access URL** | http://localhost:5173 | https://your-tunnel.ngrok.io |

---

## 🛠️ Commands

### Local Development Mode
```bash
# Terminal 1: Backend
cd backend
go run .

# Terminal 2: React
cd frontend-react
npm run dev

# Open: http://localhost:5173
```

### Cross-Platform Mode
```bash
# One-time setup (keeps running)
./scripts/start-ngrok.sh

# Terminal 1: Build & Serve
cd frontend-react
npm run build
cd ../backend
source ../.env && go run .

# Open: https://your-tunnel.ngrok.io
```

---

## 🔄 Switching Between Modes

### From Local → Cross-Platform
1. Stop backend (Ctrl+C)
2. Start ngrok: `./scripts/start-ngrok.sh` (if not running)
3. Build React: `cd frontend-react && npm run build`
4. Restart backend: `cd ../backend && source ../.env && go run .`

### From Cross-Platform → Local
1. Stop backend (Ctrl+C)
2. Restart without env: `cd backend && go run .`
3. Start React dev: `cd frontend-react && npm run dev`

---

## ⚠️ Common Pitfalls

### "SecurityError: RPID mismatch"
- **Cause**: Accessing cross-platform mode via localhost:5173
- **Fix**: Use the ngrok URL shown in backend startup

### "Failed to fetch" errors
- **Cause**: Backend not running or wrong mode
- **Fix**: Check backend is running and using correct mode

### "Passkeys not syncing to iOS"
- **Cause**: Using local development mode
- **Fix**: Switch to cross-platform mode with ngrok

### "Changes not reflecting immediately"
- **Cause**: Using cross-platform mode (no hot reload)
- **Fix**: Rebuild React (`npm run build`) or switch to local mode

---

## 💡 Pro Tips

1. **Start with local mode** for UI development
2. **Switch to cross-platform** only when testing mobile apps
3. **Keep ngrok running** all day - no need to restart
4. **Use hybrid workflow**: develop locally, test cross-platform

---

## 🔍 How to Check Current Mode

Look at backend startup output:

**Local Mode:**
```
🔐 RPID: localhost
📍 Mode: Local Development
```

**Cross-Platform Mode:**
```
🔐 RPID: abc123.ngrok.io
🌍 Mode: ngrok (Cross-Platform)
```