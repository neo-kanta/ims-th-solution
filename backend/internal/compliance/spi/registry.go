package spi

import (
	"fmt"
	"sync"
)

var globalRegistry = &RuleRegistry{
	rules: make(map[string]RuleEvaluator),
}

// Register adds a rule type to the global registry.
// Called from init() in each rule package. Panics on duplicate TypeID.
func Register(rule RuleEvaluator) {
	globalRegistry.MustRegister(rule)
}

// GlobalRegistry returns the singleton registry.
func GlobalRegistry() *RuleRegistry {
	return globalRegistry
}

// RuleRegistry holds all registered rule types. Thread-safe for reads after startup.
type RuleRegistry struct {
	mu    sync.RWMutex
	rules map[string]RuleEvaluator
}

// NewRegistry creates an empty registry (for testing).
func NewRegistry() *RuleRegistry {
	return &RuleRegistry{rules: make(map[string]RuleEvaluator)}
}

// MustRegister adds a rule, panicking on duplicate TypeID.
func (r *RuleRegistry) MustRegister(rule RuleEvaluator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	meta := rule.Metadata()
	if _, exists := r.rules[meta.TypeID]; exists {
		panic(fmt.Sprintf("compliance: duplicate rule type registration: %s", meta.TypeID))
	}
	r.rules[meta.TypeID] = rule
}

// Get returns a rule evaluator by TypeID.
func (r *RuleRegistry) Get(typeID string) (RuleEvaluator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.rules[typeID]
	return rule, ok
}

// All returns all registered evaluators.
func (r *RuleRegistry) All() []RuleEvaluator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]RuleEvaluator, 0, len(r.rules))
	for _, rule := range r.rules {
		result = append(result, rule)
	}
	return result
}

// TypeIDs returns all registered type IDs.
func (r *RuleRegistry) TypeIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.rules))
	for id := range r.rules {
		ids = append(ids, id)
	}
	return ids
}
