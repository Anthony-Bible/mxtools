# System Patterns

## Architecture
- Modular Go monorepo: cmd/, pkg/, internal/, docs/, memory-bank/
- Core engine/orchestrator coordinates diagnostics and aggregation
- CLI, API, and UI layers interface with core engine
- Containerized deployment using Docker (multi-stage build: Go backend + React UI)
- Kubernetes manifests for infrastructure-as-code deployment
- CI/CD pipeline via GitHub Actions for automated builds and deployments

## Key Technical Decisions
- Use Go for concurrency, performance, and ecosystem
- Standard library for basic networking; miekg/dns for advanced DNS
- Worker pools for parallel checks
- Caching for DNS/blacklist
- Rate limiting and input validation throughout
- Containerization for consistent builds and deployments
- Kubernetes for scalable, portable infrastructure
- GitHub Actions for automated CI/CD

## Design Patterns
- Command pattern for CLI/API commands
- Factory pattern for diagnostics
- Adapter pattern for integrating third-party libraries
- Strategy pattern for result formatting
- Multi-stage Docker build for efficient image creation
- Infrastructure-as-code for deployment (Kubernetes YAML)

## Component Relationships
- CLI/API/UI -> Engine -> Diagnostics (DNS, Blacklist, SMTP, etc.)
- Diagnostics -> Output, Logging, Caching
- Docker image encapsulates backend and UI for deployment
- Kubernetes manages service lifecycle and scaling
- CI/CD pipeline automates build, test, and deploy

## Critical Implementation Paths
- Orchestrator dispatches and aggregates checks
- Error handling and logging are centralized
- Extensible for new diagnostics and output formats
- Automated build and deployment via CI/CD
