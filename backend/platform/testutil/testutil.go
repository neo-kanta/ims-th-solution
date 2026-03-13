package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// NewJSONRequest creates an *http.Request with a JSON body for testing.
func NewJSONRequest(method, target string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			panic("testutil: failed to encode request body: " + err.Error())
		}
	}
	req := httptest.NewRequest(method, target, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// DecodeJSONResponse decodes the response recorder body into the target struct.
func DecodeJSONResponse(w *httptest.ResponseRecorder, target interface{}) error {
	return json.NewDecoder(w.Body).Decode(target)
}
