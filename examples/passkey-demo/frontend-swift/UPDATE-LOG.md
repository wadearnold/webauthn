# iOS App Updates - Dynamic Domain Support

## What Was Changed

Updated the iOS app to dynamically display the correct web URL for cross-platform testing instead of hardcoded `passkey-demo.local` references.

### Changes Made:

1. **DashboardView.swift** - Line 276:
   - **Before**: `"Use the same passkeys on the web at https://passkey-demo.local:5173"`
   - **After**: `"Use the same passkeys on the web at \(getCurrentWebURL())"`

2. **Added Dynamic URL Function**:
   ```swift
   private func getCurrentWebURL() -> String {
       // Get the current ngrok URL from APIConfiguration
       if let ngrokURL = APIConfiguration.ngrokURL {
           return ngrokURL
       } else {
           return "http://localhost:8080"
       }
   }
   ```

## User Experience Improvement

Now when users complete registration or login, the dashboard will show:

**In ngrok mode:**
> "Use the same passkeys on the web at https://abc123.ngrok-free.app"

**In localhost mode:**
> "Use the same passkeys on the web at http://localhost:8080"

This makes it clear where users should go to test cross-platform passkey functionality and eliminates confusion about outdated domain references.

## Impact

- ✅ Users see the correct URL for their current configuration
- ✅ No more confusion about `passkey-demo.local` domain
- ✅ Seamless experience for cross-platform testing
- ✅ Works automatically with setup scripts

## Testing

After building the updated app:
1. Register a passkey on iOS
2. Check the dashboard shows your current ngrok URL
3. Visit that URL in Safari to test cross-platform passkey login