package api

import (
	"log"
	"mxclone/internal/api/errors"
	"mxclone/internal/api/handlers"
	"mxclone/internal/api/middleware"
	v1 "mxclone/internal/api/v1"
	"mxclone/internal/api/validation"
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
	dnsHandler          *handlers.DNSHandler
	dnsblHandler        *handlers.DNSBLHandler
	smtpHandler         *handlers.SMTPHandler
	emailAuthHandler    *handlers.EmailAuthHandler
	networkToolsHandler *handlers.NetworkToolsHandler
	docsHandler         *handlers.DocsHandler
	// Add other handlers here

	// Middleware
	rateLimiter *middleware.RateLimiter
	logger      *middleware.Logger
	validator   *middleware.Validator

	// Validators
	jsonValidator  *validation.JSONValidator
	paramValidator *validation.ParamValidator

	// Error handler
	errorHandler *errors.ErrorHandler

	// API versioning
	versionedHandler *version.VersionedHandler
}

// NewServer creates a new API server with all dependencies injected
func NewServer(
	dnsService input.DNSPort,
	dnsblService input.DNSBLPort,
	smtpService input.SMTPPort,
	emailAuthService input.EmailAuthPort,
	networkToolsService input.NetworkToolsPort,
	// Add other services here
	logger *logging.Logger,
) *Server {
	// Create error handler first so we can attach it to middleware
	errorHandler := errors.NewErrorHandler(logger)

	// Create and configure rate limiter
	rateLimiter := middleware.NewRateLimiter(logger)
	rateLimiter.WithErrorHandler(errorHandler)
	rateLimiter.Start() // Start the background cleanup routine

	server := &Server{
		dnsHandler:          handlers.NewDNSHandler(dnsService),
		dnsblHandler:        handlers.NewDNSBLHandler(dnsblService),
		smtpHandler:         handlers.NewSMTPHandler(smtpService),
		emailAuthHandler:    handlers.NewEmailAuthHandler(emailAuthService),
		networkToolsHandler: handlers.NewNetworkToolsHandler(networkToolsService),
		docsHandler:         handlers.NewDocsHandler(logger),
		// Initialize other handlers
		rateLimiter:      rateLimiter,
		logger:           middleware.NewLogger(logger),
		validator:        middleware.NewValidator(logger),
		jsonValidator:    validation.NewJSONValidator(),
		paramValidator:   validation.NewParamValidator(),
		errorHandler:     errorHandler,
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
		s.smtpHandler,
		s.emailAuthHandler,
		s.networkToolsHandler,
		s.docsHandler,
		// Other handlers would be added here
		s.validator,
		s.jsonValidator,
		s.paramValidator,
		s.logger,
		s.rateLimiter,
		s.errorHandler,
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

	mux := http.NewServeMux()

	// Handle versioned API requests
	mux.Handle("/api/", s.versionedHandler)

	// Legacy (non-versioned) API endpoints for backward compatibility
	mux.HandleFunc("GET /api/health", s.withMiddleware(s.handleHealth))
	mux.HandleFunc("POST /api/dns", s.withMiddleware(s.dnsHandler.HandleDNSLookup))
	mux.HandleFunc("POST /api/blacklist", s.withMiddleware(s.dnsblHandler.HandleDNSBLCheck))
	mux.HandleFunc("POST /api/smtp", s.withMiddleware(s.smtpHandler.HandleSMTPCheck))
	mux.HandleFunc("POST /api/email-auth", s.withMiddleware(s.emailAuthHandler.HandleEmailAuth))
	mux.HandleFunc("POST /api/network-tools", s.withMiddleware(s.networkToolsHandler.HandleNetworkTools))

	// Root handler to serve UI files and handle client-side routing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Attempt to serve the requested file if it exists and is not a directory
		filePath := filepath.Join(uiDist, r.URL.Path)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, filePath)
			return
		}
		// Fallback to serving index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(uiDist, "index.html"))
	})

	log.Println("[api] Starting server on :8080, serving UI from", uiDist)
	return http.ListenAndServe(":8080", mux)
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

// withValidation applies validation middleware to handlers
func (s *Server) withValidation(handler http.HandlerFunc, validateFunc func([]byte) (bool, map[string]interface{})) http.HandlerFunc {
	// Apply validation middleware to the handler
	return s.logger.Log(s.rateLimiter.Limit(s.validator.ValidateJSON(handler, validateFunc)))
}

// withParamValidation applies URL parameter validation middleware to handlers
func (s *Server) withParamValidation(handler http.HandlerFunc, validateFunc func(map[string]string) (bool, map[string]interface{})) http.HandlerFunc {
	// Apply validation middleware to the handler
	return s.logger.Log(s.rateLimiter.Limit(s.validator.ValidateParams(handler, validateFunc)))
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

	// Handle legacy API requests based on path and method
	switch {
	case r.URL.Path == "/api/health" && r.Method == http.MethodGet:
		s.withMiddleware(s.handleHealth)(w, r)
	case r.URL.Path == "/api/dns" && r.Method == http.MethodPost:
		s.withMiddleware(s.dnsHandler.HandleDNSLookup)(w, r)
	case r.URL.Path == "/api/blacklist" && r.Method == http.MethodPost:
		s.withMiddleware(s.dnsblHandler.HandleDNSBLCheck)(w, r)
	case r.URL.Path == "/api/smtp" && r.Method == http.MethodPost:
		s.withMiddleware(s.smtpHandler.HandleSMTPCheck)(w, r)
	case r.URL.Path == "/api/email-auth" && r.Method == http.MethodPost:
		s.withMiddleware(s.emailAuthHandler.HandleEmailAuth)(w, r)
	case r.URL.Path == "/api/network-tools" && r.Method == http.MethodPost:
		s.withMiddleware(s.networkToolsHandler.HandleNetworkTools)(w, r)
	// Other endpoints would be added here
	default:
		http.NotFound(w, r)
	}
}

// CreateServer creates and returns an HTTP server mux for testing
func CreateServer() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}

// StartAPIServer creates and starts the API server
func StartAPIServer(
	dnsService input.DNSPort,
	dnsblService input.DNSBLPort,
	smtpService input.SMTPPort,
	emailAuthService input.EmailAuthPort,
	networkToolsService input.NetworkToolsPort,
	// Other services
	logger *logging.Logger,
) error {
	server := NewServer(
		dnsService,
		dnsblService,
		smtpService,
		emailAuthService,
		networkToolsService,
		// Other services
		logger,
	)
	return server.Start()
}
