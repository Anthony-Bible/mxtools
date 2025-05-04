# Tech Context

## Technologies Used
- Go (Golang)
- miekg/dns, net/http, standard library
- CLI: Cobra
- API: net/http, encoding/json
- UI: (planned) React or Vite-based frontend

## Development Setup
- Go modules for dependency management
- Standard Go project structure (cmd/, pkg/, internal/)
- Unit and integration tests in pkg/
- Documentation in docs/

## Technical Constraints
- Must run on Linux, Windows, macOS
- No external dependencies for core diagnostics (except miekg/dns)
- Secure by default (no InsecureSkipVerify, input validation)

## Dependencies
- miekg/dns for advanced DNS
- Cobra for CLI
- (Planned) React/Vite for UI

## Tool Usage Patterns
- Worker pools for concurrency
- Caching for repeated lookups
- Centralized logging and error handling
