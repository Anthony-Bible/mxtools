# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MXTools is a comprehensive Go-based email and network diagnostics tool (MXToolbox clone) that provides DNS lookups, blacklist checks, SMTP diagnostics, email authentication validation, and network tools via CLI, API, and Web UI.

## Build and Development Commands

### Go Backend
```bash
# Build CLI binary
go build -o mxclone .

# Build API server  
go build -o api ./cmd/api

# Run from source
go run main.go <command>

# Run tests
go test ./...

# Run specific test
go test -v ./pkg/dns/
```

### Web UI (React + TypeScript)
```bash
# Install dependencies
cd ui && npm install

# Development server
cd ui && npm run dev

# Build for production
cd ui && npm run build

# Run tests
cd ui && npm test

# E2E tests
cd ui && npm run cypress:run
```

### Docker
```bash
# Build container (multi-stage: Go backend + React UI)
docker build -t mxtools .

# Run container
docker run -p 8080:8080 mxtools
```

## Architecture

### CLI Structure (Cobra Framework)
- **Entry Point**: `main.go` → `cmd/mxclone/commands.Execute()`
- **Root Command**: `cmd/mxclone/commands/root.go` (handles config, orchestration)
- **Subcommands**: 
  - `dns` - DNS lookups (A, AAAA, MX, TXT, etc.)
  - `blacklist` - DNSBL checks against spam databases
  - `smtp` - SMTP server diagnostics
  - `auth` - Email authentication (SPF, DKIM, DMARC)
  - `network` - Network tools (ping, traceroute, whois)
  - `api` - HTTP API server

### API Server
- **Location**: `internal/api/server.go`
- **Port**: 8080 (serves both API endpoints and React UI)
- **Rate Limiting**: 10 requests/minute per IP
- **Endpoints**: `/api/dns`, `/api/blacklist`, `/api/smtp`, `/api/auth`, `/api/network/*`

### Configuration System
- **Location**: `internal/config/config.go`
- **Uses**: Viper for configuration management
- **Sources**: YAML files, environment variables (MXCLONE_ prefix), defaults
- **Key Settings**: Worker pools, timeouts, DNS resolvers, DNSBL zones, cache TTLs

### Package Organization
- `cmd/` - Application entry points (CLI and API binaries)
- `internal/` - Private application logic (config, API server, cache)
- `pkg/` - Reusable packages for core functionality
- `ui/` - React TypeScript frontend
- `k8s/` - Kubernetes deployment manifests

### Key Patterns

#### Orchestration Engine
- Worker pool pattern for concurrent operations in `pkg/orchestration/`
- Context-based timeout handling throughout
- Centralized result aggregation and error handling

#### Input Validation
- All user inputs are sanitized and validated
- Structured error types with proper error wrapping
- Timeout and retry mechanisms for network operations

#### Container Deployment
- Multi-stage Dockerfile builds both Go backend and React UI
- Kubernetes deployment with CAP_NET_ADMIN for network tools
- 2 replicas with proper health checks

## Dependencies

Key external dependencies:
- `github.com/miekg/dns` - Advanced DNS operations
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

## Development Notes

- Uses Go 1.23 with standard project layout
- React UI built with Vite and TypeScript
- E2E testing with Cypress
- Memory bank documentation in `memory-bank/` directory
- Deployment configurations include both Docker and Kubernetes manifests