package anthropic

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

const sseLineBufferMax = 1 << 20

// toolAccum accumulates a streaming tool_use content block. Anthropic streams
// the tool input as a sequence of input_json_delta partial-JSON fragments
// between content_block_start and content_block_stop.
type toolAccum struct {
	id      string
	name    string
	jsonBuf strings.Builder
}

// stream is the Anthropic-side ChatStreamReader. It surfaces text deltas as
// they arrive and accumulates any tool_use blocks; completed tool calls are
// exposed via Result() once the stream ends. Next is single-threaded.
type stream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
	result  service.ChatResult
	err     error
	done    bool

	tools map[int]*toolAccum // content-block index → accumulating tool call
}

func newStream(body io.ReadCloser) *stream {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), sseLineBufferMax)
	return &stream{
		body:    body,
		scanner: scanner,
		tools:   map[int]*toolAccum{},
		result: service.ChatResult{
			StopReason: valueobject.StopReasonUnknown,
		},
	}
}

// Next returns the next text delta. False means the stream ended — check Err
// and Result afterwards. Tool-use blocks are accumulated silently and do not
// produce text deltas.
func (s *stream) Next(ctx context.Context) (service.TextDelta, bool) {
	if s.done {
		return service.TextDelta{}, false
	}

	for {
		if err := ctx.Err(); err != nil {
			s.err = err
			s.finish()
			return service.TextDelta{}, false
		}

		eventName, data, ok := s.readEvent()
		if !ok {
			if scanErr := s.scanner.Err(); scanErr != nil {
				s.err = scanErr
			}
			s.finish()
			return service.TextDelta{}, false
		}

		switch eventName {
		case "content_block_start":
			var ev contentBlockStart
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			if ev.ContentBlock.Type == "tool_use" {
				s.tools[ev.Index] = &toolAccum{id: ev.ContentBlock.ID, name: ev.ContentBlock.Name}
			}

		case "content_block_delta":
			var ev contentBlockDelta
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			switch ev.Delta.Type {
			case "text_delta":
				if ev.Delta.Text != "" {
					return service.TextDelta{Text: ev.Delta.Text}, true
				}
			case "input_json_delta":
				if t := s.tools[ev.Index]; t != nil {
					t.jsonBuf.WriteString(ev.Delta.PartialJSON)
				}
			}

		case "message_delta":
			var ev messageDelta
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			s.result.StopReason = mapStopReason(ev.Delta.StopReason)

		case "message_stop":
			s.finish()
			return service.TextDelta{}, false

		case "error":
			var ev errorPayload
			_ = json.Unmarshal([]byte(data), &ev)
			detail := ev.Error.Message
			if detail == "" {
				detail = "unknown error"
			}
			s.err = fmt.Errorf("anthropic stream error (%s): %s", ev.Error.Type, detail)
			s.result.StopReason = valueobject.StopReasonError
			s.finish()
			return service.TextDelta{}, false

		default:
			// ping, message_start, content_block_stop, and unknown future
			// event types — ignored, per docs guidance to handle unknown
			// events gracefully.
		}
	}
}

// finish finalizes accumulated tool calls into the result, once, in stable
// content-block index order.
func (s *stream) finish() {
	if s.done {
		return
	}
	s.done = true

	if len(s.tools) == 0 {
		return
	}
	indices := make([]int, 0, len(s.tools))
	for idx := range s.tools {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	for _, idx := range indices {
		t := s.tools[idx]
		input := strings.TrimSpace(t.jsonBuf.String())
		if input == "" {
			input = "{}"
		}
		s.result.ToolCalls = append(s.result.ToolCalls, service.ToolCall{
			ID:    t.id,
			Name:  t.name,
			Input: json.RawMessage(input),
		})
	}
}

func (s *stream) Result() service.ChatResult { return s.result }
func (s *stream) Err() error                 { return s.err }

func (s *stream) Close() error {
	if s.body == nil {
		return nil
	}
	err := s.body.Close()
	s.body = nil
	if err != nil && !errors.Is(err, io.ErrClosedPipe) {
		return err
	}
	return nil
}

// readEvent assembles one SSE event (event name + joined data) from the
// scanner. Returns false on EOF / scanner error.
func (s *stream) readEvent() (string, string, bool) {
	var (
		eventName string
		dataBuf   strings.Builder
		anyField  bool
	)
	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			if anyField {
				return eventName, dataBuf.String(), true
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			anyField = true
		case strings.HasPrefix(line, "data:"):
			payload := strings.TrimPrefix(line, "data:")
			if strings.HasPrefix(payload, " ") {
				payload = payload[1:]
			}
			if dataBuf.Len() > 0 {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(payload)
			anyField = true
		}
	}
	return "", "", false
}

func mapStopReason(s string) valueobject.StopReason {
	switch s {
	case "end_turn":
		return valueobject.StopReasonEndTurn
	case "tool_use":
		return valueobject.StopReasonToolUse
	case "max_tokens":
		return valueobject.StopReasonMaxTokens
	case "stop_sequence":
		return valueobject.StopReasonStopSequence
	case "":
		return valueobject.StopReasonUnknown
	default:
		return valueobject.StopReasonUnknown
	}
}
