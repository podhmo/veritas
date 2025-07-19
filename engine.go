// Package veritas provides a dynamic, type-safe, and extensible validation library for Go.
package veritas

import (
	"log/slog"

	"github.com/google/cel-go/cel"
	lru "github.com/hashicorp/golang-lru/v2"
)

// Engine is the core component that manages base configurations and program caching.
// It does not hold a CEL environment itself, but provides the base options to create them.
type Engine struct {
	baseOpts     []cel.EnvOption
	programCache *lru.Cache[string, cel.Program]
	logger       *slog.Logger
}

// NewEngine creates a new validation engine.
func NewEngine(logger *slog.Logger, funcs ...cel.EnvOption) (*Engine, error) {
	// Add support for common CEL features.
	opts := DefaultEnvOptions()
	opts = append(opts, funcs...)
	opts = append(opts, cel.HomogeneousAggregateLiterals())

	cache, err := lru.New[string, cel.Program](256)
	if err != nil {
		return nil, err
	}

	return &Engine{
		baseOpts:     opts,
		programCache: cache,
		logger:       logger,
	}, nil
}

// getProgram compiles a CEL expression against a given environment and returns a usable program.
// It uses an LRU cache to avoid re-compiling frequently used expressions.
// The cache key is just the rule string, implying that rules are environment-agnostic enough
// for this library's use case, which might be a simplifying assumption.
func (e *Engine) getProgram(env *cel.Env, rule string) (cel.Program, error) {
	if prog, ok := e.programCache.Get(rule); ok {
		e.logger.Debug("cache hit", "rule", rule)
		return prog, nil
	}

	e.logger.Debug("cache miss", "rule", rule)

	ast, issues := env.Compile(rule)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	prog, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	e.programCache.Add(rule, prog)
	return prog, nil
}
