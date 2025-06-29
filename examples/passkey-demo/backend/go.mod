module github.com/go-webauthn/webauthn/examples/passkey-demo/backend

go 1.23.0

toolchain go1.24.4

require (
	github.com/go-webauthn/webauthn v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
)

require (
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/go-webauthn/x v0.1.21 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/go-tpm v0.9.5 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/go-webauthn/webauthn => ../../..
