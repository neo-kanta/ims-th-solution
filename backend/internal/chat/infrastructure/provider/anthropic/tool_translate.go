package anthropic

import (
	"encoding/json"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// translateTools converts canonical ToolSpecs into Anthropic tool definitions.
// This is the ONLY place Anthropic's tool schema shape is known; the same
// canonical ToolSpec is translated differently by each provider adapter, which
// is what lets one MCP tool work across all providers.
func translateTools(tools []service.ToolSpec) []anthropicTool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]anthropicTool, 0, len(tools))
	for _, t := range tools {
		schema := t.InputSchema
		if len(schema) == 0 {
			schema = json.RawMessage(`{"type":"object"}`)
		}
		out = append(out, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: schema,
		})
	}
	return out
}

// buildMessages converts the canonical conversation into Anthropic messages.
//
// Mapping:
//   - user/assistant plain text → {role, content:"..."}
//   - assistant with tool calls → {role:"assistant", content:[text?, tool_use...]}
//   - tool results             → {role:"user", content:[tool_result...]}
//
// Anthropic requires tool_result blocks to be sent in a USER message, which is
// why a canonical RoleTool message is emitted as role "user" here.
func buildMessages(msgs []service.ConversationMessage) []anthropicMessage {
	out := make([]anthropicMessage, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case valueobject.RoleAssistant:
			if len(m.ToolCalls) == 0 {
				out = append(out, anthropicMessage{Role: "assistant", Content: m.Content})
				continue
			}
			blocks := make([]any, 0, len(m.ToolCalls)+1)
			if m.Content != "" {
				blocks = append(blocks, textBlock{Type: "text", Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				input := tc.Input
				if len(input) == 0 {
					input = json.RawMessage(`{}`)
				}
				blocks = append(blocks, toolUseBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: input,
				})
			}
			out = append(out, anthropicMessage{Role: "assistant", Content: blocks})

		case valueobject.RoleTool:
			blocks := make([]any, 0, len(m.ToolResults))
			for _, tr := range m.ToolResults {
				blocks = append(blocks, toolResultBlock{
					Type:      "tool_result",
					ToolUseID: tr.ToolCallID,
					Content:   tr.Content,
					IsError:   tr.IsError,
				})
			}
			out = append(out, anthropicMessage{Role: "user", Content: blocks})

		case valueobject.RoleUser:
			out = append(out, anthropicMessage{Role: "user", Content: m.Content})

		default:
			// system messages are handled via the top-level System field.
		}
	}
	return out
}
