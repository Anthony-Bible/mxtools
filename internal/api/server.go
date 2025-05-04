package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mxclone/pkg/dns"
	"mxclone/pkg/dnsbl"
	"mxclone/pkg/emailauth"
	"mxclone/pkg/networktools"
	"mxclone/pkg/smtp"
	"mxclone/pkg/types"
	"mxclone/pkg/validation"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	rateLimit      = 10 // requests per minute per IP
	rateLimitStore = make(map[string][]time.Time)
	rateLimitMu    sync.Mutex
)

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		rateLimitMu.Lock()
		times := rateLimitStore[ip]
		now := time.Now()
		var recent []time.Time
		for _, t := range times {
			if now.Sub(t) < time.Minute {
				recent = append(recent, t)
			}
		}
		if len(recent) >= rateLimit {
			rateLimitMu.Unlock()
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
			return
		}
		recent = append(recent, now)
		rateLimitStore[ip] = recent
		rateLimitMu.Unlock()
		next(w, r)
	}
}

func StartAPIServer() {
	// Serve static UI files
	uiDist := os.Getenv("UI_DIST_PATH")
	if uiDist == "" {
		uiDist = "./ui/dist"
	}
	fs := http.FileServer(http.Dir(uiDist))

	// Serve index.html for non-API, non-file routes (for React Router)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		// Fallback to API handlers
		defaultMux := http.DefaultServeMux
		defaultMux.ServeHTTP(w, r)
	})

	http.HandleFunc("/api/health", rateLimitMiddleware(healthHandler))
	http.HandleFunc("/api/dns", rateLimitMiddleware(dnsHandler))
	http.HandleFunc("/api/blacklist", rateLimitMiddleware(blacklistHandler))
	http.HandleFunc("/api/smtp", rateLimitMiddleware(smtpHandler))
	http.HandleFunc("/api/auth", rateLimitMiddleware(authHandler))
	http.HandleFunc("/api/network/ping", rateLimitMiddleware(pingHandler))
	http.HandleFunc("/api/network/traceroute", rateLimitMiddleware(tracerouteHandler))
	http.HandleFunc("/api/network/whois", rateLimitMiddleware(whoisHandler))

	log.Println("[api] Starting server on :8080, serving UI from", uiDist)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[api] Server failed: %v", err)
	}
}

func isAPIPath(path string) bool {
	return len(path) >= 5 && path[:5] == "/api/"
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func dnsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req types.CheckRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	if err := validation.ValidateDomain(req.Target); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid domain"})
		return
	}
	result, err := dns.LookupAll(r.Context(), req.Target)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func blacklistHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req types.CheckRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	ip := req.Target
	if err := validation.ValidateIP(ip); err != nil {
		if derr := validation.ValidateDomain(ip); derr != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid IP or domain"})
			return
		}
		ips, err := net.LookupIP(ip)
		if err != nil || len(ips) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "could not resolve domain to IP"})
			return
		}
		ip = ips[0].String()
	}
	zones := []string{"bl.spamcop.net", "dnsbl.sorbs.net"}
	result := dnsbl.CheckMultipleBlacklists(r.Context(), ip, zones, 10*time.Second)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func smtpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req types.CheckRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	host := req.Target
	if err := validation.ValidateDomain(host); err != nil && validation.ValidateIP(host) != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid domain or IP"})
		return
	}
	result, err := smtp.CheckSMTP(r.Context(), host, nil, 10*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req types.CheckRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	domain := req.Target
	if err := validation.ValidateDomain(domain); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid domain"})
		return
	}
	result, err := emailauth.CheckEmailAuth(r.Context(), domain, 10*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req types.CheckRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	target := req.Target
	if err := validation.ValidateDomain(target); err != nil && validation.ValidateIP(target) != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid domain or IP"})
		return
	}
	result, err := networktools.PingWithPrivilegeCheck(r.Context(), target, 4, 5*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if result != nil {
		type pingResultWithMs struct {
			Target     string  `json:"target"`
			Sent       int     `json:"sent"`
			Received   int     `json:"received"`
			PacketLoss float64 `json:"packetLoss"`
			MinRTT     string  `json:"minRttMs"`
			MaxRTT     string  `json:"maxRttMs"`
			AvgRTT     string  `json:"avgRttMs"`
			Error      string  `json:"error,omitempty"`
		}
		json.NewEncoder(w).Encode(pingResultWithMs{
			Target:     result.Target,
			Sent:       result.Sent,
			Received:   result.Received,
			PacketLoss: result.PacketLoss,
			MinRTT:     fmt.Sprintf("%.3fms", float64(result.MinRTT)/float64(time.Millisecond)),
			MaxRTT:     fmt.Sprintf("%.3fms", float64(result.MaxRTT)/float64(time.Millisecond)),
			AvgRTT:     fmt.Sprintf("%.3fms", float64(result.AvgRTT)/float64(time.Millisecond)),
			Error:      result.Error,
		})
		return
	}
	json.NewEncoder(w).Encode(result)
}

func tracerouteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req types.CheckRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	target := req.Target
	if err := validation.ValidateDomain(target); err != nil && validation.ValidateIP(target) != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid domain or IP"})
		return
	}
	result, err := networktools.TracerouteWithPrivilegeCheck(r.Context(), target, 30, 5*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func whoisHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req types.CheckRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	target := req.Target
	if err := validation.ValidateDomain(target); err != nil && validation.ValidateIP(target) != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid domain or IP"})
		return
	}
	result, err := networktools.WhoisWithReferral(r.Context(), target, 10*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
