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
	isPtr := val.Kind() == reflect.Ptr
	if isPtr {
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

	// Create a new environment with the type of the object registered.
	// This allows CEL to understand the object's structure.
	env, err := v.engine.env.Extend(cel.Types(reflect.New(typ).Interface()), cel.Declarations(cel.Variable("this", cel.ObjectType(typ.Name()))))
	if err != nil {
		v.logger.Error("failed to extend env with object type", "type", typeName, "error", err)
		return NewFatalError(fmt.Sprintf("failed to register type %s: %v", typeName, err))
	}

	// 1. Type-level validation
	typeVars := map[string]any{
		"this": obj,
	}
	for _, rule := range ruleSet.TypeRules {
		err := v.evaluateRuleWithEnv(env, rule, typeVars, typeName, "")
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// 2. Field-level validation
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		rules, ok := ruleSet.FieldRules[field.Name]
		if !ok {
			continue
		}

		fieldVars := map[string]any{
			"this": fieldVal.Interface(),
		}
		for _, rule := range rules {
			// Note: We use the original engine's env for field rules,
			// as they operate on primitive types ('this').
			err := v.evaluateRule(rule, fieldVars, typeName, field.Name)
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


// evaluateRule compiles and runs a single CEL rule using the validator's main engine.
func (v *Validator) evaluateRule(rule string, vars map[string]any, typeName, fieldName string) error {
	prog, err := v.engine.getProgram(rule, cel.Variable("this", cel.DynType))
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

// evaluateRuleWithEnv compiles and runs a single CEL rule using a custom environment.
// This is used for type-level rules where the object's type needs to be registered at runtime.
// This version does not use caching.
func (v *Validator) evaluateRuleWithEnv(env *cel.Env, rule string, vars map[string]any, typeName, fieldName string) error {
	ast, issues := env.Compile(rule)
	if issues != nil && issues.Err() != nil {
		v.logger.Error("failed to compile rule with custom env", "rule", rule, "error", issues.Err())
		return NewFatalError(fmt.Sprintf("rule compilation error for %s: %s", typeName, issues.Err()))
	}

	prog, err := env.Program(ast)
	if err != nil {
		v.logger.Error("failed to create program with custom env", "rule", rule, "error", err)
		return NewFatalError(fmt.Sprintf("program creation error for %s: %s", typeName, err))
	}

	out, _, err := prog.Eval(vars)
	if err != nil {
		v.logger.Error("failed to evaluate rule with custom env", "rule", rule, "type", typeName, "field", fieldName, "error", err)
		return NewValidationError(typeName, fieldName, "evaluation error")
	}

	if valid, ok := out.Value().(bool); !ok || !valid {
		return NewValidationError(typeName, fieldName, rule)
	}

	return nil
}
