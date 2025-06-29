# iOS Simulator DNS Resolution for passkey-demo.local

## Problem
The error "A server with the specified hostname could not be found" occurs because the iOS simulator cannot resolve the custom domain `passkey-demo.local`.

## ✅ **IMPORTANT**: Why passkey-demo.local is Required

The iOS app **MUST** use `passkey-demo.local` (same as web frontend) to demonstrate true cross-platform passkey compatibility. Using different domains would create separate passkey scopes and defeat the demo's purpose.

## Solutions

### Solution 1: Reset iOS Simulator (Recommended)

iOS Simulator should inherit your Mac's `/etc/hosts` file automatically, but sometimes gets stuck:

```bash
# 1. Close Xcode and Simulator completely
killall "Simulator"
killall "Xcode"

# 2. Reset simulator DNS cache
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder

# 3. Restart Simulator
open -a Simulator

# 4. Test DNS resolution in simulator
# Open Safari in simulator and try: https://passkey-demo.local:8080/api/health
```

### Solution 2: Verify /etc/hosts Configuration

Ensure your Mac's hosts file is correctly configured:

```bash
# Check current hosts file
cat /etc/hosts | grep passkey-demo

# Should show:
# 127.0.0.1 passkey-demo.local
# 127.0.0.1 api.passkey-demo.local

# If missing, add it:
sudo bash -c 'echo "127.0.0.1 passkey-demo.local" >> /etc/hosts'
```

### Solution 3: iOS Simulator Network Reset

If simulator still can't resolve the domain:

```bash
# 1. Reset all simulators (nuclear option)
xcrun simctl shutdown all
xcrun simctl erase all

# 2. Restart Mac networking
sudo ifconfig en0 down
sudo ifconfig en0 up

# 3. Restart Simulator
open -a Simulator
```

### Solution 4: Test with HTTP (Temporary Debug)

If HTTPS isn't working, temporarily test with HTTP to isolate the issue:

1. **Update backend to allow HTTP CORS:**
   ```go
   // In middleware.go, ensure HTTP origins are included:
   "http://passkey-demo.local:8080",
   ```

2. **Temporarily use HTTP in iOS app:**
   ```swift
   // In APIService.swift, temporarily change:
   private let baseURL = "http://passkey-demo.local:8080/api"
   ```

3. **Allow HTTP in iOS (for testing only):**
   
   Add to `PasskeyDemo/Info.plist`:
   ```xml
   <key>NSAppTransportSecurity</key>
   <dict>
       <key>NSExceptionDomains</key>
       <dict>
           <key>passkey-demo.local</key>
           <dict>
               <key>NSExceptionAllowsInsecureHTTPLoads</key>
               <true/>
           </dict>
       </dict>
   </dict>
   ```

### Solution 5: Physical Device Testing

For physical iOS devices, you need network-level configuration:

1. **Connect device to same network as Mac**
2. **Use router-level DNS or mDNS setup**
3. **Or use a development proxy like ngrok**

## Testing Checklist

- [ ] `/etc/hosts` contains `127.0.0.1 passkey-demo.local`
- [ ] Backend running with HTTPS: `https://passkey-demo.local:8080`
- [ ] iOS Simulator DNS cache flushed
- [ ] Simulator can access: `https://passkey-demo.local:8080/api/health`
- [ ] Web frontend works: `https://passkey-demo.local:5173`
- [ ] Both use the same RPID for cross-platform passkey sharing

## Expected Cross-Platform Flow

1. **Register passkey on web** at `https://passkey-demo.local:5173`
2. **Same passkey appears in iOS app** (if synced via iCloud)
3. **Can authenticate on iOS** using web-created passkey
4. **Can authenticate on web** using iOS-created passkey

## Debug Commands

```bash
# Test DNS resolution on Mac
nslookup passkey-demo.local
# Should return: 127.0.0.1

# Test backend connectivity
curl -k https://passkey-demo.local:8080/api/health

# Check iOS Simulator logs
xcrun simctl spawn booted log stream --predicate 'subsystem contains "com.apple.network"'
```

## If All Else Fails

As a last resort for development, you can temporarily use ngrok to test the cross-platform concept:

```bash
# Install ngrok
brew install ngrok

# Expose backend with HTTPS
ngrok http 8080

# Update both web and iOS to use ngrok URL
# Example: https://abc123.ngrok.io/api
```

But remember: **the goal is to demonstrate that both platforms use the same domain for true passkey compatibility!**