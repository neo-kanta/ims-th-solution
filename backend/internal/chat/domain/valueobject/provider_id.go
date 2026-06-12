package valueobject

// ProviderID identifies the active LLM backend. The chat module is provider-
// agnostic; everything downstream of the ChatProvider interface treats this
// as opaque.
type ProviderID string

const (
	ProviderAnthropic        ProviderID = "anthropic"
	ProviderOpenAI           ProviderID = "openai"
	ProviderGemini           ProviderID = "gemini"
	ProviderOpenAICompatible ProviderID = "openai_compatible"
)

// IsValid reports whether p is one of the known provider IDs.
func (p ProviderID) IsValid() bool {
	switch p {
	case ProviderAnthropic, ProviderOpenAI, ProviderGemini, ProviderOpenAICompatible:
		return true
	}
	return false
}

// String returns the enum value as a string.
func (p ProviderID) String() string { return string(p) }
