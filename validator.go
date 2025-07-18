package veritas

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/cel-go/cel"
)

// Validator performs validation on Go objects based on a set of rules.
type Validator struct {
	engine    *Engine
	rules     map[string]ValidationRuleSet
	logger    *slog.Logger
}

// NewValidator creates a new validator with the given engine and rule provider.
func NewValidator(engine *Engine, provider RuleProvider, logger *slog.Logger) (*Validator, error) {
	rules, err := provider.GetRuleSets()
	if err != nil {
		return nil, fmt.Errorf("failed to get rule sets: %w", err)
	}

	return &Validator{
		engine:    engine,
		rules:     rules,
		logger:    logger,
	}, nil
}

// Validate applies the configured rules to the given object.
// It returns a joined error containing all validation failures.
func (v *Validator) Validate(obj any) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	typeName := typ.Name()

	ruleSet, ok := v.rules[typeName]
	if !ok {
		// No rules for this type, so it's valid by default.
		return nil
	}

	var allErrors []error
	vars := map[string]any{"this": obj}

	// 1. Type-level validation
	for _, rule := range ruleSet.TypeRules {
		// For type rules, the fieldName is empty.
		err := v.evaluateRule(rule, obj, vars, typeName, "")
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// 2. Field-level validation
	for fieldName, rules := range ruleSet.FieldRules {
		for _, rule := range rules {
			err := v.evaluateRule(rule, obj, vars, typeName, fieldName)
			if err != nil {
				allErrors = append(allErrors, err)
			}
		}
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	return nil
}

// evaluateRule compiles (with caching) and runs a single CEL rule.
func (v *Validator) evaluateRule(rule string, obj any, vars map[string]any, typeName, fieldName string) error {
	// Dynamically register the type of the object for this evaluation.
	// This makes the validator flexible to any type without pre-registration.
	opts := []cel.EnvOption{
		cel.Types(obj),
		cel.Variable("this", cel.ObjectType(typeName)),
	}

	prog, err := v.engine.getProgram(rule, opts...)
	if err != nil {
		v.logger.Error("failed to compile rule", "rule", rule, "error", err)
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
