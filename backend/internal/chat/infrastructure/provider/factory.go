// Package provider builds the active ChatProvider based on config. Slice A
// only knows about Anthropic; Slice B extends Build with the remaining
// adapters behind the same switch.
package provider

import (
	"errors"
	"fmt"
	"strings"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/infrastructure/provider/anthropic"
)

// Config is the chat-providers section of platform config, mapped into a
// provider-neutral shape so platform/config doesn't need to know about each
// adapter's specifics.
type Config struct {
	ActiveProvider valueobject.ProviderID

	AnthropicAPIKey string
	AnthropicModel  string

	// Slice B will add OpenAIAPIKey, GeminiAPIKey, OpenAICompatibleBaseURL,
	// OpenAICompatibleAPIKey, OpenAICompatibleModel here.
}

// Build returns the configured provider. An unknown ActiveProvider is an
// error — fail fast at startup rather than per-request.
func Build(cfg Config) (service.ChatProvider, error) {
	if !cfg.ActiveProvider.IsValid() {
		return nil, fmt.Errorf("chat: unknown LLM_PROVIDER %q", cfg.ActiveProvider)
	}

	switch cfg.ActiveProvider {
	case valueobject.ProviderAnthropic:
		if strings.TrimSpace(cfg.AnthropicAPIKey) == "" {
			return nil, errors.New("chat: ANTHROPIC_API_KEY is required when LLM_PROVIDER=anthropic")
		}
		return anthropic.New(anthropic.Config{
			APIKey: cfg.AnthropicAPIKey,
			Model:  cfg.AnthropicModel,
		})

	case valueobject.ProviderOpenAI, valueobject.ProviderGemini, valueobject.ProviderOpenAICompatible:
		return nil, fmt.Errorf("chat: provider %q not yet implemented (Slice B)", cfg.ActiveProvider)
	}

	return nil, fmt.Errorf("chat: unsupported LLM_PROVIDER %q", cfg.ActiveProvider)
}
