package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"mxclone/pkg/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// APIEndpointDoc represents documentation for a specific API endpoint
type APIEndpointDoc struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Description string            `json:"description"`
	Parameters  []APIParameterDoc `json:"parameters,omitempty"`
	Request     interface{}       `json:"request,omitempty"`
	Response    interface{}       `json:"response,omitempty"`
	Example     string            `json:"example,omitempty"`
}

// APIParameterDoc represents documentation for an API parameter
type APIParameterDoc struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // "path", "query", "header", "body"
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	Example     interface{} `json:"example,omitempty"`
}

// APIGroupDoc represents documentation for a group of related API endpoints
type APIGroupDoc struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Endpoints   []APIEndpointDoc `json:"endpoints"`
}

// APIDocs represents the full API documentation
type APIDocs struct {
	Version     string        `json:"version"`
	Description string        `json:"description"`
	BaseURL     string        `json:"baseUrl"`
	Groups      []APIGroupDoc `json:"groups"`
}

// DocsHandler handles API documentation requests
type DocsHandler struct {
	logger     *logging.Logger
	docs       *APIDocs
	specPath   string
	openAPIDoc *v3.Document
}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler(logger *logging.Logger) *DocsHandler {
	specPath := filepath.Join("docs", "openapi.yaml")
	handler := &DocsHandler{
		logger:   logger,
		specPath: specPath,
	}

	// Load OpenAPI spec
	doc, err := handler.loadOpenAPISpec()
	if err != nil {
		logger.Error("Failed to load OpenAPI spec. Documentation will be unavailable.", "error", err)
		// No fallback to hardcoded docs, h.docs remains nil
	} else {
		handler.openAPIDoc = doc
		handler.docs = handler.convertOpenAPIToAPIDocs()
	}

	return handler
}

// HandleDocs handles the API documentation endpoint
func (h *DocsHandler) HandleDocs(w http.ResponseWriter, r *http.Request) {
	// Check if docs is available
	if h.docs == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "API documentation is not available",
		})
		return
	}

	// Check if a specific group is requested
	group := r.URL.Query().Get("group")
	if group != "" {
		h.handleGroupDocs(w, r, group)
		return
	}

	// Return all documentation if no specific group requested
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.docs)
}

// handleGroupDocs returns documentation for a specific group
func (h *DocsHandler) handleGroupDocs(w http.ResponseWriter, r *http.Request, group string) {
	group = strings.ToLower(group)

	// Check if docs is available before trying to access h.docs.Groups
	if h.docs == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "API documentation is not available",
		})
		return
	}

	// Find the requested group
	for _, g := range h.docs.Groups {
		if strings.ToLower(g.Name) == group {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(g)
			return
		}
	}

	// Group not found
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "Documentation group not found",
	})
}

// loadOpenAPISpec loads and parses OpenAPI specification from file
func (h *DocsHandler) loadOpenAPISpec() (*v3.Document, error) {
	// Check if OpenAPI spec file exists
	if _, err := os.Stat(h.specPath); os.IsNotExist(err) {
		return nil, errors.New("OpenAPI specification file not found")
	}

	// Read the file contents
	data, err := os.ReadFile(h.specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI specification: %w", err)
	}

	// Parse OpenAPI document
	document, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}

	// Get the model
	v3Model, errs := document.BuildV3Model()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to build OpenAPI model: %v", errs)
	}

	return &v3Model.Model, nil
}

// convertOpenAPIToAPIDocs converts OpenAPI spec to internal API docs format
func (h *DocsHandler) convertOpenAPIToAPIDocs() *APIDocs {
	doc := h.openAPIDoc

	// Create base API documentation
	apiDocs := &APIDocs{
		Version:     doc.Info.Version,
		Description: doc.Info.Description,
		BaseURL:     getBaseURL(doc),
		Groups:      []APIGroupDoc{},
	}

	// Group endpoints by tags
	tagGroups := make(map[string][]APIEndpointDoc)
	tagDescriptions := make(map[string]string)

	// Extract tag descriptions
	for _, tag := range doc.Tags {
		tagDescriptions[tag.Name] = tag.Description
	}

	// Process paths and operations
	pathKeys := doc.Paths.PathItems.KeysFromNewest()
	for path := range pathKeys {
		pathItem, _ := doc.Paths.PathItems.Get(path)
		processPath(path, pathItem, tagGroups)
	}

	// Create API groups from tags
	for tag, endpoints := range tagGroups {
		description := tagDescriptions[tag]
		if description == "" {
			description = fmt.Sprintf("%s operations", tag)
		}

		apiDocs.Groups = append(apiDocs.Groups, APIGroupDoc{
			Name:        tag,
			Description: description,
			Endpoints:   endpoints,
		})
	}

	return apiDocs
}

// processPath processes an OpenAPI path and adds endpoints to tag groups
func processPath(path string, pathItem *v3.PathItem, tagGroups map[string][]APIEndpointDoc) {
	// Process each HTTP method
	methods := map[string]*v3.Operation{
		"GET":    pathItem.Get,
		"POST":   pathItem.Post,
		"PUT":    pathItem.Put,
		"DELETE": pathItem.Delete,
		"PATCH":  pathItem.Patch,
	}

	for method, operation := range methods {
		if operation != nil {
			endpoint := createEndpointDoc(path, method, operation)

			// Add endpoint to each tag group
			for _, tag := range operation.Tags {
				if _, ok := tagGroups[tag]; !ok {
					tagGroups[tag] = []APIEndpointDoc{}
				}
				tagGroups[tag] = append(tagGroups[tag], endpoint)
			}
		}
	}
}

// createEndpointDoc creates an API endpoint doc from OpenAPI operation
func createEndpointDoc(path string, method string, operation *v3.Operation) APIEndpointDoc {
	endpoint := APIEndpointDoc{
		Path:        path,
		Method:      method,
		Description: operation.Description,
		Parameters:  []APIParameterDoc{},
	}

	// Add parameters
	for _, param := range operation.Parameters {
		apiParam := APIParameterDoc{
			Name:        param.Name,
			In:          param.In,
			Description: param.Description,
			Required:    param.Required != nil && *param.Required,
		}

		// Extract parameter type
		if param.Schema != nil {
			schema := param.Schema.Schema()
			if schema != nil && len(schema.Type) > 0 {
				apiParam.Type = schema.Type[0]
				exampleVal := getSchemaExample(schema)
				if exampleVal != nil {
					apiParam.Example = exampleVal
				}
			}
		}

		endpoint.Parameters = append(endpoint.Parameters, apiParam)
	}

	// Process request body
	if operation.RequestBody != nil && operation.RequestBody.Content != nil {
		contentKeys := operation.RequestBody.Content.KeysFromNewest()
		for contentType := range contentKeys {
			mediaType, _ := operation.RequestBody.Content.Get(contentType)

			if contentType == "application/json" && mediaType.Schema != nil {
				// Extract example from schema
				schema := mediaType.Schema.Schema()
				if schema != nil {
					endpoint.Request = extractSchemaExample(schema)
				}
			}
		}
	}

	// Process responses
	if operation.Responses != nil && operation.Responses.Codes != nil {
		statusCodes := operation.Responses.Codes.KeysFromNewest()
		for statusCode := range statusCodes {
			response, _ := operation.Responses.Codes.Get(statusCode)

			// Check if it's a success response (2xx)
			if strings.HasPrefix(statusCode, "2") && response.Content != nil {
				contentTypes := response.Content.KeysFromNewest()
				for contentType := range contentTypes {
					mediaType, _ := response.Content.Get(contentType)

					if contentType == "application/json" && mediaType.Schema != nil {
						schema := mediaType.Schema.Schema()
						if schema != nil {
							endpoint.Response = extractSchemaExample(schema)
							break // Just use the first successful response with JSON content
						}
					}
				}
				if endpoint.Response != nil {
					break // Found a response, no need to check other status codes
				}
			}
		}
	}

	// Create example curl command
	endpoint.Example = createCurlExample(path, method, endpoint.Parameters, endpoint.Request)

	return endpoint
}

// extractSchemaExample extracts example data from schema
func extractSchemaExample(schema *base.Schema) interface{} {
	if schema.Example != nil {
		return schema.Example
	}

	// Create example based on schema type
	if len(schema.Type) > 0 {
		schemaType := schema.Type[0]
		switch schemaType {
		case "object":
			if schema.Properties != nil {
				example := make(map[string]interface{})
				propKeys := schema.Properties.KeysFromNewest()
				for name := range propKeys {
					prop, _ := schema.Properties.Get(name)

					propSchema := prop.Schema()
					if propSchema != nil {
						example[name] = getSchemaExample(propSchema)
					}
				}
				return example
			}
		case "array":
			if schema.Items != nil {
				// Get schema from items
				items := schema.Items.A
				if items != nil {
					itemSchema := items.Schema()
					if itemSchema != nil {
						// Return an array with a single example item
						return []interface{}{getSchemaExample(itemSchema)}
					}
				}
			}
		}

		return getDefaultExampleForType(schemaType)
	}

	return nil
}

// getSchemaExample gets an example value from a schema
func getSchemaExample(schema *base.Schema) interface{} {
	if schema.Example != nil {
		return schema.Example
	}

	if len(schema.Type) > 0 {
		return getDefaultExampleForType(schema.Type[0])
	}

	return nil
}

// getDefaultExampleForType returns a default example for a given type
func getDefaultExampleForType(typeName string) interface{} {
	switch typeName {
	case "string":
		return "example"
	case "integer":
		return 42
	case "number":
		return 42.0
	case "boolean":
		return true
	case "array":
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
	default:
		return nil
	}
}

// getBaseURL extracts base URL from OpenAPI document
func getBaseURL(doc *v3.Document) string {
	if len(doc.Servers) > 0 && doc.Servers[0].URL != "" {
		// Extract path part from URL
		url := doc.Servers[0].URL
		parts := strings.Split(url, "/")
		if len(parts) > 3 { // Has path components
			return strings.Join(parts[3:], "/")
		}
	}
	return "/api/v1" // Default base URL
}

// createCurlExample creates an example curl command
func createCurlExample(path string, method string, params []APIParameterDoc, body interface{}) string {
	baseURL := "http://localhost:8080/api/v1"

	// Replace path parameters with example values
	pathWithParams := path
	for _, param := range params {
		if param.In == "path" && param.Example != nil {
			pathWithParams = strings.Replace(
				pathWithParams,
				fmt.Sprintf("{%s}", param.Name),
				fmt.Sprintf("%v", param.Example),
				-1,
			)
		}
	}

	curl := fmt.Sprintf("curl -X %s %s%s", method, baseURL, pathWithParams)

	// Add query parameters
	queryParams := []string{}
	for _, param := range params {
		if param.In == "query" && param.Example != nil {
			queryParams = append(queryParams, fmt.Sprintf("%s=%v", param.Name, param.Example))
		}
	}

	if len(queryParams) > 0 {
		curl += "?" + strings.Join(queryParams, "&")
	}

	// Add request body if POST/PUT/PATCH
	if (method == "POST" || method == "PUT" || method == "PATCH") && body != nil {
		bodyJSON, err := json.Marshal(body)
		if err == nil {
			curl += fmt.Sprintf(" -d '%s'", string(bodyJSON))
		}
	}

	return curl
}
