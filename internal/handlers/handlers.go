// Package handlers provide HTTP handlers for the ipecho-api service.
// It includes endpoints for returning client IP addresses in plaintext and JSON formats.
package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"net/netip"

	ipmw "github.com/jacob-lineberry/ipecho-api/internal/middleware"
)

// clientIP retrieves the client IP from request context.
// falls back to parsing RemoteAddr if context value is missing.
func clientIP(r *http.Request) string {
	// try context first (set by CloudRunClientIP middleware)
	if v := r.Context().Value(ipmw.ClientIPKey); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}

	// graceful fallback: parse RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr // no port present, use as-is
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		return ip.String()
	}
	return host
}

// PlainIP returns the client's IP address as plaintext.
// this is the primary endpoint for cURL users.
//
// Example:
//
//	curl -4 https://ipecho.dev
//	203.0.113.42
func PlainIP(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(ip + "\n"))
}

// JSONIP returns the client's IP address as JSON.
//
// Example:
//
//	curl -4 https://ipecho.dev/json
//	{"ip":"203.0.113.42"}
func JSONIP(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(struct {
		IP string `json:"ip"`
	}{IP: ip})
}

// Health returns a simple healthcheck response.
// Used by Cloud Run for liveness/readiness probes.
//
// Example:
//
//	curl https://ipecho.dev/health
//	ok
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
