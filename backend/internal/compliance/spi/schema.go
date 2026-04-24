package spi

import "encoding/json"

// ParameterSchema wraps a JSON Schema document that validates rule parameters.
type ParameterSchema struct {
	Schema json.RawMessage `json:"schema"`
}

// Validate checks that rawParams conforms to the schema.
// Called at rule-instance-version creation time, never at evaluation time.
// For PoC, this performs basic JSON validity check. Production should use
// a full JSON Schema validator (e.g. santhosh-tekuri/jsonschema).
func (ps ParameterSchema) Validate(rawParams json.RawMessage) error {
	var v interface{}
	return json.Unmarshal(rawParams, &v)
}
