package veritas

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/cel-go/cel"
)

// Validator performs validation on Go objects based on a set of rules.
// It holds a CEL environment configured for the specific types it validates.
type Validator struct {
	engine *Engine
	env    *cel.Env // Validator-specific environment, created fresh.
	rules  map[string]ValidationRuleSet
	logger *slog.Logger
}

// NewValidator creates a new validator.
// It creates a new CEL environment specifically for the types it needs to validate.
func NewValidator(engine *Engine, provider RuleProvider, logger *slog.Logger, typesToRegister ...any) (*Validator, error) {
	rules, err := provider.GetRuleSets()
	if err != nil {
		return nil, fmt.Errorf("failed to get rule sets: %w", err)
	}

	// Start with the base options from the engine.
	opts := make([]cel.EnvOption, len(engine.baseOpts))
	copy(opts, engine.baseOpts)

	// Add type registrations for the validator.
	for _, t := range typesToRegister {
		val := reflect.ValueOf(t)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		// This is the critical part: cel.Types() is now used with cel.NewEnv, not env.Extend.
		opts = append(opts, cel.Types(val.Interface()))
	}
	// Declare 'this' as a variable that can hold the object under validation.
	opts = append(opts, cel.Variable("this", cel.DynType))

	// Create a new environment from scratch with all options.
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	return &Validator{
		engine: engine,
		env:    env,
		rules:  rules,
		logger: logger,
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

	ruleSet, ok := v.rules[typeName]
	if !ok {
		return nil
	}

	var allErrors []error
	vars := map[string]any{"this": obj}

	evaluate := func(rule, fieldName string) error {
		prog, err := v.engine.getProgram(v.env, rule)
		if err != nil {
			v.logger.Error("failed to compile rule", "rule", rule, "type", typeName, "error", err)
			return NewFatalError(fmt.Sprintf("rule compilation error for %s: %s", typeName, err))
		}

		out, _, err := prog.Eval(vars)
		if err != nil {
			v.logger.Error("failed to evaluate rule", "rule", rule, "type", typeName, "field", fieldName, "error", err)
			return NewValidationError(typeName, fieldName, "evaluation error")
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
