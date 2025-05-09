package v1

import (
	"context"
	"mxclone/internal/api/errors"
	"mxclone/internal/api/handlers"
	"mxclone/internal/api/middleware"
	"mxclone/internal/api/validation"
	"net/http"
)

// Router manages API v1 routing
type Router struct {
	mux                 *http.ServeMux
	dnsHandler          *handlers.DNSHandler
	dnsblHandler        *handlers.DNSBLHandler
	smtpHandler         *handlers.SMTPHandler
	emailAuthHandler    *handlers.EmailAuthHandler
	networkToolsHandler *handlers.NetworkToolsHandler
	docsHandler         *handlers.DocsHandler
	validator           *middleware.Validator
	jsonValidator       *validation.JSONValidator
	paramValidator      *validation.ParamValidator
	logger              *middleware.Logger
	rateLimiter         *middleware.RateLimiter
	errorHandler        *errors.ErrorHandler
}

// NewRouter creates a new v1 Router with all dependencies
func NewRouter(
	dnsHandler *handlers.DNSHandler,
	dnsblHandler *handlers.DNSBLHandler,
	smtpHandler *handlers.SMTPHandler,
	emailAuthHandler *handlers.EmailAuthHandler,
	networkToolsHandler *handlers.NetworkToolsHandler,
	docsHandler *handlers.DocsHandler,
	validator *middleware.Validator,
	jsonValidator *validation.JSONValidator,
	paramValidator *validation.ParamValidator,
	logger *middleware.Logger,
	rateLimiter *middleware.RateLimiter,
	errorHandler *errors.ErrorHandler,
) *Router {
	r := &Router{
		mux:                 http.NewServeMux(),
		dnsHandler:          dnsHandler,
		dnsblHandler:        dnsblHandler,
		smtpHandler:         smtpHandler,
		emailAuthHandler:    emailAuthHandler,
		networkToolsHandler: networkToolsHandler,
		docsHandler:         docsHandler,
		validator:           validator,
		jsonValidator:       jsonValidator,
		paramValidator:      paramValidator,
		logger:              logger,
		rateLimiter:         rateLimiter,
		errorHandler:        errorHandler,
	}
	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	// Documentation endpoint - GET method
	r.mux.HandleFunc("GET /docs/", r.docsHandler.HandleDocs)

	// DNS routes - POST method
	r.mux.HandleFunc("POST /dns", r.withValidation(r.dnsHandler.HandleDNSLookup, r.jsonValidator.ValidateDNSRequestJSON))

	// DNS route with path parameter
	r.mux.HandleFunc("POST /dns/{domain}", func(w http.ResponseWriter, req *http.Request) {
		domain := req.PathValue("domain")
		params := map[string]string{"domain": domain}
		valid, errs := r.paramValidator.ValidateDomainParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid domain parameter", errs)
			return
		}
		r.dnsHandler.HandleDNSLookup(w, req)
	})

	// Blacklist routes - POST method
	r.mux.HandleFunc("POST /blacklist", r.withValidation(r.dnsblHandler.HandleDNSBLCheck, r.jsonValidator.ValidateBlacklistRequestJSON))

	// Blacklist route with path parameter
	r.mux.HandleFunc("POST /blacklist/{target}", func(w http.ResponseWriter, req *http.Request) {
		target := req.PathValue("target")
		// Target can be either a domain or an IP address
		validDomain, _ := r.paramValidator.ValidateDomainParam(map[string]string{"domain": target})
		validIP, _ := r.paramValidator.ValidateIPParam(map[string]string{"ip": target})

		if !validDomain && !validIP {
			r.errorHandler.HandleValidationError(w, "Invalid target parameter (must be a valid domain or IP address)", nil)
			return
		}

		r.dnsblHandler.HandleDNSBLCheck(w, req)
	})

	// SMTP routes
	r.mux.HandleFunc("POST /smtp/connect/{host}", func(w http.ResponseWriter, req *http.Request) {
		host := req.PathValue("host")
		params := map[string]string{"host": host}
		valid, errs := r.paramValidator.ValidateHostParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid host parameter", errs)
			return
		}
		r.smtpHandler.HandleSMTPConnect(w, req)
	})

	r.mux.HandleFunc("POST /smtp/starttls/{host}", func(w http.ResponseWriter, req *http.Request) {
		host := req.PathValue("host")
		params := map[string]string{"host": host}
		valid, errs := r.paramValidator.ValidateHostParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid host parameter", errs)
			return
		}
		r.smtpHandler.HandleSMTPStartTLS(w, req)
	})

	r.mux.HandleFunc("POST /smtp/relay-test", r.withValidation(r.smtpHandler.HandleSMTPRelayTest, r.jsonValidator.ValidateSMTPRelayTestRequestJSON))

	// Email Authentication routes
	r.mux.HandleFunc("POST /auth/spf/{domain}", func(w http.ResponseWriter, req *http.Request) {
		domain := req.PathValue("domain")
		if domain == "" {
			r.errorHandler.HandleValidationError(w, "Domain parameter is required", nil)
			return
		}
		params := map[string]string{"domain": domain}
		valid, errs := r.paramValidator.ValidateDomainParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid domain parameter", errs)
			return
		}
		r.emailAuthHandler.HandleSPFCheck(w, req)
	})

	// New route for DKIM check without selector
	r.mux.HandleFunc("POST /auth/dkim/{domain}", func(w http.ResponseWriter, req *http.Request) {
		domain := req.PathValue("domain")
		params := map[string]string{"domain": domain}
		valid, errs := r.paramValidator.ValidateDomainParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid domain parameter", errs)
			return
		}
		r.emailAuthHandler.HandleDKIMCheck(w, req)
	})

	r.mux.HandleFunc("POST /auth/dkim/{domain}/{selector}", func(w http.ResponseWriter, req *http.Request) {
		domain := req.PathValue("domain")
		selector := req.PathValue("selector")
		params := map[string]string{"domain": domain, "selector": selector}
		valid, errs := r.paramValidator.ValidateDKIMParams(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid DKIM parameters", errs)
			return
		}
		r.emailAuthHandler.HandleDKIMCheck(w, req)
	})

	r.mux.HandleFunc("POST /auth/dmarc/{domain}", func(w http.ResponseWriter, req *http.Request) {
		domain := req.PathValue("domain")
		params := map[string]string{"domain": domain}
		valid, errs := r.paramValidator.ValidateDomainParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid domain parameter", errs)
			return
		}
		r.emailAuthHandler.HandleDMARCCheck(w, req)
	})

	// Network Tools routes
	r.mux.HandleFunc("POST /network/ping/{host}", func(w http.ResponseWriter, req *http.Request) {
		host := req.PathValue("host")
		params := map[string]string{"host": host}
		valid, errs := r.paramValidator.ValidateHostParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid host parameter", errs)
			return
		}
		ctx := context.WithValue(req.Context(), "host", host)
		r.networkToolsHandler.HandlePing(w, req.WithContext(ctx))
	})

	r.mux.HandleFunc("POST /network/traceroute/{host}", func(w http.ResponseWriter, req *http.Request) {
		host := req.PathValue("host")
		params := map[string]string{"host": host}
		valid, errs := r.paramValidator.ValidateHostParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid host parameter", errs)
			return
		}
		ctx := context.WithValue(req.Context(), "host", host)
		r.networkToolsHandler.HandleTraceroute(w, req.WithContext(ctx))
	})

	r.mux.HandleFunc("POST /network/whois/{domain}", func(w http.ResponseWriter, req *http.Request) {
		domain := req.PathValue("domain")
		params := map[string]string{"domain": domain}
		valid, errs := r.paramValidator.ValidateDomainParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid domain parameter", errs)
			return
		}
		ctx := context.WithValue(req.Context(), "domain", domain)
		r.networkToolsHandler.HandleWhois(w, req.WithContext(ctx))
	})

	// Async traceroute job endpoints
	r.mux.HandleFunc("POST /network/traceroute/{host}/async", func(w http.ResponseWriter, req *http.Request) {
		host := req.PathValue("host")
		params := map[string]string{"host": host}
		valid, errs := r.paramValidator.ValidateHostParam(params)
		if !valid {
			r.errorHandler.HandleValidationError(w, "Invalid host parameter", errs)
			return
		}
		ctx := context.WithValue(req.Context(), "host", host)
		r.networkToolsHandler.HandleTracerouteAsync(w, req.WithContext(ctx))
	})

	r.mux.HandleFunc("GET /network/traceroute/result/{jobId}", func(w http.ResponseWriter, req *http.Request) {
		jobId := req.PathValue("jobId")
		if jobId == "" {
			r.errorHandler.HandleValidationError(w, "jobId parameter is required", nil)
			return
		}
		ctx := context.WithValue(req.Context(), "jobId", jobId)
		r.networkToolsHandler.HandleTracerouteJobResult(w, req.WithContext(ctx))
	})

	// Health check endpoint - GET method
	r.mux.HandleFunc("GET /health", r.handleHealth)
	r.mux.HandleFunc("GET /health/", r.handleHealth)

	// Catch-all for 404s within the v1 router
	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		r.errorHandler.HandleNotFound(w, "Endpoint not found: "+req.URL.Path)
	})
}

// Handler returns an http.HandlerFunc that routes API v1 requests
func (r *Router) Handler() http.HandlerFunc {
	return r.mux.ServeHTTP
}

// handleHealth handles the health check endpoint
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","version":"v1"}`))
}

// withValidation applies JSON validation middleware to a handler
func (r *Router) withValidation(handler http.HandlerFunc, validateFunc func([]byte) (bool, map[string]interface{})) http.HandlerFunc {
	return r.validator.ValidateJSON(handler, validateFunc)
}
