# System Patterns

## Architecture Overview

The MXClone system follows a layered, hexagonal architecture pattern (also known as ports and adapters architecture) to ensure modularity, testability, and separation of concerns.

```
┌───────────────┐        ┌───────────────┐        ┌───────────────┐
│  Interfaces   │        │    Domain     │        │  Infrastructure│
│  (CLI/API/UI) │◄─────► │    Logic     │◄─────► │    & External  │
│               │        │               │        │    Services    │
└───────────────┘        └───────────────┘        └───────────────┘
```

## Key Architectural Components

### 1. Command Layer (`cmd/mxclone/`)
- Entry points for application (CLI commands)
- Uses Cobra framework for command structure
- Routes commands to appropriate domain services
- Handles input parsing and output formatting

### 2. Domain Layer
- Core business logic for DNS, SMTP, Email Auth, etc.
- Domain models representing key entities
- Service interfaces (ports) defining core operations
- Pure business logic independent of external dependencies

### 3. Ports Layer (`ports/`)
- Input ports: Service interfaces used by command layers
- Output ports: Repository interfaces for external dependencies
- Defines clear boundaries between layers

### 4. Adapters Layer (`adapters/`)
- Primary adapters: Implement service interfaces
- Secondary adapters: Implement repository interfaces
- Converts between external systems and domain models

### 5. Infrastructure Layer (`internal/`, `pkg/`)
- Configuration management
- Dependency injection
- API server setup
- Shared utilities, validation, error handling

## Design Patterns in Use

### 1. Dependency Injection
- Container-based DI system in `internal/di/`
- Services and repositories registered at startup
- Enables testability through mock injection

### 2. Repository Pattern
- Abstracts data access through repository interfaces
- Repositories handle external system communication
- Enables swapping implementations (e.g., in-memory vs network)

### 3. Strategy Pattern
- Multiple implementations of core interfaces (e.g., DNS lookup strategies)
- Runtime selection of appropriate strategy

### 4. Command Pattern
- CLI commands encapsulate specific use cases
- Each command maps to domain service operations

### 5. Factory Pattern
- Creation of complex objects via factories
- Particularly used in DI container

### 6. Adapter Pattern
- Converts external service responses to domain models

## Component Relationships

### DNS Diagnostics Flow
```
CLI/API Request → DnsCmd → DNSService Interface → DNSService Implementation 
                → DNS Repository Interface → DNS Repository Implementation 
                → External DNS Servers → Repository → Service → Output
```

### Email Authentication Flow
```
CLI/API Request → AuthCmd → EmailAuthService Interface → EmailAuthService Implementation 
                → EmailAuth Repository Interface → EmailAuth Repository Implementation 
                → External DNS/Email Servers → Repository → Service → Output
```

### SMTP Testing Flow
```
CLI/API Request → SMTPCmd → SMTPService Interface → SMTPService Implementation 
                → SMTP Repository Interface → SMTP Repository Implementation 
                → External SMTP Servers → Repository → Service → Output
```

## Critical Implementation Paths

### DNS Lookup Path
1. User provides domain and record type
2. CLI validates input
3. DNS service performs lookup
4. Results converted to appropriate output format
5. Displayed to user

### Blacklist Check Path
1. User provides IP or domain
2. CLI validates input
3. DNSBL service queries multiple blocklist providers
4. Results aggregated
5. Displayed to user

### API Request Path
1. HTTP request to API endpoint
2. Request validation
3. Route to appropriate handler
4. Handler calls domain service
5. Service performs operation
6. Results converted to JSON
7. HTTP response returned

## Key Technical Decisions

1. **Hexagonal Architecture**: Ensures separation of concerns and testability
2. **Go Language**: Provides performance, concurrency, and networking capabilities
3. **Cobra CLI Framework**: Structured command hierarchy and flag management
4. **Custom DI Container**: Lightweight dependency management
5. **Repository Pattern**: Isolates external dependencies
6. **REST API with JSON**: Standard, language-agnostic communication
