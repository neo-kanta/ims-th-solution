package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

// IPAllowlist returns middleware that restricts access to the given CIDR ranges.
func IPAllowlist(allowedCIDRs []string) func(http.Handler) http.Handler {
	var nets []*net.IPNet
	for _, cidr := range allowedCIDRs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		nets = append(nets, ipNet)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(nets) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			clientIPStr := GetClientIP(r)
			clientIP := net.ParseIP(clientIPStr)
			if clientIP == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "unable to determine client IP"})
				return
			}

			for _, ipNet := range nets {
				if ipNet.Contains(clientIP) {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "access denied: IP not in allowlist"})
		})
	}
}
