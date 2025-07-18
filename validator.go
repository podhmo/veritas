package veritas

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

// TypeAdapter converts a Go object into a map representation that CEL can understand.
// This is necessary to bypass issues with direct Go struct registration in cel-go.
type TypeAdapter func(obj any) (map[string]any, error)

// Validator performs validation on Go objects based on a set of rules.
type Validator struct {
	engine  *Engine
	env     *cel.Env // Validator-specific environment.
	rules   map[string]ValidationRuleSet
	adapters map[string]TypeAdapter
	logger  *slog.Logger
}

// NewValidator creates a new validator.
// It requires a map of type names to TypeAdapter functions to handle object-to-map conversion.
func NewValidator(engine *Engine, provider RuleProvider, logger *slog.Logger, adapters map[string]TypeAdapter) (*Validator, error) {
	rules, err := provider.GetRuleSets()
	if err != nil {
		return nil, fmt.Errorf("failed to get rule sets: %w", err)
	}

	// Start with the base options from the engine.
	opts := make([]cel.EnvOption, len(engine.baseOpts))
	copy(opts, engine.baseOpts)

	// Declare 'this' as a variable that can hold the object under validation.
	// We use types.NewMapType to ensure CEL treats 'this' as a map.
	opts = append(opts, cel.Variable("this", types.NewMapType(types.StringType, types.DynType)))

	// Create a new environment from scratch with all options.
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	return &Validator{
		engine:  engine,
		env:     env,
		rules:   rules,
		adapters: adapters,
		logger:  logger,
	}, nil
}

// Validate applies the configured rules to the given object.
func (v *Validator) Validate(obj any) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	typeName := typ.Name()
	if typeName == "" {
		typeName = typ.String()
	}

	// Find the adapter for the type.
	adapter, ok := v.adapters[typeName]
	if !ok {
		// If no adapter is found, we cannot validate. This is a configuration error.
		return NewFatalError(fmt.Sprintf("no TypeAdapter registered for type %s", typeName))
	}

	// Find the rule set for the type.
	ruleSet, ok := v.rules[typeName]
	if !ok {
		// No rules for this type, so it's valid by default.
		return nil
	}

	// Convert the object to a map using the adapter.
	objMap, err := adapter(obj)
	if err != nil {
		v.logger.Error("failed to adapt object", "type", typeName, "error", err)
		return NewFatalError(fmt.Sprintf("TypeAdapter error for %s: %s", typeName, err))
	}

	var allErrors []error
	vars := map[string]any{"this": objMap}

	evaluate := func(rule, fieldName string) error {
		prog, err := v.engine.getProgram(v.env, rule)
		if err != nil {
			v.logger.Error("failed to compile rule", "rule", rule, "type", typeName, "error", err)
			return NewFatalError(fmt.Sprintf("rule compilation error for %s: %s", typeName, err))
		}

		out, _, err := prog.Eval(vars)
		if err != nil {
			v.logger.Error("failed to evaluate rule", "rule", rule, "type", typeName, "field", fieldName, "error", err)
			return NewValidationError(typeName, fieldName, fmt.Sprintf("evaluation error: %s", err))
		}

		if valid, ok := out.Value().(bool); !ok || !valid {
			return NewValidationError(typeName, fieldName, rule)
		}
		return nil
	}

	for _, rule := range ruleSet.TypeRules {
		if err := evaluate(rule, ""); err != nil {
			allErrors = append(allErrors, err)
		}
	}

	for fieldName, rules := range ruleSet.FieldRules {
		for _, rule := range rules {
			if err := evaluate(rule, fieldName); err != nil {
				allErrors = append(allErrors, err)
			}
		}
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}
