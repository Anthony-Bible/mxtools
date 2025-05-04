# System Patterns

## Architecture
- Modular Go monorepo: cmd/, pkg/, internal/, docs/, memory-bank/
- Core engine/orchestrator coordinates diagnostics and aggregation
- CLI, API, and UI layers interface with core engine

## Key Technical Decisions
- Use Go for concurrency, performance, and ecosystem
- Standard library for basic networking; miekg/dns for advanced DNS
- Worker pools for parallel checks
- Caching for DNS/blacklist
- Rate limiting and input validation throughout

## Design Patterns
- Command pattern for CLI/API commands
- Factory pattern for diagnostics
- Adapter pattern for integrating third-party libraries
- Strategy pattern for result formatting

## Component Relationships
- CLI/API/UI -> Engine -> Diagnostics (DNS, Blacklist, SMTP, etc.)
- Diagnostics -> Output, Logging, Caching

## Critical Implementation Paths
- Orchestrator dispatches and aggregates checks
- Error handling and logging are centralized
- Extensible for new diagnostics and output formats
