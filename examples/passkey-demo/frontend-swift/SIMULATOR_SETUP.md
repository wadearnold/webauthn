# iOS Simulator/Device Network Setup for passkey-demo.local

## Problem
The error "A server with the specified hostname could not be found" occurs because iOS devices/simulators cannot resolve the custom domain `passkey-demo.local` without additional configuration.

## Solutions

### Option 1: Use IP Address (Quickest for Testing)

1. **Find your Mac's IP address:**
   ```bash
   ifconfig | grep "inet " | grep -v 127.0.0.1
   # Look for something like: inet 192.168.1.100
   ```

2. **Update the iOS app's API base URL:**
   
   Edit `PasskeyDemo/Services/APIService.swift`:
   ```swift
   // Replace this line:
   private let baseURL = "https://passkey-demo.local:8080/api"
   
   // With your Mac's IP:
   private let baseURL = "https://192.168.1.100:8080/api"
   ```

3. **Update backend CORS:**
   
   Add your IP to `backend/middleware.go` allowed origins:
   ```go
   "https://192.168.1.100:8080", // Your Mac's IP
   ```

### Option 2: Configure DNS for iOS Simulator

1. **For iOS Simulator on the same Mac as backend:**
   
   The simulator should inherit your Mac's `/etc/hosts` file. If not working:
   
   ```bash
   # Reset simulator
   xcrun simctl shutdown all
   xcrun simctl erase all
   
   # Restart simulator
   ```

2. **Use localhost fallback:**
   
   Edit `PasskeyDemo/Services/APIService.swift`:
   ```swift
   // For simulator testing, use localhost
   private let baseURL = "https://localhost:8080/api"
   ```

### Option 3: Physical Device Setup

1. **Using Charles Proxy or similar:**
   - Install Charles Proxy on your Mac
   - Configure iOS device to use Mac as HTTP proxy
   - Charles will resolve local domains

2. **Using ngrok (for HTTPS):**
   ```bash
   # Install ngrok
   brew install ngrok
   
   # Expose your backend
   ngrok http 8080
   
   # Use the ngrok URL in your app
   # Example: https://abc123.ngrok.io/api
   ```

### Option 4: Temporary HTTP Testing (Not Recommended)

1. **Allow HTTP in Info.plist:**
   
   Add to `PasskeyDemo/Info.plist`:
   ```xml
   <key>NSAppTransportSecurity</key>
   <dict>
       <key>NSAllowsArbitraryLoads</key>
       <true/>
   </dict>
   ```

2. **Use HTTP URLs:**
   ```swift
   private let baseURL = "http://localhost:8080/api"
   ```

## Recommended Development Setup

### For Quick Testing:
1. Use your Mac's IP address (Option 1)
2. Ensure backend is running with HTTPS certificates
3. Make sure both devices are on the same network

### For Production-like Testing:
1. Use ngrok to expose your backend with a real HTTPS URL
2. Update backend CORS to allow ngrok domain
3. Test with real cross-platform scenarios

## Backend Updates Needed

1. **Update main.go to listen on all interfaces:**
   ```go
   // Change from:
   server := &http.Server{
       Addr:    ":8080",
       Handler: handler,
   }
   
   // To:
   server := &http.Server{
       Addr:    "0.0.0.0:8080", // Listen on all interfaces
       Handler: handler,
   }
   ```

2. **Add IP-based origins to CORS:**
   ```go
   // In middleware.go, add:
   "https://192.168.1.100:8080", // Your Mac's IP
   "http://192.168.1.100:8080",  // HTTP fallback
   ```

## Testing Checklist

- [ ] Backend running with HTTPS certificates
- [ ] Backend listening on all interfaces (0.0.0.0)
- [ ] iOS app using correct URL (IP or localhost)
- [ ] CORS configured for the URL being used
- [ ] Both devices on same network (if using IP)
- [ ] Certificates trusted (if using HTTPS with IP)

## Quick Fix for Your Current Error

The fastest solution is to update your `APIService.swift`:

```swift
// Change this:
private let baseURL = "https://passkey-demo.local:8080/api"

// To this (for simulator):
private let baseURL = "https://localhost:8080/api"

// Or to your Mac's IP (for device):
private let baseURL = "https://192.168.1.100:8080/api" // Replace with your IP
```

Then ensure your backend CORS includes the matching origin.