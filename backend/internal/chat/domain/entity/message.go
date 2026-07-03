package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// Message is one persisted turn in a session's transcript.
//
// ProvenanceMap and RawProviderPayload are empty in Slice A. Slice C
// populates RawProviderPayload from the provider's accumulated response;
// Slice D fills ProvenanceMap with the figure-to-tool-result links the
// guardrail layer requires.
type Message struct {
	ID                 uuid.UUID
	SessionID          uuid.UUID
	Role               valueobject.MessageRole
	Content            string
	ProvenanceMap      json.RawMessage
	RawProviderPayload json.RawMessage
	// CorrelationID ties this message to the request/turn that produced it.
	CorrelationID string
	CreatedAt     time.Time
}
