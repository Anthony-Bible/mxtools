# mxclone

## Overview

`mxclone` is a command-line utility written in Go designed for diagnosing email server configurations and performing various network checks. It provides tools for inspecting DNS records (including MX, SPF, DKIM, DMARC), checking against DNS blacklists (DNSBLs), verifying SMTP server connectivity, and running common network diagnostics like ping, traceroute, and whois.

## Features

*   **DNS Checks:** Query various DNS record types (MX, A, TXT, SPF, DKIM, DMARC).
*   **DNS Blacklist (DNSBL) Checks:** Check domains or IPs against common DNS blacklists.
*   **Email Authentication:** Verify SPF, DKIM, and DMARC records.
*   **SMTP Checks:** Test SMTP server connectivity and capabilities.
*   **Network Tools:** Perform ping, traceroute, and whois lookups.
*   **Health Checks:** Basic health endpoint checks.
*   **Structured Output:** Provides clear output for diagnostics.
*   **Caching:** Caches results to speed up repeated queries (configurable).
*   **Rate Limiting:** Built-in rate limiting for external services.

## Architecture

*   **Language:** Go
*   **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
*   **Structure:** Follows standard Go project layout (`cmd`, `internal`, `pkg`).
    *   `cmd/mxclone`: Main application entry point and command definitions.
    *   `internal`: Internal application logic (caching, configuration).
    *   `pkg`: Reusable packages for core functionalities (DNS, SMTP, Network Tools, etc.).

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

(Optional: Add details about configuration file locations or environment variables if applicable, potentially managed via `internal/config`).

## Contributing

(Optional: Add contribution guidelines if this is an open project).

## License

(Optional: Specify the project's license, e.g., MIT, Apache 2.0).
