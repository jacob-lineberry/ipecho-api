// Package middleware provides HTTP middleware for IP extraction and proxy handling.
package middleware

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type contextKey string

const ClientIPKey contextKey = "clientIP"

// CloudRunClientIP returns middleware that extracts the real client IP
// from X-Forwarded-For header set by Cloud Run's load balancer.
func CloudRunClientIP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var clientIP string

			// try X-Forwarded-For first (set by GCR)
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				// take the leftmost IP (original client)
				if raw := strings.TrimSpace(strings.Split(xff, ",")[0]); raw != "" {
					if ip, err := netip.ParseAddr(raw); err == nil {
						clientIP = ip.String()
					}
				}
			}

			// fallback: parse RemoteAddr if X-Forwarded-For is blank
			if clientIP == "" {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					host = r.RemoteAddr // no port present, use as-is
				}
				if ip, err := netip.ParseAddr(host); err == nil {
					clientIP = ip.String()
				} else {
					clientIP = host // last resort: use as-is
				}
			}
			ctx := context.WithValue(r.Context(), ClientIPKey, clientIP)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
