// Package veritas provides a dynamic, type-safe, and extensible validation library for Go.
package veritas

import (
	"log/slog"

	"github.com/google/cel-go/cel"
	"github.com/hashicorp/golang-lru/v2"
)

// Engine is the core component that manages the CEL environment and program caching.
// It is responsible for compiling and caching CEL expressions for efficient re-use.
type Engine struct {
	env    *cel.Env
	cache  *lru.Cache[string, cel.Program]
	logger *slog.Logger
}

// NewEngine creates a new validation engine.
// It initializes the CEL environment and the LRU cache for compiled programs.
func NewEngine(logger *slog.Logger, funcs ...cel.EnvOption) (*Engine, error) {
	env, err := cel.NewEnv(funcs...)
	if err != nil {
		return nil, err
	}

	cache, err := lru.New[string, cel.Program](128) // Default cache size
	if err != nil {
		return nil, err
	}

	return &Engine{
		env:    env,
		cache:  cache,
		logger: logger,
	}, nil
}

// getProgram compiles a CEL expression and returns a usable program.
// It uses an LRU cache to avoid re-compiling frequently used expressions.
func (e *Engine) getProgram(rule string) (cel.Program, error) {
	if prog, ok := e.cache.Get(rule); ok {
		e.logger.Debug("cache hit", "rule", rule)
		return prog, nil
	}

	e.logger.Debug("cache miss", "rule", rule)
	ast, issues := e.env.Compile(rule)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	prog, err := e.env.Program(ast)
	if err != nil {
		return nil, err
	}

	e.cache.Add(rule, prog)
	return prog, nil
}