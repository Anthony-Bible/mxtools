# API Architecture Documentation

## Hexagonal Architecture in the API Layer

The MXClone API implements hexagonal architecture (also known as ports and adapters) to maintain clean separation between domain logic and infrastructure concerns. This document explains how the API layer is structured and how it integrates with the rest of the application.

## Overview

The hexagonal architecture organizes the application into concentric layers:

1. **Domain Core** - Contains business logic and entities without external dependencies
2. **Port Interfaces** - Define contracts for inputs and outputs (domain APIs)
3. **Adapters** - Implement the port interfaces to connect with external systems

Our API implementation is a **primary adapter** that drives the application by calling the input ports.

## API Layer Structure

The API layer is organized as follows:

```
internal/api/
├── handlers/            # HTTP handlers (primary adapters)
│   ├── dns_handler.go   # DNS lookup handler
│   ├── dnsbl_handler.go # Blacklist check handler
│   └── ...              # Other handlers
├── middleware/          # HTTP middleware (cross-cutting concerns)
│   └── middleware.go    # Rate limiting and logging middleware
├── models/              # API request/response models
│   └── models.go        # Model definitions and domain-to-API mappings
├── validation/          # Request validation
│   └── validation.go    # Validation logic for API requests
├── v1/                  # API version 1 implementation
│   └── router.go        # Router for v1 endpoints
├── version/             # API versioning support
│   └── version.go       # Version extraction and routing
└── server.go            # Main API server implementation
```

## Dependency Flow

One of the key benefits of hexagonal architecture is controlling the direction of dependencies:

```
[API Layer (Primary Adapter)] → [Input Ports] → [Domain Core]
                                                     ↓
                                [Output Ports] ← [Secondary Adapters]
```

In our implementation, the API handlers (primary adapters) depend on the input port interfaces, not concrete implementations. This ensures:

1. The API layer doesn't know or care about the implementation details
2. We can easily swap implementations as long as they implement the interface
3. The domain logic doesn't depend on the API layer (dependency inversion)

## API Handler Example

Here's how a typical API handler follows hexagonal architecture principles:

```go
// DNSHandler encapsulates handlers for DNS operations
type DNSHandler struct {
    dnsService input.DNSPort // Using the interface (port) instead of direct implementation
}

// HandleDNSLookup handles DNS lookup requests
func (h *DNSHandler) HandleDNSLookup(w http.ResponseWriter, r *http.Request) {
    // Parse request...
    
    // Use the port interface to call domain logic
    result, err := h.dnsService.LookupAll(r.Context(), req.Target)
    
    // Map domain result to API response and send...
}
```

## API Models

To maintain clean separation between the domain and API layer, we have dedicated API models that map to and from domain models:

```go
// FromDNSResult converts a domain DNS result to an API response
func FromDNSResult(result *dns.DNSResult) *DNSResponse {
    // Map domain-specific model to API response model
    return &DNSResponse{
        Records: result.Records,
        Timing:  result.Timing.String(),
    }
}
```

This prevents domain model changes from directly affecting the API contract, providing a stable API for consumers.

## Versioning

The API uses a versioned structure to ensure backward compatibility and allow for future evolution:

1. Each API version has its own router in a dedicated package (e.g., `v1/router.go`)
2. The version matcher extracts version information from the URL
3. Requests are routed to the appropriate version handler

This allows us to evolve the API without breaking existing clients and maintain multiple API versions simultaneously.

## Middleware as Adapters

Middleware components (like rate limiting and logging) are implemented as proper adapters that follow the same hexagonal principles:

1. They have a clear, single responsibility
2. They depend on abstractions, not concrete implementations
3. They are composable and can be chained together

## Testing

The hexagonal architecture greatly simplifies testing the API layer:

1. We can mock the input ports to test API handlers without calling real services
2. We can test different error conditions by configuring the mocks appropriately
3. Integration tests can verify that the entire request-response flow works correctly

## Benefits Realized

Our hexagonal architecture implementation in the API layer provides several benefits:

1. **Separation of concerns** - API handling logic is separate from business logic
2. **Testability** - We can easily test the API layer with mocks
3. **Flexibility** - We can change implementation details without affecting the API
4. **Maintainability** - Clear boundaries make the code easier to understand
5. **Evolvability** - Versioning support allows the API to evolve without breaking clients

## Future Enhancements

Potential enhancements to the API architecture include:

1. More comprehensive request validation with JSON Schema
2. API documentation generation directly from code
3. Automated client library generation
4. Additional API versions with new features while maintaining backward compatibility