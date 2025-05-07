package api

import (
	"log"
	"mxclone/internal/api/handlers"
	"mxclone/internal/api/middleware"
	v1 "mxclone/internal/api/v1"
	"mxclone/internal/api/version"
	"mxclone/pkg/logging"
	"mxclone/ports/input"
	"net/http"
	"os"
	"path/filepath"
)

// Server represents the API server
type Server struct {
	// Handlers
	dnsHandler   *handlers.DNSHandler
	dnsblHandler *handlers.DNSBLHandler
	// Add other handlers here

	// Middleware
	rateLimiter *middleware.RateLimiter
	logger      *middleware.Logger

	// API versioning
	versionedHandler *version.VersionedHandler
}

// NewServer creates a new API server with all dependencies injected
func NewServer(
	dnsService input.DNSPort,
	dnsblService input.DNSBLPort,
	// Add other services here
	logger *logging.Logger,
) *Server {
	server := &Server{
		dnsHandler:   handlers.NewDNSHandler(dnsService),
		dnsblHandler: handlers.NewDNSBLHandler(dnsblService),
		// Initialize other handlers
		rateLimiter:      middleware.NewRateLimiter(10, logger), // 10 requests per minute
		logger:           middleware.NewLogger(logger),
		versionedHandler: version.NewVersionedHandler(),
	}

	// Set up versioned API routes
	server.setupVersionedRoutes()

	return server
}

// setupVersionedRoutes sets up the versioned API routes
func (s *Server) setupVersionedRoutes() {
	// Create v1 router
	v1Router := v1.NewRouter(
		s.dnsHandler,
		s.dnsblHandler,
		// Other handlers would be added here
	)

	// Register v1 handler with versioned handler
	s.versionedHandler.RegisterHandler(version.V1, s.withMiddleware(v1Router.Handler()))
}

// Start starts the API server
func (s *Server) Start() error {
	// Serve static UI files
	uiDist := os.Getenv("UI_DIST_PATH")
	if uiDist == "" {
		uiDist = "./ui/dist"
	}
	fs := http.FileServer(http.Dir(uiDist))

	// Root handler to serve UI files and handle client-side routing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle versioned API requests
		if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
			s.versionedHandler.ServeHTTP(w, r)
			return
		}

		// Handle non-API requests
		if r.URL.Path == "/" || !isAPIPath(r.URL.Path) {
			// If the file exists, serve it; otherwise, serve index.html
			filePath := filepath.Join(uiDist, r.URL.Path)
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
			// Serve index.html for client-side routing
			http.ServeFile(w, r, filepath.Join(uiDist, "index.html"))
			return
		}

		// Fallback to default mux for legacy API paths
		defaultMux := http.DefaultServeMux
		defaultMux.ServeHTTP(w, r)
	})

	// Legacy (non-versioned) API endpoints for backward compatibility
	// These will eventually be deprecated and removed
	http.HandleFunc("/api/health", s.withMiddleware(s.handleHealth))
	http.HandleFunc("/api/dns", s.withMiddleware(s.dnsHandler.HandleDNSLookup))
	http.HandleFunc("/api/blacklist", s.withMiddleware(s.dnsblHandler.HandleDNSBLCheck))

	// Other API routes would be registered here for backward compatibility
	// http.HandleFunc("/api/smtp", s.withMiddleware(s.smtpHandler.HandleSMTPCheck))
	// ...

	log.Println("[api] Starting server on :8080, serving UI from", uiDist)
	return http.ListenAndServe(":8080", nil)
}

// isAPIPath checks if the URL path is an API endpoint
func isAPIPath(path string) bool {
	return len(path) >= 5 && path[:5] == "/api/"
}

// withMiddleware applies common middleware to handlers
func (s *Server) withMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	// Apply middleware in order: logging, rate limiting
	return s.logger.Log(s.rateLimiter.Limit(handler))
}

// handleHealth is a simple health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle versioned API requests
	if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
		s.versionedHandler.ServeHTTP(w, r)
		return
	}

	// Handle legacy API requests based on path
	switch r.URL.Path {
	case "/api/health":
		s.withMiddleware(s.handleHealth)(w, r)
	case "/api/dns":
		s.withMiddleware(s.dnsHandler.HandleDNSLookup)(w, r)
	case "/api/blacklist":
		s.withMiddleware(s.dnsblHandler.HandleDNSBLCheck)(w, r)
	// Other endpoints would be added here
	default:
		http.NotFound(w, r)
	}
}

// StartAPIServer creates and starts the API server
func StartAPIServer(
	dnsService input.DNSPort,
	dnsblService input.DNSBLPort,
	// Other services
	logger *logging.Logger,
) error {
	server := NewServer(
		dnsService,
		dnsblService,
		// Other services
		logger,
	)
	return server.Start()
}
