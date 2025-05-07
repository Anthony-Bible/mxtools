# System Patterns

## Hexagonal Architecture (Ports and Adapters)
- Core architecture pattern implementing clear separation of concerns
- Domain core contains pure business logic with no external dependencies
- Ports define interface contracts between layers
  - Input ports: Use cases that application offers to the outside world
  - Output ports: External dependencies required by the domain
- Adapters implement the interfaces defined by ports
  - Primary adapters: Drive the application (API, CLI, UI)
  - Secondary adapters: Driven by the application (Repositories, External services)
- Benefits: Testability, maintainability, flexibility to change infrastructure

## Dependency Injection
- Used for wiring components together while maintaining loose coupling
- Implemented in internal/di/container.go
- Enables easy swapping of implementation details without changing business logic
- Makes testing easier with mock implementations

## Module Organization
- Domain modules organized by business capability:
  - DNS - DNS lookup functionality
  - DNSBL - DNS blacklist checking
  - EmailAuth - SPF, DKIM, DMARC verification
  - NetworkTools - Ping, traceroute, WHOIS
  - SMTP - SMTP server testing
- Each module contains:
  - Domain models in domain/[module]/model.go
  - Port definitions in ports/input/[module]_port.go and ports/output/[module]_repository.go
  - Primary adapter implementation in adapters/primary/[module]_service.go
  - Secondary adapter implementation in adapters/secondary/[module]_repository.go
  - Implementation details in pkg/[module]/

## Command Pattern (CLI)
- Cobra-based hierarchical command structure in cmd/mxclone/commands/
- Each module has its own command group
- Commands interact with the domain through the port interfaces

## Repository Pattern
- Used for data access abstraction
- Defined in ports/output/[module]_repository.go
- Implemented in adapters/secondary/[module]_repository.go
- Provides clean separation between business logic and data access

## Service Pattern
- Core application services defined as input ports in ports/input/[module]_port.go
- Implemented in adapters/primary/[module]_service.go
- Orchestrates business logic and interacts with repositories

## Worker Pool Pattern
- Implemented in pkg/orchestration/worker_pool.go
- Used for concurrent execution of DNS lookups and network diagnostics
- Manages resource usage and prevents overwhelming target systems

## Rate Limiting
- Implemented in pkg/ratelimit/
- Protects target systems from excessive requests
- Configurable per operation type

## Error Handling
- Centralized error types in pkg/errors/
- Consistent error wrapping and propagation
- Differentiation between business errors and technical errors

## Validation
- Input validation in pkg/validation/
- Ensures data integrity before processing
- Prevents security issues from malformed input

## Testing Patterns
- Unit tests for business logic
- Integration tests for external dependencies
- E2E tests with Cypress for UI flows
- Test doubles (mocks/stubs) for isolating components
