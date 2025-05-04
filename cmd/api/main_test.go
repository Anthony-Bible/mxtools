package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/health", nil)
	rw := httptest.NewRecorder()
	healthHandler(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}
}

func TestDNSHandler_BadRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/dns", bytes.NewBuffer([]byte("bad json")))
	rw := httptest.NewRecorder()
	dnsHandler(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rw.Code)
	}
}

func TestDNSHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/dns", nil)
	rw := httptest.NewRecorder()
	dnsHandler(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rw.Code)
	}
}

func TestDNSHandler_Validation(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"target": "not_a_domain"})
	req := httptest.NewRequest("POST", "/api/dns", bytes.NewBuffer(body))
	rw := httptest.NewRecorder()
	dnsHandler(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rw.Code)
	}
}
