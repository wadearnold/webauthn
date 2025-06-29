# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Context

This is a fork of `github.com/go-webauthn/webauthn`. The fork is located at `github.com/wadearnold/webauthn`.

### Local Development vs Upstream Commits

**CRITICAL**: This repository requires different import path strategies for local development versus upstream commits:

#### Local Development (Fork)
When working on this fork (`github.com/wadearnold/webauthn`), the go.mod should reference the fork:

```go
// go.mod for local development
module github.com/wadearnold/webauthn

// Imports will use the fork path
import (
    "github.com/wadearnold/webauthn/protocol"
    "github.com/wadearnold/webauthn/webauthn"
)
```

#### Upstream Commits (go-webauthn)
When preparing commits for the upstream repository (`github.com/go-webauthn/webauthn`), ensure the module path uses the upstream:

```go
// go.mod for upstream commits
module github.com/go-webauthn/webauthn

// Imports use the upstream path
import (
    "github.com/go-webauthn/webauthn/protocol"
    "github.com/go-webauthn/webauthn/webauthn"
)
```

#### Development Workflow
1. **During development**: Work with fork imports for local testing
2. **Before upstream PR**: Switch module path and imports to `github.com/go-webauthn/webauthn`
3. **Always verify**: Run `go test ./...` and `golangci-lint run` before submitting upstream

#### Why This Matters
- **Local development**: Test changes in isolation on your fork
- **Upstream compatibility**: PRs must use correct import paths for CI/CD
- **Module resolution**: Go modules require consistent import paths

## Development Commands

### Building and Testing
```bash
# Run all tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./webauthn
go test -v ./protocol
go test -v ./metadata

# Run tests with race detection
go test -race ./...

# Build the library
go build ./...
```

### Code Quality
```bash
# Run linting (requires golangci-lint installed)
golangci-lint run

# Format code
go fmt ./...
gofmt -w .

# Check imports
goimports -w -local github.com/go-webauthn/webauthn .
```

## Architecture Overview

This is a Go library implementing the W3C WebAuthn specification for passwordless authentication. The codebase is organized into three main layers:

1. **Protocol Layer** (`/protocol/`): Low-level WebAuthn protocol implementation
   - `webauthncbor/`: CBOR encoding/decoding for WebAuthn data structures
   - `webauthncose/`: COSE (CBOR Object Signing and Encryption) implementation
   - Core types: `CredentialCreation`, `CredentialAssertion`, `ParsedCredential`
   - Attestation format handlers (TPM, U2F, Apple, Android, Packed)

2. **WebAuthn Layer** (`/webauthn.go`, `/login.go`, `/registration.go`): High-level API
   - Main entry point: `webauthn.WebAuthn` struct
   - Key flows: `BeginRegistration`, `FinishRegistration`, `BeginLogin`, `FinishLogin`
   - Session management for multi-step authentication flows
   - Configuration through `webauthn.Config`

3. **Metadata Layer** (`/metadata/`): Authenticator metadata validation
   - Provider interface for fetching authenticator metadata
   - Cached and memory provider implementations
   - MDS (Metadata Service) client support

## Key Development Patterns

- **Error Handling**: Use wrapped errors with `fmt.Errorf` and the custom `Error` type in `/protocol/errors.go`
- **Testing**: Every feature should have corresponding tests using `testify/assert`
- **Validation**: Input validation happens at protocol layer, business logic at webauthn layer
- **CBOR/COSE**: Use the `webauthncbor` and `webauthncose` packages for all encoding/decoding
- **Commit Messages**: Follow Angular convention: `<type>(<scope>): <summary>` where type is one of: build, ci, docs, feat, fix, perf, refactor, revert, release, test

## Important Implementation Details

- The library maintains backward compatibility with `github.com/duo-labs/webauthn` while adding WebAuthn Level 3 features
- Credential storage is handled by the integrating application through the `User` and `Credential` interfaces
- Session data must be stored between Begin/Finish operations (typically in server session storage)
- All cryptographic operations follow W3C WebAuthn and FIDO2 specifications exactly
- TPM attestation requires additional parsing through `github.com/google/go-tpm`

## Common Tasks

When implementing new attestation formats:
1. Add the format handler in `/protocol/attestation_<format>.go`
2. Register it in `/protocol/attestation.go`
3. Add comprehensive tests including test vectors from the spec
4. Update the attestation documentation

When adding new WebAuthn features:
1. Check the W3C WebAuthn spec for the exact requirements
2. Implement at the protocol layer first
3. Expose through the high-level API in `/webauthn.go`
4. Add both unit and integration tests
5. Update examples if the feature affects the API

## Claude Development Files Organization

### `.claude/` Directory Structure

All Claude-generated files are organized in the `.claude/` directory to keep the project root clean:

```
.claude/
├── docs/           # Documentation and architectural guides
├── scripts/        # Automation scripts and build tools
├── coverage/       # Test coverage reports (gitignored)
└── archive/        # Historical/deprecated files
```

#### File Placement Guidelines

**`.claude/docs/`** - Documentation files:
- Architecture guides (BASE_ABSTRACTIONS.md, IMPLEMENTATION_GUIDE.md)
- Design decisions (ERROR_DESIGN_PROPOSAL.md, TYPE_SAFETY_ANALYSIS.md)
- Migration documentation (MIGRATION_STATUS.md, WRAPPER_MIGRATION_PLAN.md)
- Field mapping guides (XML_TO_GO_MAPPING.md)
- Test strategies and coverage analysis

**`.claude/scripts/`** - Automation scripts:
- Python scripts for code generation (`apply_*.py`, `fix_*.py`)
- Shell scripts for batch operations (`*.sh`)
- Refactoring and migration tools
- Build and deployment automation

**`.claude/coverage/`** - Coverage reports (gitignored):
- `cover.out`, `coverage.out`, `coverage.txt`
- Test coverage analysis files
- Performance benchmarks

**`.claude/archive/`** - Deprecated files:
- Old test outputs
- Temporary files kept for reference
- Legacy scripts no longer needed

#### Best Practices for Claude

1. **New Scripts**: Always place automation scripts in `.claude/scripts/`
2. **Documentation**: Create architectural docs in `.claude/docs/`
3. **Temporary Files**: Use `.claude/archive/` for files that may be removed later
4. **Coverage**: Let coverage reports go to `.claude/coverage/` (gitignored)
5. **Large Test Files**: Use modularization script for files >30K bytes