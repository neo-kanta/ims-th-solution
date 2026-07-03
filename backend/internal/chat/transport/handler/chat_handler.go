// Package handler holds chat HTTP transport handlers. Handlers parse
// requests, hand them to application commands, and forward events to the
// client. They contain no business logic.
package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
	"github.com/neo-kanta/ims-th-solution/backend/platform/validation"
)

// ChatHandler exposes the streaming chat endpoint.
type ChatHandler struct {
	cmd *command.SendMessage
}

// NewChatHandler wires the handler.
func NewChatHandler(cmd *command.SendMessage) *ChatHandler {
	return &ChatHandler{cmd: cmd}
}

// SendMessage streams an assistant reply for one user turn over SSE.
// @Summary      Send a chat message and stream the response
// @Description  Streams text/event-stream. Each frame's `data:` payload is a ChatStreamEvent (see response model). Frames have event names: session_started, text, done, error. Use fetch() + ReadableStream on the client — EventSource cannot carry an Authorization header.
// @Tags         Chat
// @Accept       json
// @Produce      text/event-stream
// @Param        request body request.SendMessageRequest true "Chat message"
// @Success      200 {object} response.ChatStreamEvent "One example frame payload; the response is a stream of these"
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /chat [post]
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil || claims.Subject == "" {
		httputil.Unauthorized(w, "missing user context")
		return
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.Unauthorized(w, "invalid user identity")
		return
	}

	var body request.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	if fieldErrors := validation.ValidateStruct(&body); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, httputil.ErrorResponse{
			Error:   "validation failed",
			Details: fieldErrors,
		})
		return
	}

	var sessionID uuid.UUID
	if body.SessionID != "" {
		sid, perr := uuid.Parse(body.SessionID)
		if perr != nil {
			httputil.BadRequest(w, "invalid session_id")
			return
		}
		sessionID = sid
	}

	in := command.SendMessageInput{
		SessionID: sessionID,
		UserID:    userID,
		Content:   body.Content,
		IPAddress: middleware.GetClientIP(r),
		UserAgent: r.UserAgent(),
		Model:     body.Model,
		Provider:  body.Provider,
		// Forward the caller's bearer token so MCP tools can call the IMS
		// REST API with the user's own identity and permissions. It is never
		// shown to the model nor written to audit.
		AuthToken: bearerToken(r),
		// Correlation id (chi RequestID) ties this turn end-to-end across
		// audit, message/tool rows and logs.
		CorrelationID: chimw.GetReqID(r.Context()),
	}

	events, err := h.cmd.Run(r.Context(), in)
	if err != nil {
		// Validation / setup errors must surface BEFORE we switch to SSE,
		// otherwise the client can't tell a failed turn from an empty stream.
		httputil.BadRequest(w, safeErrorMessage(err))
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		httputil.InternalError(w, "streaming not supported by this server")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	// Hint to reverse proxies (nginx, GCP) not to buffer the stream.
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	for ev := range events {
		dto := response.ChatStreamEvent{
			Kind:       string(ev.Kind),
			SessionID:  ev.SessionID,
			MessageID:  ev.MessageID,
			Text:       ev.Text,
			ToolName:   ev.ToolName,
			ToolStatus: ev.ToolStatus,
			StopReason: string(ev.StopReason),
			Error:      ev.Error,
		}
		for _, s := range ev.Sources {
			dto.Sources = append(dto.Sources, response.ChatSourceRef{
				ToolName: s.ToolName, Server: s.Server, State: s.State, AsOf: s.AsOf,
			})
		}
		for _, b := range ev.Bindings {
			dto.Bindings = append(dto.Bindings, response.ChatFigureBinding{
				Figure: b.Figure, Source: b.Source, ToolName: b.ToolName,
			})
		}
		if len(ev.Unverified) > 0 {
			dto.Unverified = append(dto.Unverified, ev.Unverified...)
		}
		payload, err := json.Marshal(dto)
		if err != nil {
			slog.Error("chat: marshal sse payload failed", "error", err)
			continue
		}
		// SSE frame: event line + data line + blank line terminator.
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", dto.Kind, payload); err != nil {
			// Client disconnected — drain the channel so the goroutine exits.
			drain(events)
			return
		}
		flusher.Flush()
	}
}

// safeErrorMessage returns a stable, non-leaky message. Detailed errors stay
// in slog; the client sees a generic phrase.
func safeErrorMessage(err error) string {
	switch err.Error() {
	case "user id required", "message content required", "session not found",
		"session does not belong to user", "invalid session_id":
		return err.Error()
	}
	if len(err.Error()) > 0 && len(err.Error()) < 120 {
		return err.Error()
	}
	return "unable to start chat turn"
}

func drain(ch <-chan command.StreamEvent) {
	for range ch {
	}
}

// bearerToken extracts the raw JWT from the Authorization header. The auth
// middleware has already validated it; we forward the same token to MCP tools
// so their REST calls run as the user. Returns "" if absent/malformed.
func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return h[len(prefix):]
	}
	return ""
}
