// Package anthropic is the ChatProvider implementation backed by the
// Anthropic Messages API streaming endpoint.
//
// Slice A: no tools — only text generation. The adapter is the only place in
// the chat module that knows about Anthropic-specific event names or JSON
// shapes; the rest of the codebase consumes the canonical ChatStreamReader.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// Defaults verified against https://platform.claude.com/docs/en/api/messages-streaming
// on 2026-06-05. The streaming surface is stable since 2023-06-01; the model
// id changes with each Claude release.
const (
	defaultBaseURL    = "https://api.anthropic.com"
	defaultAPIVersion = "2023-06-01"
	defaultModel      = "claude-haiku-4-5-20251001"
	defaultMaxTokens  = 1024
)

// Config configures the Anthropic adapter.
type Config struct {
	APIKey     string        // required
	BaseURL    string        // default: https://api.anthropic.com
	APIVersion string        // default: 2023-06-01
	Model      string        // default: claude-haiku-4-5-20251001
	Timeout    time.Duration // request connect timeout; the stream itself is long-lived. Default 30s.
}

// Adapter implements service.ChatProvider against the Anthropic Messages API.
type Adapter struct {
	cfg    Config
	client *http.Client
}

// New constructs the adapter. APIKey is required.
func New(cfg Config) (*Adapter, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("anthropic: APIKey is required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = defaultAPIVersion
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Adapter{
		cfg: cfg,
		client: &http.Client{
			// No overall timeout — the stream is long-lived. The per-request
			// dial/header timeout is enforced by the transport below.
			Transport: &http.Transport{
				ResponseHeaderTimeout: cfg.Timeout,
				IdleConnTimeout:       90 * time.Second,
				MaxIdleConnsPerHost:   4,
			},
		},
	}, nil
}

// Name returns the canonical provider id.
func (a *Adapter) Name() valueobject.ProviderID { return valueobject.ProviderAnthropic }

// Generate opens a streaming Messages request and returns a reader the agent
// loop drives via Next/Result/Err/Close.
func (a *Adapter) Generate(ctx context.Context, req service.ChatRequest) (service.ChatStreamReader, error) {
	body, err := a.buildRequestBody(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: build request: %w", err)
	}

	url := strings.TrimRight(a.cfg.BaseURL, "/") + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic: new request: %w", err)
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("accept", "text/event-stream")
	httpReq.Header.Set("anthropic-version", a.cfg.APIVersion)
	httpReq.Header.Set("x-api-key", a.cfg.APIKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: do: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return nil, fmt.Errorf("anthropic: http %d: %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}

	return newStream(resp.Body), nil
}

// buildRequestBody serializes the chat request into the Anthropic Messages
// payload shape, including tool definitions and tool_choice when tools are
// present. Tool translation lives in tool_translate.go.
func (a *Adapter) buildRequestBody(req service.ChatRequest) ([]byte, error) {
	model := req.Model
	if strings.TrimSpace(model) == "" {
		model = a.cfg.Model
	}
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	payload := anthropicRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		Messages:    buildMessages(req.Messages),
		System:      req.SystemPrompt,
		Temperature: req.Temperature,
		Stream:      true,
		Tools:       translateTools(req.Tools),
	}
	if len(payload.Tools) > 0 {
		// "auto" lets the model decide whether to call a tool or answer
		// directly. We never force a tool call.
		payload.ToolChoice = map[string]any{"type": "auto"}
	}
	return json.Marshal(payload)
}

// Compile-time interface conformance.
var _ service.ChatProvider = (*Adapter)(nil)
