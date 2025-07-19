package veritas

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
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
	engine       *Engine
	objectEnv    *cel.Env // For object-level rules (e.g., self.field > 10)
	fieldEnv     *cel.Env // For field-level rules (e.g., self.size() > 0)
	rules        map[string]ValidationRuleSet
	genericRules map[string]ValidationRuleSet // Map from base type name to rule set
	adapters     map[string]TypeAdapter
	logger       *slog.Logger
}

// ValidatorOption is an option for configuring a Validator.
type ValidatorOption func(*validatorOptions)

type validatorOptions struct {
	engine   *Engine
	provider RuleProvider
	logger   *slog.Logger
	adapters map[string]TypeAdapter
}

// WithEngine sets the CEL engine for the validator.
func WithEngine(engine *Engine) ValidatorOption {
	return func(o *validatorOptions) {
		o.engine = engine
	}
}

// WithRuleProvider sets the rule provider for the validator.
func WithRuleProvider(provider RuleProvider) ValidatorOption {
	return func(o *validatorOptions) {
		o.provider = provider
	}
}

// WithLogger sets the logger for the validator.
func WithLogger(logger *slog.Logger) ValidatorOption {
	return func(o *validatorOptions) {
		o.logger = logger
	}
}

// WithTypeAdapters sets the type adapters for the validator.
func WithTypeAdapters(adapters map[string]TypeAdapter) ValidatorOption {
	return func(o *validatorOptions) {
		for k, v := range adapters {
			o.adapters[k] = v
		}
	}
}

// NewValidator creates a new validator with the given options.
// If no rule provider is specified, it defaults to using the global registry.
func NewValidator(opts ...ValidatorOption) (*Validator, error) {
	// Default options
	options := &validatorOptions{
		logger:   slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})),
		adapters: make(map[string]TypeAdapter),
	}

	// Apply user-provided options
	for _, opt := range opts {
		opt(options)
	}

	// Default engine if not provided
	if options.engine == nil {
		engine, err := NewEngine(options.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create default engine: %w", err)
		}
		options.engine = engine
	}

	// Default provider if not provided
	if options.provider == nil {
		options.provider = NewRuleProviderFromRegistry()
	}

	rules, err := options.provider.GetRuleSets()
	if err != nil {
		return nil, fmt.Errorf("failed to get rule sets: %w", err)
	}

	// Pre-process rules to separate generic rules for efficient lookup.
	genericRules := make(map[string]ValidationRuleSet)
	for key, ruleSet := range rules {
		if ruleSet.GenericTypeName != "" {
			// The key is the base name (e.g., "mypackage.Box")
			genericRules[key] = ruleSet
			// The original `rules` map can keep the specific generic rule,
			// or we can remove it. For now, let's keep it for simplicity,
			// as it won't be matched directly by non-generic types.
			options.logger.Debug("registered generic rule", "baseName", key, "signature", ruleSet.GenericTypeName)
		} else {
			options.logger.Debug("loaded rule", "key", key)
		}
	}

	// Environment for object-level validation where 'self' is a map.
	objOpts := make([]cel.EnvOption, len(options.engine.baseOpts))
	copy(objOpts, options.engine.baseOpts)
	objOpts = append(objOpts, cel.Variable("self", types.NewMapType(types.StringType, types.DynType)))
	objectEnv, err := cel.NewEnv(objOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create object CEL environment: %w", err)
	}

	// Environment for field-level validation where 'self' is a dynamic type.
	fieldOpts := make([]cel.EnvOption, len(options.engine.baseOpts))
	copy(fieldOpts, options.engine.baseOpts)
	fieldOpts = append(fieldOpts, cel.Variable("self", types.DynType))
	fieldEnv, err := cel.NewEnv(fieldOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create field CEL environment: %w", err)
	}

	return &Validator{
		engine:       options.engine,
		objectEnv:    objectEnv,
		fieldEnv:     fieldEnv,
		rules:        rules,
		genericRules: genericRules,
		adapters:     options.adapters,
		logger:       options.logger,
	}, nil
}

// Validate applies the configured rules to the given object, including nested structs.
func (v *Validator) Validate(ctx context.Context, obj any) error {
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
		parts := strings.Split(pkgPath, "/")
		pkgName := parts[len(parts)-1]
		// typ.String() can be "pkg.Type" or "pkg.Type[T]"
		// typ.Name() is just "Type" or "Type[T]"
		// We need to construct "pkgName.Name"
		typeName = fmt.Sprintf("%s.%s", pkgName, typ.Name())
	} else {
		typeName = typ.String()
	}

	// Also normalize the top-level type name if it's generic.
	originalTypeName := typeName
	if genericMarkerPos := strings.LastIndex(typeName, "["); genericMarkerPos != -1 {
		baseName := typeName[:genericMarkerPos]
		v.logger.Debug("detected generic type at top level", "original", typeName, "base", baseName)
		if typeSpecName, ok := v.getGenericTypeName(baseName); ok {
			v.logger.Debug("found matching generic rule at top level", "from", baseName, "to", typeSpecName)
			typeName = typeSpecName
		}
	}

	if _, ok := v.adapters[typeName]; !ok {
		// Fallback for anonymous structs or other edge cases.
		if _, ok := v.adapters[originalTypeName]; !ok {
			return NewFatalError(fmt.Sprintf("no TypeAdapter registered for type %s or %s", typeName, originalTypeName))
		}
		typeName = originalTypeName
	}

	// Use a helper function to perform the validation recursively.
	v.validateRecursive(ctx, obj, &allErrors)

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}

// getTypeName constructs a predictable type name string (e.g., "sources.User") from a reflect.Type.
func (v *Validator) getTypeName(typ reflect.Type) string {
	if pkgPath := typ.PkgPath(); pkgPath != "" {
		name := typ.Name()
		if genericMarkerPos := strings.LastIndex(name, "["); genericMarkerPos != -1 {
			name = name[:genericMarkerPos]
		}
		return fmt.Sprintf("%s.%s", pkgPath, name)
	}
	return typ.String() // Handle anonymous or built-in types
}

// dereferenceAndAdapt handles the crucial step of preparing a value for CEL evaluation.
// It dereferences pointers and, if the underlying value is a struct with a registered
// TypeAdapter, it uses the adapter to convert the struct to a map[string]any.
func (v *Validator) dereferenceAndAdapt(value any) any {
	rv := reflect.ValueOf(value)

	// We only need to do something special for pointers.
	if rv.Kind() != reflect.Ptr {
		return value
	}
	if rv.IsNil() {
		return nil // CEL handles nil as 'null'.
	}

	elem := rv.Elem()
	elemInterface := elem.Interface()

	// If the pointer points to a struct, try to adapt it.
	if elem.Kind() == reflect.Struct {
		typeName := v.getTypeName(elem.Type())
		lookupName := typeName
		// Also check for the generic version of the type name.
		if genericMarkerPos := strings.LastIndex(typeName, "["); genericMarkerPos != -1 {
			baseName := typeName[:genericMarkerPos]
			if _, ok := v.genericRules[baseName]; ok {
				lookupName = baseName
			}
		}

		if adapter, ok := v.adapters[lookupName]; ok {
			adapted, err := adapter(elemInterface)
			if err == nil {
				return adapted // Success! Return the map.
			}
			v.logger.Warn("TypeAdapter failed for value", "type", typeName, "error", err)
		}
	}

	// For non-struct pointers, or if the adapter fails, return the dereferenced value.
	return elemInterface
}

// validateRecursive is the internal helper that performs the actual validation.
func (v *Validator) validateRecursive(ctx context.Context, obj any, allErrors *[]error) {
	// Check for context cancellation before proceeding.
	select {
	case <-ctx.Done():
		*allErrors = append(*allErrors, ctx.Err())
		return
	default:
	}

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
	typeName := v.getTypeName(typ)

	// Normalize generic type names for rule lookup.
	lookupName := typeName
	isGeneric := false
	if genericMarkerPos := strings.LastIndex(typeName, "["); genericMarkerPos != -1 {
		baseName := typeName[:genericMarkerPos]
		v.logger.Debug("detected generic type", "original", typeName, "base", baseName)
		if _, ok := v.genericRules[baseName]; ok {
			lookupName = baseName
			isGeneric = true
		}
	}

	// Get the adapter for the current type.
	// For generic types, we look up the adapter using the base name.
	adapter, hasAdapter := v.adapters[lookupName]
	if !hasAdapter {
		v.logger.Debug("no TypeAdapter, but continuing to recurse", "type", lookupName)
	}

	// Get the rule set.
	var ruleSet ValidationRuleSet
	var hasRules bool
	if isGeneric {
		ruleSet, hasRules = v.genericRules[lookupName]
	} else {
		ruleSet, hasRules = v.rules[lookupName]
	}

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
		adaptedMapForTypeRules := make(map[string]any, len(objMap))
		for k, val := range objMap {
			adaptedMapForTypeRules[k] = val // Start with the original value
			// For type rules, we need to adapt the values *within* the 'self' map.
			adaptedMapForTypeRules[k] = v.dereferenceAndAdapt(val)
		}
		objectVars := map[string]any{"self": adaptedMapForTypeRules}

		for _, rule := range ruleSet.TypeRules {
			prog, err := v.engine.getProgram(v.objectEnv, rule)
			if err != nil {
				v.logger.Error("failed to compile type rule", "rule", rule, "type", typeName, "error", err)
				*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("type rule compilation error for %s: %s", typeName, err)))
				continue
			}

			out, _, err := prog.ContextEval(ctx, objectVars)
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
			// Check for context cancellation before each field validation.
			select {
			case <-ctx.Done():
				*allErrors = append(*allErrors, ctx.Err())
				return
			default:
			}

			fieldVal, ok := objMap[fieldName]
			if !ok {
				v.logger.Warn("field not found in adapted map", "field", fieldName, "type", typeName)
				continue
			}

			// For field rules, 'self' is the field's value itself.
			// We adapt it before passing it to CEL.
			adaptedFieldVal := v.dereferenceAndAdapt(fieldVal)
			fieldVars := map[string]any{"self": adaptedFieldVal}

			for _, rule := range rules {
				prog, err := v.engine.getProgram(v.fieldEnv, rule)
				if err != nil {
					v.logger.Error("failed to compile field rule", "rule", rule, "type", typeName, "field", fieldName, "error", err)
					*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("field rule compilation error for %s.%s: %s", typeName, fieldName, err)))
					continue
				}

				out, _, err := prog.ContextEval(ctx, fieldVars)
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
			v.validateRecursive(ctx, fieldVal.Interface(), allErrors)

		case reflect.Ptr:
			// Only recurse on pointers to structs.
			if fieldVal.Type().Elem().Kind() == reflect.Struct {
				v.validateRecursive(ctx, fieldVal.Interface(), allErrors)
			}

		case reflect.Slice:
			// Iterate over slice elements and validate them if they are structs.
			for j := 0; j < fieldVal.Len(); j++ {
				elem := fieldVal.Index(j)
				if elem.CanInterface() {
					v.validateRecursive(ctx, elem.Interface(), allErrors)
				}
			}

		case reflect.Map:
			// Iterate over map values and validate them if they are structs.
			iter := fieldVal.MapRange()
			for iter.Next() {
				elem := iter.Value()
				if elem.CanInterface() {
					v.validateRecursive(ctx, elem.Interface(), allErrors)
				}
			}
		}
	}
}

// getGenericTypeName finds a generic type name from the rules map that matches a base name.
// For example, given "main.Box", it might find "main.Box[T]".
func (v *Validator) getGenericTypeName(baseName string) (string, bool) {
	// This is inefficient but works for the purpose of this library where the number
	// of rules is not expected to be astronomically large.
	// A better approach might be to pre-process the rule keys into a more searchable structure.
	for key := range v.rules {
		if strings.HasPrefix(key, baseName+"[") {
			return key, true
		}
	}
	return "", false
}

// NewValidatorFromJSONFile creates a new validator from a JSON file.
// It is a convenience function that wraps NewValidator with a JSONRuleProvider.
func NewValidatorFromJSONFile(filePath string, opts ...ValidatorOption) (*Validator, error) {
	provider := NewJSONRuleProvider(filePath)
	allOpts := []ValidatorOption{WithRuleProvider(provider)}
	allOpts = append(allOpts, opts...)
	return NewValidator(allOpts...)
}
