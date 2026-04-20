package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

func TestIPAllowlist_EmptyListPassesThrough(t *testing.T) {
	handler := middleware.IPAllowlist([]string{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIPAllowlist_NilListPassesThrough(t *testing.T) {
	handler := middleware.IPAllowlist(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIPAllowlist_AllowsMatchingCIDR(t *testing.T) {
	handler := middleware.IPAllowlist([]string{"10.0.0.0/8"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.1.2.3:5678"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIPAllowlist_BlocksNonMatchingIP(t *testing.T) {
	handler := middleware.IPAllowlist([]string{"10.0.0.0/8"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:5678"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIPAllowlist_SupportsBareIPAddress(t *testing.T) {
	handler := middleware.IPAllowlist([]string{"192.168.1.100"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:9999"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.101:9999"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)
}

func TestIPAllowlist_MultipleCIDRs(t *testing.T) {
	cidrs := []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	handler := middleware.IPAllowlist(cidrs)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name     string
		ip       string
		expected int
	}{
		{"10.x allowed", "10.5.5.5:1234", http.StatusOK},
		{"172.16.x allowed", "172.16.1.1:1234", http.StatusOK},
		{"192.168.x allowed", "192.168.100.1:1234", http.StatusOK},
		{"public IP blocked", "8.8.8.8:1234", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestIPAllowlist_UsesXForwardedFor_WithTrustedProxy(t *testing.T) {
	middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig([]string{"1.2.3.4"}))
	defer middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig(nil))

	handler := middleware.IPAllowlist([]string{"10.0.0.0/8"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIPAllowlist_IgnoresXForwardedFor_WithoutTrustedProxy(t *testing.T) {
	middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig(nil))

	handler := middleware.IPAllowlist([]string{"10.0.0.0/8"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIPAllowlist_SkipsInvalidCIDRs(t *testing.T) {
	handler := middleware.IPAllowlist([]string{"not-a-cidr", "10.0.0.0/8"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.1.1.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
