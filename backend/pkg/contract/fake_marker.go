package contract

// FakeMarker is implemented by every test or development fake of a
// cross-module contract. The contract-check binary inspects each resolved
// binding for this interface; bindings that report a fake name are rejected
// unless explicitly allow-listed via an environment variable (typically
// IMS_ALLOW_FAKE_<NAME>).
//
// Real implementations MUST NOT implement this interface — its presence is
// the signal that a fake is in production wiring.
type FakeMarker interface {
	// FakeName returns a stable UPPER_SNAKE identifier the contract-check
	// uses to construct its env-gate variable name.
	FakeName() string
}
