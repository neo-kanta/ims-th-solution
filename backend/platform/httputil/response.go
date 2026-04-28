package httputil

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse is the standard envelope for successful API responses.
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse is the standard envelope for error API responses.
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// JSON writes a JSON response with the given status code and data.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

// OK writes a 200 JSON response with the given data.
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, SuccessResponse{Data: data})
}

// Created writes a 201 JSON response with the given data.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, SuccessResponse{Data: data})
}

// NoContent writes a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 error response.
func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, ErrorResponse{Error: message})
}

// Unauthorized writes a 401 error response.
func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, ErrorResponse{Error: message})
}

// Forbidden writes a 403 error response.
func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, ErrorResponse{Error: message})
}

// NotFound writes a 404 error response.
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, ErrorResponse{Error: message})
}

// Conflict writes a 409 error response. Used when a request conflicts with
// the current state of the target resource (e.g. duplicate override, stale state).
func Conflict(w http.ResponseWriter, message string) {
	JSON(w, http.StatusConflict, ErrorResponse{Error: message})
}

// UnprocessableEntity writes a 422 error response. Used when the request is
// syntactically valid but references an entity/state the server cannot act on
// (e.g. an unregistered rule type ID).
func UnprocessableEntity(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: message})
}

// InternalError writes a 500 error response.
func InternalError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, ErrorResponse{Error: message})
}
