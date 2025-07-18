package veritas

import "github.com/google/cel-go/cel"

// DefaultFunctions returns a list of common, useful custom functions for CEL.
func DefaultFunctions() []cel.EnvOption {
	// TODO: Implement custom functions like:
	// - strings.ToUpper(string) -> string
	// - strings.ToLower(string) -> string
	// - matches(string, regexp) -> bool
	return []cel.EnvOption{}
}