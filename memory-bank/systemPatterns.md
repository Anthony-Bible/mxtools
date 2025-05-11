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
- OpenAPI specification and schema validation

### 6. UI Layer (`ui/`)
- React/TypeScript frontend
- API client for communicating with backend
- Component-based UI architecture
- State management for diagnostic results
- Progressive updates for long-running operations (like traceroute)

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

### 7. Contract-First API Design
- OpenAPI specification serves as the contract between frontend and backend
- API endpoints implement the contract defined in the specification
- Strong schema validation ensures adherence to the contract

### 8. Progressive UI Updates Pattern
- Long-running operations provide incremental results
- Frontend polls for updates and renders partial data
- Improves user experience for operations like traceroute

## Component Relationships

### DNS Diagnostics Flow
```
CLI/API Request → DnsCmd/API Handler → DNSService Interface → DNSService Implementation 
                → DNS Repository Interface → DNS Repository Implementation 
                → External DNS Servers → Repository → Service → Output
```

### Email Authentication Flow
```
CLI/API Request → AuthCmd/API Handler → EmailAuthService Interface → EmailAuthService Implementation 
                → EmailAuth Repository Interface → EmailAuth Repository Implementation 
                → External DNS/Email Servers → Repository → Service → Output
```

### SMTP Testing Flow
```
CLI/API Request → SMTPCmd/API Handler → SMTPService Interface → SMTPService Implementation 
                → SMTP Repository Interface → SMTP Repository Implementation 
                → External SMTP Servers → Repository → Service → Output
```

### Web UI to API Flow
```
User Interaction → React Component → API Client → API Endpoint
                → Domain Service → Repository → External Service
                → Response Processing → UI State Update → UI Rendering
```

### Traceroute Progressive Updates Flow
```
User Request → API Endpoint → Background Job → Job Queue → Worker Process
             → Incremental Result Collection → Poll API → UI Update
             → Final Result or Timeout Handling → Complete UI Rendering
```

## Critical Implementation Paths

### DNS Lookup Path
1. User provides domain and record type
2. CLI/API validates input
3. DNS service performs lookup
4. Results converted to appropriate output format
5. Displayed to user/returned via API

### Blacklist Check Path
1. User provides IP or domain
2. CLI/API validates input
3. DNSBL service queries multiple blocklist providers
4. Results aggregated
5. Displayed to user/returned via API

### API Request Path
1. HTTP request to API endpoint
2. Request validation against OpenAPI schema
3. Route to appropriate handler
4. Handler calls domain service
5. Service performs operation
6. Results converted to JSON according to schema
7. HTTP response returned

### Progressive Traceroute Path
1. User initiates traceroute request
2. API endpoint creates background job
3. Job executor begins traceroute process
4. Each hop discovery updates job status
5. Frontend polls job status endpoint
6. Progressive UI updates as hops are discovered
7. Final result or timeout handling with partial results

## Key Technical Decisions

1. **Hexagonal Architecture**: Ensures separation of concerns and testability
2. **Go Language**: Provides performance, concurrency, and networking capabilities
3. **Cobra CLI Framework**: Structured command hierarchy and flag management
4. **Custom DI Container**: Lightweight dependency management
5. **Repository Pattern**: Isolates external dependencies
6. **REST API with JSON**: Standard, language-agnostic communication
7. **OpenAPI Specification**: Contract-first API design with strong typing
8. **React/TypeScript**: Type safety and component-based UI architecture
9. **Progressive Updates**: Improved UX for long-running operations
10. **Background Job System**: Handling long-running operations without blocking

## API Endpoint Design

The API follows a consistent RESTful pattern:

### Endpoint Structure
- All endpoints under `/api/v1/` prefix
- Resource-based paths (e.g., `/api/v1/dns`, `/api/v1/network/ping`)
- HTTP POST for operations
- JSON request/response bodies
- Consistent error format across all endpoints

### Common API Features
- Strong input validation
- Detailed error responses
- Rate limiting
- Standardized response structures
- Content negotiation
- CORS support for frontend integration

### API Documentation
- OpenAPI specifications in `docs/` directory
- Detailed schema definitions for all data types
- Query parameter documentation
- Response format documentation
- Status code documentation

## Distributed Job Status Management Pattern

To support reliable async job tracking (such as progressive traceroute) in distributed/Kubernetes deployments, MXClone uses a shared storage backend for job status. The default implementation is Redis, but the JobStore interface allows for other backends (e.g., PostgreSQL, etcd).

- All application replicas access and update job status via the shared store.
- Enables horizontal scaling and consistent polling/progressive updates.
- See `docs/shared-storage.md` for configuration, deployment, and security details.

### Example Flow
```
User Request → API Endpoint → Background Job → JobStore (Redis) → Worker Process
             → Incremental Result Collection → Poll API → UI Update
             → Final Result or Timeout Handling → Complete UI Rendering
```
