package version

import (
	"encoding/json"
	"net/http"
	"regexp"
)

// APIVersion represents an API version
type APIVersion string

const (
	// V1 represents API version 1
	V1 APIVersion = "v1"

	// Future versions would be added here
	// V2 APIVersion = "v2"
)

// VersionMatcher is responsible for matching and extracting version information from URLs
type VersionMatcher struct {
	pattern *regexp.Regexp
}

// NewVersionMatcher creates a new VersionMatcher
func NewVersionMatcher() *VersionMatcher {
	// Match paths like /api/v1/resource or /api/v2/resource
	pattern := regexp.MustCompile(`^/api/(v[0-9]+)/(.+)$`)
	return &VersionMatcher{
		pattern: pattern,
	}
}

// ExtractVersion extracts the API version and resource path from a URL path
func (vm *VersionMatcher) ExtractVersion(path string) (APIVersion, string, bool) {
	matches := vm.pattern.FindStringSubmatch(path)
	if matches == nil || len(matches) != 3 {
		return "", "", false
	}

	return APIVersion(matches[1]), matches[2], true
}

// VersionInfo represents API version information
type VersionInfo struct {
	Version         string `json:"version"`
	ReleaseDate     string `json:"releaseDate"`
	LatestVersion   string `json:"latestVersion"`
	DeprecationDate string `json:"deprecationDate,omitempty"`
}

// VersionedHandler provides versioning for API handlers
type VersionedHandler struct {
	handlers map[APIVersion]http.HandlerFunc
	matcher  *VersionMatcher

	// Default handlers
	notFoundHandler  http.HandlerFunc
	methodNotAllowed http.HandlerFunc
	versionNotFound  http.HandlerFunc
}

// NewVersionedHandler creates a new VersionedHandler
func NewVersionedHandler() *VersionedHandler {
	vh := &VersionedHandler{
		handlers: make(map[APIVersion]http.HandlerFunc),
		matcher:  NewVersionMatcher(),
	}

	// Set up default handlers
	vh.notFoundHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Resource not found",
			"code":  http.StatusNotFound,
		})
	}

	vh.methodNotAllowed = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Method not allowed",
			"code":  http.StatusMethodNotAllowed,
		})
	}

	vh.versionNotFound = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":         "API version not found",
			"code":          http.StatusNotFound,
			"versions":      []string{"v1"},
			"latestVersion": "v1",
		})
	}

	return vh
}

// RegisterHandler registers a handler for a specific API version
func (vh *VersionedHandler) RegisterHandler(version APIVersion, handler http.HandlerFunc) {
	vh.handlers[version] = handler
}

// ServeHTTP implements the http.Handler interface
func (vh *VersionedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set content type for all API responses
	w.Header().Set("Content-Type", "application/json")

	// Extract version from URL path
	version, resourcePath, ok := vh.matcher.ExtractVersion(r.URL.Path)
	if !ok {
		vh.notFoundHandler(w, r)
		return
	}

	// Find handler for the version
	handler, found := vh.handlers[version]
	if !found {
		vh.versionNotFound(w, r)
		return
	}

	// Update request URL path to contain only the resource path (without version)
	r.URL.Path = "/" + resourcePath

	// Call the handler
	handler(w, r)
}
