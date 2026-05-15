package main

import (
	"fmt"
	"reflect"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// envLookup mirrors os.LookupEnv with a smaller surface so tests can inject
// a deterministic environment.
type envLookup func(key string) (string, bool)

// binding describes one cross-module contract the contract-check inspects.
type binding struct {
	// Name is a human-readable label, e.g. "investment.ComplianceChecker".
	Name string
	// Resolved is the binding produced by the production wiring. Pass the
	// concrete value typed as the contract's interface, NOT a typed nil.
	Resolved any
	// AllowFakeEnv overrides the default IMS_ALLOW_FAKE_<FakeName> env name.
	// When empty, the default form is used.
	AllowFakeEnv string
}

// verificationOutcome reports the verdict for one binding.
type verificationOutcome struct {
	Name    string
	Status  string // "ok", "fake-allowed", "fake-rejected", "nil"
	Message string
}

// verifyBindings runs each binding through the validator and returns the
// outcomes plus a non-nil error if any binding failed. The function is
// pure: it does not log, exit, or read os.Getenv directly.
func verifyBindings(bindings []binding, lookup envLookup) ([]verificationOutcome, error) {
	if lookup == nil {
		return nil, fmt.Errorf("envLookup must not be nil")
	}

	outcomes := make([]verificationOutcome, 0, len(bindings))
	var failed []string

	for _, b := range bindings {
		outcome := verifyBinding(b, lookup)
		outcomes = append(outcomes, outcome)
		switch outcome.Status {
		case "nil", "fake-rejected":
			failed = append(failed, fmt.Sprintf("%s: %s", outcome.Name, outcome.Message))
		}
	}

	if len(failed) > 0 {
		return outcomes, fmt.Errorf("contract bindings rejected: %v", failed)
	}
	return outcomes, nil
}

// verifyBinding inspects a single binding for nil and fake-marker status.
func verifyBinding(b binding, lookup envLookup) verificationOutcome {
	out := verificationOutcome{Name: b.Name}

	if isNilBinding(b.Resolved) {
		out.Status = "nil"
		out.Message = "binding is nil — no real implementation wired"
		return out
	}

	if marker, ok := b.Resolved.(contract.FakeMarker); ok {
		fakeName := marker.FakeName()
		envName := b.AllowFakeEnv
		if envName == "" {
			envName = "IMS_ALLOW_FAKE_" + fakeName
		}
		if v, ok := lookup(envName); ok && v == "true" {
			out.Status = "fake-allowed"
			out.Message = fmt.Sprintf("bound to fake %q (allowed by %s=true)", fakeName, envName)
			return out
		}
		out.Status = "fake-rejected"
		out.Message = fmt.Sprintf("bound to fake %q; set %s=true to allow", fakeName, envName)
		return out
	}

	out.Status = "ok"
	return out
}

// isNilBinding handles the standard Go interface-vs-pointer nil trap so
// callers can pass typed nil pointers through binding.Resolved without
// false positives.
func isNilBinding(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return rv.IsNil()
	}
	return false
}
