# Tech Context

## Technologies Used
- Go (Golang)
- miekg/dns, net/http, standard library
- CLI: Cobra
- API: net/http, encoding/json
- UI: React (TypeScript) with Vite
- Docker (multi-stage builds)
- Kubernetes (YAML manifests for deployment)
- GitHub Actions (CI/CD)

## Development Setup
- Go modules for dependency management
- Standard Go project structure (cmd/, pkg/, internal/)
- Unit and integration tests in pkg/
- Documentation in docs/
- Frontend project in ui/ using npm
- Dockerfile and .dockerignore for containerized builds
- Kubernetes manifests in k8s/ for deployment
- GitHub Actions workflow in .github/workflows/ for automated build and deployment

## Technical Constraints
- Must run on Linux, Windows, macOS
- No external dependencies for core diagnostics (except miekg/dns)
- Secure by default (no InsecureSkipVerify, input validation)
- Docker image must be minimal and production-ready
- Kubernetes manifests must be portable and easy to customize

## Dependencies
- Go: miekg/dns, Cobra
- UI: React, Vite, react-router-dom, axios, @testing-library/react, cypress
- CI/CD: GitHub Actions
- Deployment: Docker, Kubernetes

## Tool Usage Patterns
- Worker pools for concurrency
- Caching for repeated lookups
- Centralized logging and error handling
- Vite dev server with proxy for API requests
- Multi-stage Docker builds for efficient images
- Kubernetes manifests for infrastructure-as-code
- GitHub Actions for automated build, test, and Docker image creation
