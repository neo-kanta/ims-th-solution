package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// imsClient is a thin HTTP client for the IMS REST API. The MCP server is a
// read-only window onto the SAME endpoints the dashboard uses, so every
// figure it returns is authoritative and already permission-scoped by the
// caller's token. The MCP server never touches the database directly.
type imsClient struct {
	baseURL string
	http    *http.Client
}

func newIMSClient(baseURL string) *imsClient {
	return &imsClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// imsResponse is the outcome of one upstream GET. Body is the verbatim
// response payload — we never reshape financial numbers.
type imsResponse struct {
	Status int
	Body   []byte
}

// get performs an authenticated GET against the IMS REST API. The bearer
// token is the end user's own JWT, forwarded out-of-band from the chat
// backend via MCP _meta — so all existing permission and data-scoping
// middleware applies exactly as it would for a dashboard request.
func (c *imsClient) get(ctx context.Context, token, path string) (*imsResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("missing caller auth token")
	}
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call IMS API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
	if err != nil {
		return nil, fmt.Errorf("read IMS response: %w", err)
	}
	return &imsResponse{Status: resp.StatusCode, Body: body}, nil
}
