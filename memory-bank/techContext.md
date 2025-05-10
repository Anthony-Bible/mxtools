# Technical Context

## Technologies Used

### Backend (Go)
- **Go Language**: Core programming language (using modules)
- **Cobra**: CLI framework for structured commands
- **net/http**: Standard library HTTP server
- **miekg/dns**: Low-level DNS operations beyond standard library
- **json**: JSON encoding/decoding
- **context**: Context management for timeout and cancellation
- **testing**: Standard testing library

### Web UI (TypeScript/React)
- **TypeScript**: Typed JavaScript for frontend development
- **React**: UI component library
- **Vite**: Frontend build tooling
- **npm**: Package management
- **CSS3/HTML5**: Frontend styling and markup
- **Cypress**: End-to-end testing

### API Documentation
- **OpenAPI 3.1**: API specification format
- **OpenAPI DevTools**: Tools for generating and validating specifications
- **Swagger UI**: Interactive API documentation visualization

### Infrastructure
- **Docker**: Containerization
- **Kubernetes**: Container orchestration (optional deployment)
- **GitHub Actions**: CI/CD automation

## Development Setup

### Local Development
- **Go**: Version 1.18+ (per go.mod)
- **IDE**: GoLand, VS Code, or similar with Go support
- **Node.js**: Version 18+ for Web UI development
- **Git**: Version control

### Build Process
- Go build for backend CLI and API server
- npm build (Vite) for Web UI
- Docker build for containerized deployment

### Testing Strategy
- Unit testing with Go's standard testing package
- Integration testing for repository implementations
- API testing with HTTP client
- UI testing with Cypress
- End-to-end testing of complete flows

## Technical Constraints

### Performance Constraints
- DNS lookups subject to network latency
- SMTP connections affected by external server response times
- Rate limiting for outbound requests to prevent abuse
- Concurrent request handling to improve throughput
- Background job system for long-running operations

### Security Constraints
- Input validation for all user-provided values
- Output sanitization
- API rate limiting to prevent abuse
- No storage of sensitive credentials
- CORS configuration for API security

### Compatibility Constraints
- Go 1.18+ compatibility
- Browser compatibility for Web UI (modern browsers)
- DNS RFC compliance
- SMTP protocol standard compliance
- OpenAPI 3.1 specification compliance

### Deployment Constraints
- Local machine installation
- Docker container deployment
- Kubernetes deployment (optional)

## Dependencies Management

### Backend Dependencies
- Managed through Go modules (go.mod)
- Minimal external dependencies philosophy
- Preference for standard library where feasible

### Frontend Dependencies
- Managed through npm (package.json)
- Regular security audits
- TypeScript for type safety

## Tool Usage Patterns

### DNS Tools
- Standard DNS lookup using Go's net package
- Advanced DNS lookup using miekg/dns package
- DNS over HTTPS/TLS when available

### Email Authentication Tools
- SPF record validation and interpretation
- DKIM record lookup and validation
- DMARC policy checking

### Network Tools
- Ping using ICMP echo
- Traceroute using UDP or ICMP with progressive updates
- WHOIS lookups against regional registries

### SMTP Tools
- SMTP connection testing
- SMTP transaction simulation
- TLS certificate validation

## API Documentation Approach

### OpenAPI Specification
- Multiple specification files in development:
  - `openapi-devtools-spec (3).json`: 8 core endpoints
  - `spec4.son`: Expanded 13 endpoints with query parameters
- Detailed schema definitions for all request/response types
- Comprehensive type validation

### API Endpoint Structure
- Consistent RESTful endpoint design
- Resource-based paths with `/api/v1/` prefix
- POST operations for all diagnostic tools
- Query parameters for configuration options
- Detailed error responses

### API Response Formats
- Standardized JSON structures
- Consistent field naming conventions
- Strong typing for all fields
- Clear error messages
- Appropriate HTTP status codes

## Integration Points

### External Services
- Public DNS servers
- SMTP servers
- DNSBL providers
- WHOIS servers
- Regional Internet Registries

### Internal Components
- CLI commands ↔ Domain services
- API handlers ↔ Domain services
- Domain services ↔ Repositories
- Repositories ↔ External services
- Frontend ↔ API endpoints

## Development Workflow

### Code Organization
- Domain-driven package structure
- Clear separation between interfaces and implementations
- Consistent error handling patterns
- Extensive comments and documentation

### API-First Development
- OpenAPI specification as the source of truth
- Schema validation against specification
- Contract-first development approach
- Consistent API patterns across endpoints

### Build and Deployment
- Local development build
- Container image build
- Kubernetes deployment (optional)

### Versioning
- Semantic versioning for releases
- API versioning for stability
