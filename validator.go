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

// Validate applies the configured rules to the given object, including nested structs.
func (v *Validator) Validate(obj any) error {
	// Keep track of all errors found during validation.
	var allErrors []error

	// Top-level validation requires a registered TypeAdapter.
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil // Nothing to validate.
		}
		val = val.Elem()
	}
	// Ensure the top-level object is a struct.
	if val.Kind() != reflect.Struct {
		return nil // Or perhaps an error? For now, we only validate structs.
	}
	typ := val.Type()
	typeName := typ.Name()
	if typeName == "" {
		typeName = typ.String()
	}

	if _, ok := v.adapters[typeName]; !ok {
		return NewFatalError(fmt.Sprintf("no TypeAdapter registered for type %s", typeName))
	}

	// Use a helper function to perform the validation recursively.
	v.validateRecursive(obj, &allErrors)

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}

// validateRecursive is the internal helper that performs the actual validation.
func (v *Validator) validateRecursive(obj any, allErrors *[]error) {
	// Dereference pointer to get the actual value.
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return // Skip validation for nil pointers.
		}
		val = val.Elem()
	}

	// We only validate structs.
	if val.Kind() != reflect.Struct {
		return
	}
	typ := val.Type()
	typeName := typ.Name()
	if typeName == "" {
		typeName = typ.String() // Handle anonymous structs
	}

	// Get the adapter for the current type. If none, we can't validate its fields with CEL.
	adapter, hasAdapter := v.adapters[typeName]
	if !hasAdapter {
		// Even without an adapter, we must still recurse into its fields.
		v.logger.Debug("no TypeAdapter, but continuing to recurse", "type", typeName)
	}

	// Get the rule set. If none, we might still need to recurse.
	ruleSet, hasRules := v.rules[typeName]

	// If there are rules and an adapter, perform CEL validation.
	if hasRules && hasAdapter {
		// Convert the object to a map for CEL evaluation.
		objMap, err := adapter(obj)
		if err != nil {
			v.logger.Error("failed to adapt object", "type", typeName, "error", err)
			*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("TypeAdapter error for %s: %v", typeName, err)))
			return // Stop validation for this object if adapter fails.
		}
		vars := map[string]any{"this": objMap}

		// Evaluation helper function.
		evaluate := func(rule, fieldName string) {
			prog, err := v.engine.getProgram(v.env, rule)
			if err != nil {
				v.logger.Error("failed to compile rule", "rule", rule, "type", typeName, "error", err)
				*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("rule compilation error for %s: %s", typeName, err)))
				return
			}

			out, _, err := prog.Eval(vars)
			if err != nil {
				v.logger.Error("failed to evaluate rule", "rule", rule, "type", typeName, "field", fieldName, "error", err)
				*allErrors = append(*allErrors, NewValidationError(typeName, fieldName, fmt.Sprintf("evaluation error: %s", err)))
				return
			}

			if valid, ok := out.Value().(bool); !ok || !valid {
				*allErrors = append(*allErrors, NewValidationError(typeName, fieldName, rule))
			}
		}

		// Apply type and field rules.
		for _, rule := range ruleSet.TypeRules {
			evaluate(rule, "")
		}
		for fieldName, rules := range ruleSet.FieldRules {
			for _, rule := range rules {
				evaluate(rule, fieldName)
			}
		}
	}

	// --- Recursive Validation Step ---
	// Iterate over the fields of the struct to find nested structs to validate.
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)

		// We need to check if the field is a struct or a pointer to a struct.
		kind := fieldVal.Kind()
		if kind == reflect.Struct || (kind == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct) {
			// Ensure we can get an interface to the field to pass to the recursive call.
			if fieldVal.CanInterface() {
				v.validateRecursive(fieldVal.Interface(), allErrors)
			}
		}
	}
}
