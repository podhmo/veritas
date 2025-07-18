package veritas

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

// TypeAdapter converts a Go object into a map representation that CEL can understand.
// This is necessary to bypass issues with direct Go struct registration in cel-go.
type TypeAdapter func(obj any) (map[string]any, error)

// Validator performs validation on Go objects based on a set of rules.
type Validator struct {
	engine   *Engine
	objectEnv *cel.Env // For object-level rules (e.g., self.field > 10)
	fieldEnv  *cel.Env // For field-level rules (e.g., self.size() > 0)
	rules    map[string]ValidationRuleSet
	adapters map[string]TypeAdapter
	logger   *slog.Logger
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

	// Environment for object-level validation where 'self' is a map.
	objOpts := make([]cel.EnvOption, len(engine.baseOpts))
	copy(objOpts, engine.baseOpts)
	objOpts = append(objOpts, cel.Variable("self", types.NewMapType(types.StringType, types.DynType)))
	objectEnv, err := cel.NewEnv(objOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create object CEL environment: %w", err)
	}

	// Environment for field-level validation where 'self' is a dynamic type.
	fieldOpts := make([]cel.EnvOption, len(engine.baseOpts))
	copy(fieldOpts, engine.baseOpts)
	fieldOpts = append(fieldOpts, cel.Variable("self", types.DynType))
	fieldEnv, err := cel.NewEnv(fieldOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create field CEL environment: %w", err)
	}

	return &Validator{
		engine:    engine,
		objectEnv: objectEnv,
		fieldEnv:  fieldEnv,
		rules:     rules,
		adapters:  adapters,
		logger:    logger,
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
	var typeName string
	if pkgPath := typ.PkgPath(); pkgPath != "" {
		// To match the parser's output, we might need to adjust how the package name is derived.
		// The parser uses `pkg.Name`, which is usually the last part of the import path.
		// For now, let's assume `typ.PkgPath()` gives us something we can work with.
		// A more robust solution might require passing the package name map from the parser.
		// Let's use a simple approach for now.
		typeName = fmt.Sprintf("%s.%s", typ.Name(), typ.Name()) // placeholder, needs fix
	} else {
		typeName = typ.String()
	}
	pkgPath := typ.PkgPath()
	pkgName := ""
	if pkgPath != "" {
		// This is a simplification. `go/packages` might give a different
		// name than the last part of the path for complex module setups.
		// But for typical cases, this is a reasonable approximation.
		parts := strings.Split(pkgPath, "/")
		pkgName = parts[len(parts)-1]
	}

	if pkgName != "" {
		typeName = fmt.Sprintf("%s.%s", pkgName, typ.Name())
	} else {
		typeName = typ.String()
	}


	if _, ok := v.adapters[typeName]; !ok {
		// Fallback for anonymous structs or other edge cases.
		if _, ok := v.adapters[typ.String()]; !ok {
			return NewFatalError(fmt.Sprintf("no TypeAdapter registered for type %s or %s", typeName, typ.String()))
		}
		typeName = typ.String()
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
	pkgPath := typ.PkgPath()
	pkgName := ""
	if pkgPath != "" {
		parts := strings.Split(pkgPath, "/")
		pkgName = parts[len(parts)-1]
	}

	var typeName string
	if pkgName != "" {
		typeName = fmt.Sprintf("%s.%s", pkgName, typ.Name())
	} else {
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

		// Apply type rules using the objectEnv.
		objectVars := map[string]any{"self": objMap}
		for _, rule := range ruleSet.TypeRules {
			prog, err := v.engine.getProgram(v.objectEnv, rule)
			if err != nil {
				v.logger.Error("failed to compile type rule", "rule", rule, "type", typeName, "error", err)
				*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("type rule compilation error for %s: %s", typeName, err)))
				continue
			}

			out, _, err := prog.Eval(objectVars)
			if err != nil {
				v.logger.Error("failed to evaluate type rule", "rule", rule, "type", typeName, "error", err)
				*allErrors = append(*allErrors, NewValidationError(typeName, "", fmt.Sprintf("evaluation error: %s", err)))
				continue
			}

			if valid, ok := out.Value().(bool); !ok || !valid {
				*allErrors = append(*allErrors, NewValidationError(typeName, "", rule))
			}
		}

		// Apply field rules using the fieldEnv.
		for fieldName, rules := range ruleSet.FieldRules {
			fieldVal, ok := objMap[fieldName]
			if !ok {
				v.logger.Warn("field not found in adapted map", "field", fieldName, "type", typeName)
				continue
			}
			fieldVars := map[string]any{"self": fieldVal}

			for _, rule := range rules {
				prog, err := v.engine.getProgram(v.fieldEnv, rule)
				if err != nil {
					v.logger.Error("failed to compile field rule", "rule", rule, "type", typeName, "field", fieldName, "error", err)
					*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("field rule compilation error for %s.%s: %s", typeName, fieldName, err)))
					continue
				}

				out, _, err := prog.Eval(fieldVars)
				if err != nil {
					v.logger.Error("failed to evaluate field rule", "rule", rule, "type", typeName, "field", fieldName, "error", err)
					*allErrors = append(*allErrors, NewValidationError(typeName, fieldName, fmt.Sprintf("evaluation error: %s", err)))
					continue
				}

				if valid, ok := out.Value().(bool); !ok || !valid {
					*allErrors = append(*allErrors, NewValidationError(typeName, fieldName, rule))
				}
			}
		}
	}
	// --- Recursive Validation Step ---
	// Iterate over the fields of the struct to find nested structs, slices, and maps.
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)

		// Ensure we can get an interface to the field to pass to the recursive call.
		if !fieldVal.CanInterface() {
			continue
		}

		switch fieldVal.Kind() {
		case reflect.Struct:
			v.validateRecursive(fieldVal.Interface(), allErrors)

		case reflect.Ptr:
			// Only recurse on pointers to structs.
			if fieldVal.Type().Elem().Kind() == reflect.Struct {
				v.validateRecursive(fieldVal.Interface(), allErrors)
			}

		case reflect.Slice:
			// Iterate over slice elements and validate them if they are structs.
			for j := 0; j < fieldVal.Len(); j++ {
				elem := fieldVal.Index(j)
				if elem.CanInterface() {
					v.validateRecursive(elem.Interface(), allErrors)
				}
			}

		case reflect.Map:
			// Iterate over map values and validate them if they are structs.
			iter := fieldVal.MapRange()
			for iter.Next() {
				elem := iter.Value()
				if elem.CanInterface() {
					v.validateRecursive(elem.Interface(), allErrors)
				}
			}
		}
	}
}
