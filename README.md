# mxclone

## Overview

`mxclone` is a command-line utility written in Go designed for diagnosing email server configurations and performing various network checks. It provides tools for inspecting DNS records (including MX, SPF, DKIM, DMARC), checking against DNS blacklists (DNSBLs), verifying SMTP server connectivity, and running common network diagnostics like ping, traceroute, and whois.

## Features

*   **DNS Checks:** Query various DNS record types (MX, A, TXT, SPF, DKIM, DMARC).
*   **DNS Blacklist (DNSBL) Checks:** Check domains or IPs against common DNS blacklists.
*   **Email Authentication:** Verify SPF, DKIM, and DMARC records.
*   **SMTP Checks:** Test SMTP server connectivity and capabilities.
*   **Network Tools:** 
    * Ping: Test connectivity with round-trip time measurement
    * Traceroute: Trace network path with progressive updates as hops are discovered
    * WHOIS: Look up domain registration information
*   **Health Checks:** Basic health endpoint checks.
*   **Structured Output:** Provides clear output for diagnostics.
*   **Caching:** Caches results to speed up repeated queries (configurable).
*   **Rate Limiting:** Built-in rate limiting for external services.
*   **Web UI:** Modern React/TypeScript interface with real-time updates for long-running operations.

## Architecture

*   **Language:** Go
*   **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
*   **Architecture Pattern:** Hexagonal Architecture (ports and adapters)
    * `domain`: Core business logic and domain models
    * `ports`: Interfaces defining input and output boundaries
    * `adapters`: Implementations of the interfaces (primary and secondary)
*   **Structure:** Follows standard Go project layout.
    *   `cmd/mxclone`: Main application entry point and command definitions.
    *   `internal`: Internal application logic (caching, configuration, API).
    *   `pkg`: Reusable packages for core functionalities (DNS, SMTP, Network Tools, etc.).
    *   `ui`: React TypeScript web interface for browser-based diagnostics.

## Prerequisites

*   Go (version specified in `go.mod`, or latest stable recommended)

## Building

1.  Clone the repository (if applicable).
2.  Navigate to the project root directory (`/home/anthony/GolandProjects/mxclone`).
3.  Build the executable:
    ```bash
    go build -o mxclone ./cmd/mxclone
    ```

## Running

Execute the compiled binary followed by the desired command and flags:

```bash
./mxclone <command> [arguments/flags]
```

To see the list of available commands and options:

```bash
./mxclone --help
```

### Available Commands

Based on the project structure, the main commands likely include:

*   `auth`: Perform email authentication checks (SPF, DKIM, DMARC).
*   `blacklist`: Check against DNS blacklists.
*   `dns`: Perform various DNS lookups.
*   `health`: Run health checks.
*   `network`: Access network tools (ping, traceroute, whois).
*   `smtp`: Perform SMTP server checks.

Use `.<command> --help` for specific command usage (e.g., `./mxclone dns --help`).

## API Usage

The API server runs on port 8080 by default. All endpoints accept POST requests with a JSON body:

Example request to `/api/dns`:

```
curl -X POST http://localhost:8080/api/dns -H 'Content-Type: application/json' -d '{"target": "example.com"}'
```

All endpoints accept a JSON body with at least a `target` field. Example endpoints:
- `/api/dns` — DNS diagnostics
- `/api/blacklist` — Blacklist checks
- `/api/smtp` — SMTP diagnostics
- `/api/auth` — Email authentication
- `/api/network/ping` — ICMP ping
- `/api/network/traceroute` — Traceroute
- `/api/network/whois` — WHOIS lookup

Responses are JSON objects with diagnostic results or error messages.

Rate limiting: Each client IP is limited to 10 requests per minute.

## API Features

The application provides a comprehensive REST API for all diagnostic tools:

*   **DNS Endpoints:** Query various DNS record types.
*   **DNSBL Endpoints:** Check against multiple blacklists.
*   **SMTP Endpoints:** Test email server connectivity.
*   **Network Tools:**
    * `GET /api/v1/network/ping/{host}`: Ping a host
    * `POST /api/v1/network/traceroute/{host}`: Start an async traceroute job
    * `GET /api/v1/network/traceroute/result/{jobId}`: Poll for progressive traceroute results
    * `GET /api/v1/network/whois/{domain}`: WHOIS lookup
*   **Health Endpoint:** Basic health check.

All API endpoints are documented with OpenAPI/Swagger.

## Running the Web UI

The Web UI is located in the `ui/` directory and provides a modern frontend for diagnostics via your browser.

### Prerequisites
- Node.js (v18 or later)

### Steps
1. Navigate to the `ui/` directory:
   ```bash
   cd ui
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the development server:
   ```bash
   npm run dev
   ```
4. Open your browser and go to [http://localhost:5173](http://localhost:5173)

The Web UI will connect to the backend API (by default at `http://localhost:8080/api`). Ensure the API server is running for full functionality.

### Building for Production
To build the Web UI for production deployment:
```bash
npm run build
```
The static files will be output to the `ui/dist/` directory. Serve these with your preferred static file server or integrate with your backend.

## Configuration

The application uses [Viper](https://github.com/spf13/viper) for configuration management. Configuration can be provided through:

1. Configuration file (`config.yaml) in one of the following locations:
   - Current working directory
   - `$HOME/.mxclone/`
   - `/etc/mxclone/`

2. Environment variables:
   - All configuration options can be set with environment variables prefixed with `MXCLONE_`
   - Example: `MXCLONE_DNS_TIMEOUT=5s`

### Core Configuration Options

```yaml
# General settings
worker_count: 10          # Number of worker goroutines for concurrent operations
log_level: "info"         # Logging level (debug, info, warn, error)
cache_dir: "/tmp/mxclone" # Directory to store cache files

# DNS settings
dns_timeout: 5            # DNS query timeout in seconds
dns_retries: 2            # Number of retries for failed DNS queries
dns_resolvers:            # Custom DNS resolvers to use
  - "8.8.8.8:53"
  - "1.1.1.1:53"
dns_cache_ttl: 300        # DNS cache TTL in seconds (5 minutes)

# Blacklist settings
blacklist_zones:          # DNSBL providers to check
  - "zen.spamhaus.org"
  - "bl.spamcop.net"
  - "dnsbl.sorbs.net"
blacklist_timeout: 10     # Blacklist query timeout in seconds
blacklist_cache_ttl: 1800 # Blacklist cache TTL in seconds (30 minutes)

# SMTP settings
smtp_timeout: 10          # SMTP connection timeout in seconds
smtp_ports:               # SMTP ports to test
  - 25
  - 465
  - 587
```

## Distributed Job Status Management & Shared Storage

For distributed deployments (e.g., Kubernetes), mxclone supports job status tracking via a shared storage backend. By default, Redis is used for this purpose. This enables reliable async job tracking (such as progressive traceroute) across multiple replicas.

- To enable Redis, set `job_store_type: "redis"` in your config or `MXCLONE_JOB_STORE_TYPE=redis` in your environment.
- See [`docs/shared-storage.md`](docs/shared-storage.md) for full configuration, deployment, and security details.

All configuration options can be overridden with environment variables by using the prefix `MXCLONE_` followed by the option name in uppercase. For example, `MXCLONE_LOG_LEVEL=debug`.

## Contributing

Contributions to mxclone are welcome! Here's how you can help:

1. **Report bugs or request features**: Open an issue describing what you found or what you'd like to see.

2. **Submit pull requests**: Make sure to:
   - Follow the existing code style and architecture patterns
   - Add appropriate tests for your changes
   - Update documentation as needed
   - Keep pull requests focused on a single concern

3. **Development workflow**:
   - Fork the repository
   - Create a feature branch (`git checkout -b feature/amazing-feature`)
   - Commit your changes (`git commit -m 'Add amazing feature'`)
   - Push to the branch (`git push origin feature/amazing-feature`)
   - Open a Pull Request

4. **Code guidelines**:
   - Maintain the hexagonal architecture pattern
   - Write unit tests for domain logic
   - Write integration tests for adapters
   - Follow Go best practices and conventions

By contributing, you agree to license your contributions under the same license as this project.

## License

This project is licensed under the MIT License - see below for details:

```
MIT License

Copyright (c) 2025 MXClone Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
