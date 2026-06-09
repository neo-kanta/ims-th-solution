// Package policy holds pure rules that don't depend on infrastructure.
package policy

import (
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// MaxHistoryMessages is the cap on prior messages resent to a provider per
// turn. A simple message-count cap avoids per-provider token budgeting in
// Slice A while still preventing unbounded growth from blowing context
// windows or running costs (correction 6). Slice D may replace this with a
// token-aware trimmer once we wire the tokenizer per provider.
const MaxHistoryMessages = 40

// TrimHistory returns the most recent MaxHistoryMessages entries, preserving
// chronological order. A leading system message is always kept; older
// user/assistant/tool turns are dropped first.
func TrimHistory(messages []entity.Message) []entity.Message {
	if len(messages) <= MaxHistoryMessages {
		return messages
	}

	var head []entity.Message
	body := messages
	if len(messages) > 0 && messages[0].Role == valueobject.RoleSystem {
		head = messages[:1]
		body = messages[1:]
	}

	keep := MaxHistoryMessages - len(head)
	if keep <= 0 {
		return head
	}
	if len(body) <= keep {
		return messages
	}

	body = body[len(body)-keep:]
	out := make([]entity.Message, 0, len(head)+len(body))
	out = append(out, head...)
	out = append(out, body...)
	return out
}
