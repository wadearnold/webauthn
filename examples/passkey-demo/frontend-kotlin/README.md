# WebAuthn Passkey Demo - Android Kotlin Frontend

A native Android Kotlin app demonstrating WebAuthn passkey authentication that works seamlessly with Google Password Manager and cross-platform keychain sync.

## 🎯 Demo Goals

Showcase that passkeys created on any device (web, iOS, Android) can be used to authenticate across platforms when using shared keychains (Google Password Manager, iCloud Keychain, etc.).

## 🛠 Implementation Plan

### Core Features to Implement

- **Passkey Registration**: Create new passkeys using Android WebAuthn APIs
- **Passwordless Authentication**: Sign in using fingerprint/face unlock
- **Cross-Platform Sync**: Demonstrate passkeys work across devices via Google PM
- **Passkey Management**: View and delete user's passkeys
- **Deep Link Authentication**: Handle authentication from external links

### Technical Stack

- **Language**: Kotlin
- **Framework**: Jetpack Compose
- **WebAuthn**: androidx.credentials (Credential Manager API)
- **Biometrics**: androidx.biometric
- **Networking**: Retrofit + OkHttp for backend communication
- **Architecture**: MVVM with ViewModels and Repository pattern
- **Minimum SDK**: API 28 (Android 9) for WebAuthn support

### Project Structure

```
PasskeyDemoAndroid/
├── app/
│   ├── src/main/
│   │   ├── java/com/example/passkeydemo/
│   │   │   ├── ui/
│   │   │   │   ├── register/
│   │   │   │   │   ├── RegisterScreen.kt
│   │   │   │   │   └── RegisterViewModel.kt
│   │   │   │   ├── login/
│   │   │   │   │   ├── LoginScreen.kt
│   │   │   │   │   └── LoginViewModel.kt
│   │   │   │   ├── dashboard/
│   │   │   │   │   ├── DashboardScreen.kt
│   │   │   │   │   └── DashboardViewModel.kt
│   │   │   │   └── profile/
│   │   │   │       ├── ProfileScreen.kt
│   │   │   │       └── ProfileViewModel.kt
│   │   │   ├── data/
│   │   │   │   ├── api/
│   │   │   │   │   ├── ApiService.kt
│   │   │   │   │   └── ApiModels.kt
│   │   │   │   ├── repository/
│   │   │   │   │   └── PasskeyRepository.kt
│   │   │   │   └── webauthn/
│   │   │   │       └── WebAuthnManager.kt
│   │   │   ├── utils/
│   │   │   │   ├── Base64Utils.kt
│   │   │   │   └── BiometricsHelper.kt
│   │   │   └── MainActivity.kt
│   │   ├── res/
│   │   └── AndroidManifest.xml
│   ├── build.gradle.kts
│   └── proguard-rules.pro
├── build.gradle.kts
└── README.md
```

### Key Implementation Components

#### 1. WebAuthn Manager (`WebAuthnManager.kt`)

```kotlin
import androidx.credentials.CredentialManager
import androidx.credentials.CreatePublicKeyCredentialRequest
import androidx.credentials.GetCredentialRequest
import androidx.credentials.GetPublicKeyCredentialOption
import androidx.credentials.PublicKeyCredential

class WebAuthnManager(private val context: Context) {
    
    private val credentialManager = CredentialManager.create(context)
    
    suspend fun createPasskey(
        challenge: String,
        userId: String,
        userName: String,
        displayName: String
    ): PublicKeyCredential {
        
        val requestJson = JSONObject().apply {
            put("challenge", challenge)
            put("rp", JSONObject().apply {
                put("name", "WebAuthn Passkey Demo")
                put("id", "localhost")
            })
            put("user", JSONObject().apply {
                put("id", userId)
                put("name", userName)
                put("displayName", displayName)
            })
            put("pubKeyCredParams", JSONArray().apply {
                put(JSONObject().apply {
                    put("alg", -7)
                    put("type", "public-key")
                })
            })
            put("authenticatorSelection", JSONObject().apply {
                put("authenticatorAttachment", "platform")
                put("userVerification", "required")
                put("requireResidentKey", true)
            })
            put("timeout", 60000)
        }.toString()
        
        val request = CreatePublicKeyCredentialRequest(requestJson)
        
        val result = credentialManager.createCredential(
            context = context as ComponentActivity,
            request = request
        )
        
        return result.credential as PublicKeyCredential
    }
    
    suspend fun authenticateWithPasskey(challenge: String, allowCredentials: List<String>? = null): PublicKeyCredential {
        val requestJson = JSONObject().apply {
            put("challenge", challenge)
            put("rpId", "localhost")
            put("userVerification", "required")
            put("timeout", 60000)
            allowCredentials?.let { creds ->
                put("allowCredentials", JSONArray().apply {
                    creds.forEach { credId ->
                        put(JSONObject().apply {
                            put("id", credId)
                            put("type", "public-key")
                        })
                    }
                })
            }
        }.toString()
        
        val getPublicKeyCredentialOption = GetPublicKeyCredentialOption(requestJson)
        val getCredRequest = GetCredentialRequest(listOf(getPublicKeyCredentialOption))
        
        val result = credentialManager.getCredential(
            context = context as ComponentActivity,
            request = getCredRequest
        )
        
        return result.credential as PublicKeyCredential
    }
}
```

#### 2. API Service (`ApiService.kt`)

```kotlin
import retrofit2.http.*

interface ApiService {
    
    @POST("api/register/begin")
    suspend fun registerBegin(@Body request: RegisterBeginRequest): RegistrationOptions
    
    @POST("api/register/finish")
    suspend fun registerFinish(@Body credential: CredentialResponse): AuthResponse
    
    @POST("api/login/begin")
    suspend fun loginBegin(@Body request: LoginBeginRequest): AuthenticationOptions
    
    @POST("api/login/finish")
    suspend fun loginFinish(@Body assertion: AssertionResponse): AuthResponse
    
    @GET("api/user/passkeys")
    suspend fun getUserPasskeys(): List<PasskeyInfo>
    
    @DELETE("api/user/passkeys/{id}")
    suspend fun deletePasskey(@Path("id") credentialId: String)
    
    @POST("api/logout")
    suspend fun logout()
}

data class RegisterBeginRequest(
    val username: String,
    val displayName: String
)

data class LoginBeginRequest(
    val username: String? = null
)
```

#### 3. Dashboard Screen (`DashboardScreen.kt`)

```kotlin
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DashboardScreen(
    viewModel: DashboardViewModel,
    onLogout: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    
    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        
        // User info header
        Card(modifier = Modifier.fillMaxWidth()) {
            Row(
                modifier = Modifier.padding(16.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column {
                    Text(
                        text = "Welcome, ${uiState.user?.displayName ?: uiState.user?.username}!",
                        style = MaterialTheme.typography.headlineSmall
                    )
                    Text(
                        text = "Signed in with passkey authentication",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Button(onClick = onLogout) {
                    Text("Sign Out")
                }
            }
        }
        
        Spacer(modifier = Modifier.height(16.dp))
        
        // Passkeys section
        Text(
            text = "Your Passkeys",
            style = MaterialTheme.typography.headlineMedium
        )
        
        if (uiState.isLoading) {
            Box(modifier = Modifier.fillMaxWidth().padding(32.dp)) {
                CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
            }
        } else if (uiState.passkeys.isEmpty()) {
            NoPasskeysCard()
        } else {
            LazyColumn(modifier = Modifier.fillMaxWidth()) {
                items(uiState.passkeys) { passkey ->
                    PasskeyItem(
                        passkey = passkey,
                        onDelete = { viewModel.deletePasskey(passkey.id) }
                    )
                }
            }
        }
        
        // Cross-platform demo information
        Spacer(modifier = Modifier.height(16.dp))
        CrossPlatformInfoCard()
    }
}

@Composable
fun PasskeyItem(
    passkey: PasskeyInfo,
    onDelete: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = "🔑 ${passkey.name}",
                        style = MaterialTheme.typography.titleMedium
                    )
                    Text(
                        text = "Created: ${passkey.createdAt}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                        text = "Transport: ${passkey.transports.joinToString()}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                IconButton(onClick = onDelete) {
                    Icon(Icons.Default.Delete, contentDescription = "Delete passkey")
                }
            }
            
            // Passkey status indicators
            Row(modifier = Modifier.padding(top = 8.dp)) {
                if (passkey.backedUp) {
                    AssistChip(
                        onClick = { },
                        label = { Text("☁️ Synced") }
                    )
                }
                if (passkey.userVerified) {
                    AssistChip(
                        onClick = { },
                        label = { Text("🔐 Biometric") }
                    )
                }
            }
        }
    }
}
```

#### 4. Repository Pattern (`PasskeyRepository.kt`)

```kotlin
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class PasskeyRepository(
    private val apiService: ApiService,
    private val webAuthnManager: WebAuthnManager
) {
    
    suspend fun registerPasskey(username: String, displayName: String): Result<AuthResponse> {
        return try {
            // Start registration
            val options = apiService.registerBegin(RegisterBeginRequest(username, displayName))
            
            // Create passkey
            val credential = webAuthnManager.createPasskey(
                challenge = options.challenge,
                userId = options.user.id,
                userName = options.user.name,
                displayName = options.user.displayName
            )
            
            // Finish registration
            val response = apiService.registerFinish(credential.toCredentialResponse())
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun authenticateWithPasskey(username: String? = null): Result<AuthResponse> {
        return try {
            // Start authentication
            val options = apiService.loginBegin(LoginBeginRequest(username))
            
            // Authenticate with passkey
            val assertion = webAuthnManager.authenticateWithPasskey(
                challenge = options.challenge,
                allowCredentials = options.allowCredentials?.map { it.id }
            )
            
            // Finish authentication
            val response = apiService.loginFinish(assertion.toAssertionResponse())
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    fun getUserPasskeys(): Flow<List<PasskeyInfo>> = flow {
        emit(apiService.getUserPasskeys())
    }
    
    suspend fun deletePasskey(credentialId: String): Result<Unit> {
        return try {
            apiService.deletePasskey(credentialId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
```

### Development Steps

1. **Setup Project**
   ```bash
   # Create new Android project in Android Studio
   # Set minSdk to 28 (Android 9)
   # Add required dependencies
   ```

2. **Dependencies (`build.gradle.kts`)**
   ```kotlin
   dependencies {
       implementation "androidx.credentials:credentials:1.2.2"
       implementation "androidx.credentials:credentials-play-services-auth:1.2.2"
       implementation "androidx.biometric:biometric:1.1.0"
       implementation "androidx.compose.ui:ui:$compose_version"
       implementation "androidx.compose.material3:material3:$material3_version"
       implementation "androidx.lifecycle:lifecycle-viewmodel-compose:$lifecycle_version"
       implementation "com.squareup.retrofit2:retrofit:2.9.0"
       implementation "com.squareup.retrofit2:converter-gson:2.9.0"
       implementation "androidx.navigation:navigation-compose:$nav_version"
   }
   ```

3. **Permissions (`AndroidManifest.xml`)**
   ```xml
   <uses-permission android:name="android.permission.INTERNET" />
   <uses-permission android:name="android.permission.USE_FINGERPRINT" />
   <uses-permission android:name="android.permission.USE_BIOMETRIC" />
   ```

4. **Implementation Phases**
   - Phase 1: Basic WebAuthn integration
   - Phase 2: UI with Jetpack Compose
   - Phase 3: Google Password Manager integration
   - Phase 4: Cross-platform testing

### Cross-Platform Sync Testing

The Android app should demonstrate:

1. **Google Password Manager**: Passkeys sync across Android devices and Chrome
2. **Cross-Platform Compatibility**: Passkeys work with iOS (via Google PM) and web
3. **Enterprise Integration**: Managed passkeys for work profiles
4. **Deep Links**: Authentication from external apps/web

### Demo Flow Examples

1. **Create on Android** → Authenticate on web browser
2. **Create on iPhone** → Authenticate on Android (via Google sync)
3. **Create on web** → Authenticate on Android app
4. **Delete from Android** → Verify removal across platforms

## 📱 Getting Started

```bash
# Prerequisites
# - Android Studio Hedgehog or newer
# - Android device with API 28+ (Android 9+)
# - Google Play Services 20.2+
# - Backend server running on localhost:8080

# Steps
1. Open Android Studio
2. Create new project with Jetpack Compose
3. Set minSdk to 28
4. Add Credential Manager dependencies
5. Implement WebAuthn integration
6. Test on physical device (required for biometrics)
```

## 🔄 Integration with Existing Demo

This Android frontend will use the **same backend API** as the React and iOS frontends, demonstrating true cross-platform passkey compatibility.

### Shared Backend Endpoints
- `POST /api/register/begin` - Start passkey registration
- `POST /api/register/finish` - Complete passkey registration  
- `POST /api/login/begin` - Start authentication
- `POST /api/login/finish` - Complete authentication
- `GET /api/user/passkeys` - Get user's passkeys
- `DELETE /api/user/passkeys/{id}` - Delete specific passkey

### Cross-Platform Testing Matrix

| Create Platform | Authenticate Platform | Sync Method | Expected Result |
|----------------|----------------------|-------------|-----------------|
| Android App | React Web | Google PM | ✅ Should work |
| React Web | Android App | Google PM | ✅ Should work |
| iOS App | Android App | Google PM | ✅ Should work |
| Android App | iOS App | Google PM | ✅ Should work |
| Android Device A | Android Device B | Google Sync | ✅ Should work |

## 🚀 Future Enhancements

- **App Links**: Deep link authentication from web/other apps
- **Instant Apps**: Lightweight authentication experiences
- **Google Assistant**: Voice-triggered authentication
- **Wear OS**: Companion app for wrist-based authentication
- **Enterprise Features**: Work profile and managed device support
- **Autofill Integration**: Seamless form filling with passkeys

## 🔧 Technical Considerations

### Google Password Manager Integration
- Requires Google Play Services 20.2+
- Passkeys automatically sync to user's Google account
- Works across Android devices and Chrome browsers
- Enterprise policies can control sync behavior

### Biometric Requirements
- Device must have secure lock screen
- Biometric enrollment required for passkey creation
- Fallback to PIN/pattern if biometrics fail
- Hardware security module preferred

### WebAuthn Attestation
- Android SafetyNet attestation available
- Play Integrity API for device verification
- Enterprise attestation for managed devices

---

**Status**: 📋 **Implementation Planned** - Ready for development