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

## Configuration

(Optional: Add details about configuration file locations or environment variables if applicable, potentially managed via `internal/config`).

## Contributing

(Optional: Add contribution guidelines if this is an open project).

## License

(Optional: Specify the project's license, e.g., MIT, Apache 2.0).
