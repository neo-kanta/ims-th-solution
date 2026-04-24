package spi

import "context"

// RuleEvaluator is the Service Provider Interface that every rule type implements.
//
// A rule type is a Go package that:
//  1. Implements this interface
//  2. Calls spi.Register() in init()
//  3. Contains zero imports from other rule packages, engine, or infrastructure
//
// Evaluate MUST be a pure function: deterministic given the same inputs,
// no side effects, no database access, no HTTP calls, no clock reads.
type RuleEvaluator interface {
	// Metadata returns the stable identity and capabilities of this rule type.
	Metadata() RuleMetadata

	// ParameterSchema returns the JSON Schema that validates parameters.
	ParameterSchema() ParameterSchema

	// DataDependencies declares what data this rule needs to evaluate.
	DataDependencies() DataDependencies

	// Evaluate runs the rule against the given input and pre-fetched data.
	// Returns EvalResult with a raw verdict. The engine applies severity cap after.
	// Errors indicate a bug or data corruption, not a compliance failure.
	// A rule that cannot determine a verdict should return VerdictBlock, not an error.
	Evaluate(ctx context.Context, input CheckInput, data DataBundle, params ParameterSet) (EvalResult, error)

	// Explain produces a human-readable explanation of a result.
	Explain(input CheckInput, result EvalResult, params ParameterSet) Explanation
}
