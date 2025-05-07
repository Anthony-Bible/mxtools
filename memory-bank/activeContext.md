# Active Context

## Current Work Focus

The project is currently focused on implementing the core CLI and domain logic for the MXClone tool. This includes:

1. Building out the command-line interface structure using Cobra
2. Implementing domain services for DNS, SMTP, blacklist, and network tools
3. Creating the primary and secondary adapters for external service communication
4. Setting up the dependency injection system

## Recent Changes

Recent development has concentrated on:

1. Establishing the CLI command structure with subcommands for various diagnostic tools
2. Implementing the DNS lookup functionality with support for different record types
3. Setting up the dependency injection container for service management
4. Creating validation utilities for user inputs
5. Implementing output formatting options (text/JSON)

## Next Steps

Immediate next steps include:

1. Complete implementation of all CLI commands
   - Finish SMTP connection testing
   - Complete blacklist checking functionality
   - Implement email authentication verification
   - Add network tools (ping, traceroute, whois)

2. Begin API server implementation
   - Create REST API handlers mapping to CLI functionality
   - Implement OpenAPI documentation
   - Add rate limiting for API endpoints
   - Set up API authentication

3. Start Web UI development
   - Set up React/TypeScript environment
   - Create UI components for different diagnostics
   - Implement API communication layer
   - Design result visualization components

## Active Decisions and Considerations

### Architecture Decisions
1. Using ports and adapters (hexagonal) architecture to separate domain logic from external dependencies
2. Implementing a custom dependency injection container for simplicity rather than using a third-party solution
3. Adopting a clear separation between domain models and external data formats

### Implementation Approach
1. Focus on CLI functionality first to establish core domain logic
2. Reuse the same domain services across CLI, API, and Web UI
3. Prioritize robust error handling and input validation
4. Implement comprehensive testing at all layers

### Performance Considerations
1. Managing concurrent DNS and SMTP requests efficiently
2. Implementing timeouts for external service calls
3. Using connection pooling where appropriate
4. Considering caching strategies for repetitive lookups

## Important Patterns and Preferences

### Code Organization
- Group by feature domain rather than technical layer
- Keep interfaces in separate `ports` packages
- Use consistent naming conventions across the project
- Maintain clear separation between primary and secondary adapters

### Error Handling
- Domain-specific error types
- Consistent error wrapping pattern
- Detailed error messages for CLI/API responses
- Graceful degradation when partial failures occur

### Testing Approach
- Unit tests for domain logic
- Integration tests for adapters
- Mocked dependencies for isolation
- Table-driven tests for comprehensive coverage

## Learnings and Project Insights

### Technical Insights
- DNS lookup implementations need to handle various edge cases like NXDOMAIN, timeouts, etc.
- SMTP diagnostic tools must accommodate different server behaviors and security settings
- Rate limiting is essential for DNSBL checks to avoid being blocked by providers

### Project Management
- Breaking down features into independent, testable components aids parallel development
- Documenting interfaces early helps establish clear boundaries
- Consistent validation patterns simplify command implementation
