package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// IPAllowlist returns middleware that restricts access to the given CIDR ranges.
// If allowedCIDRs is empty, no restriction is applied (passthrough).
func IPAllowlist(allowedCIDRs []string) func(http.Handler) http.Handler {
	var nets []*net.IPNet
	for _, cidr := range allowedCIDRs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		// Support bare IP addresses (e.g., "192.168.1.1" -> "192.168.1.1/32")
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue // skip invalid CIDRs; log in production
		}
		nets = append(nets, ipNet)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no allowlist configured, pass through
			if len(nets) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			clientIP := parseIP(extractClientIP(r))
			if clientIP == nil {
				httputil.Forbidden(w, "unable to determine client IP")
				return
			}

			for _, ipNet := range nets {
				if ipNet.Contains(clientIP) {
					next.ServeHTTP(w, r)
					return
				}
			}

			httputil.Forbidden(w, "access denied: IP not in allowlist")
		})
	}
}

func parseIP(s string) net.IP {
	// Strip port if present (e.g., "192.168.1.1:54321")
	host, _, err := net.SplitHostPort(s)
	if err != nil {
		host = s
	}
	return net.ParseIP(host)
}
