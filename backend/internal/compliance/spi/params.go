package spi

import "encoding/json"

// ParameterSet wraps validated, frozen parameters from a rule instance version.
// Rules call Decode to deserialize into their typed Params struct.
type ParameterSet struct {
	raw json.RawMessage
}

// NewParameterSet creates a ParameterSet from raw JSON.
func NewParameterSet(raw json.RawMessage) ParameterSet {
	return ParameterSet{raw: raw}
}

// Decode deserializes parameters into the target struct.
func (ps ParameterSet) Decode(target interface{}) error {
	return json.Unmarshal(ps.raw, target)
}

// Raw returns the underlying JSON for snapshotting into audit records.
func (ps ParameterSet) Raw() json.RawMessage {
	return ps.raw
}
