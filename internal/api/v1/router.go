package v1

import (
	"mxclone/internal/api/handlers"
	"net/http"
)

// Router manages API v1 routing
type Router struct {
	dnsHandler   *handlers.DNSHandler
	dnsblHandler *handlers.DNSBLHandler
	// Other handlers would be added here
}

// NewRouter creates a new v1 Router with all dependencies
func NewRouter(
	dnsHandler *handlers.DNSHandler,
	dnsblHandler *handlers.DNSBLHandler,
	// Other handlers would be parameters here
) *Router {
	return &Router{
		dnsHandler:   dnsHandler,
		dnsblHandler: dnsblHandler,
		// Other handlers would be initialized here
	}
}

// Handler returns an http.Handler that routes API v1 requests
func (r *Router) Handler() http.HandlerFunc {
	// Create a router function that handles all v1 routes
	return func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		switch {
		case path == "/dns" || path == "/dns/":
			r.dnsHandler.HandleDNSLookup(w, req)
		case path == "/blacklist" || path == "/blacklist/":
			r.dnsblHandler.HandleDNSBLCheck(w, req)
		case path == "/health" || path == "/health/":
			r.handleHealth(w, req)
		// Additional routes would be added here
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Not found","code":404}`))
		}
	}
}

// handleHealth handles the health check endpoint
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","version":"v1"}`))
}
