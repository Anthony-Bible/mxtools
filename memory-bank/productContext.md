# Product Context

## Why This Project Exists
MXToolbox is a critical tool for diagnosing domain, DNS, and email issues, but is proprietary. This project provides an open-source alternative, empowering users with transparency, extensibility, and self-hosting.

## Problems Solved
- Lack of open, self-hosted alternatives to MXToolbox
- Need for integrated diagnostics in one tool
- Desire for automation and API access
- Need for modern, scriptable, and UI-driven workflows
- Need for a system that can be extended and customized
- Lack of containerized, cloud-native deployment options

## How It Should Work
- Users can run diagnostics via CLI, API, or Web UI
- Results are clear, actionable, and exportable
- System is responsive, secure, and easy to operate
- Hexagonal architecture enables clean separation of concerns:
  - Domain core contains pure business logic
  - Ports define clear interfaces for inputs and outputs
  - Adapters implement these interfaces for real-world interactions
- Containerized deployment for easy self-hosting
- Kubernetes support for scalable, cloud-native operation

## User Experience Goals
- Fast, intuitive diagnostics
- Clear error messages and guidance
- Consistent experience across CLI, API, and UI
- Accessible to both technical and less-technical users
- Responsive design that works on all devices
- Professional styling with modern blue color scheme
- Visual feedback for all user interactions
- Support for common network diagnostic tools (DNS, SMTP, Blacklist, Auth, Ping, Traceroute, WHOIS)

## Deployment and Operations
- Docker images for containerized deployment
- Kubernetes manifests for orchestration
- CI/CD pipeline for automated builds and deployments
- Simple configuration for self-hosting
- Minimal dependencies for easier maintenance

## Extensibility
- Hexagonal architecture makes it easy to:
  - Add new diagnostic tools by implementing domain interfaces
  - Replace infrastructure components without changing business logic
  - Add new UI features without affecting core functionality
  - Integrate with existing systems via custom adapters
