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
	"github.com/google/cel-go/ext"
)

// Validator performs validation on Go objects based on a set of rules.
type Validator struct {
	engine    *Engine
	objectEnv *cel.Env // For object-level rules (e.g., self.field > 10)
	fieldEnv  *cel.Env // For field-level rules (e.g., self.size() > 0)
	rules     map[string]ValidationRuleSet
	types     []reflect.Type
	logger    *slog.Logger
}

// ValidatorOption is an option for configuring a Validator.
type ValidatorOption func(*validatorOptions)

type validatorOptions struct {
	engine   *Engine
	provider RuleProvider
	logger   *slog.Logger
	types    []reflect.Type
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

// WithTypes registers Go struct types with the validator.
// This allows the validator to work with native Go structs instead of maps.
func WithTypes(types ...any) ValidatorOption {
	return func(o *validatorOptions) {
		for _, t := range types {
			rt := reflect.TypeOf(t)
			if rt.Kind() == reflect.Ptr {
				rt = rt.Elem()
			}
			o.types = append(o.types, rt)
		}
	}
}

// NewValidator creates a new validator with the given options.
// If no rule provider is specified, it defaults to using the global registry.
func NewValidator(opts ...ValidatorOption) (*Validator, error) {
	// Default options
	options := &validatorOptions{
		logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
		types:  make([]reflect.Type, 0),
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

	for k := range rules {
		options.logger.Debug("loaded rule", "key", k)
	}

	// Create CEL environment options for native types.
	objEnvOpts := options.engine.baseOpts
	if len(options.types) > 0 {
		// Register the provided Go types with CEL.
		// ext.NativeTypes accepts a slice of either reflect.Type or struct instances.
		// We have []reflect.Type, so we convert it to []any.
		typesAsAny := make([]any, len(options.types))
		for i, t := range options.types {
			typesAsAny[i] = t
		}
		objEnvOpts = append(objEnvOpts, ext.NativeTypes(typesAsAny...))
	}
	// 'self' can be any of the registered types, so we use DynType.
	objEnvOpts = append(objEnvOpts, cel.Variable("self", types.DynType))

	objectEnv, err := cel.NewEnv(objEnvOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create object CEL environment: %w", err)
	}

	// Environment for field-level validation where 'self' is a dynamic type.
	fieldOpts := make([]cel.EnvOption, len(options.engine.baseOpts))
	copy(fieldOpts, options.engine.baseOpts)
	if len(options.types) > 0 {
		// Also register types for the field environment to handle complex types.
		typesAsAny := make([]any, len(options.types))
		for i, t := range options.types {
			typesAsAny[i] = t
		}
		fieldOpts = append(fieldOpts, ext.NativeTypes(typesAsAny...))
	}
	fieldOpts = append(fieldOpts, cel.Variable("self", types.DynType))
	fieldOpts = append(fieldOpts, cel.Declarations()) // Add standard declarations
	fieldEnv, err := cel.NewEnv(fieldOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create field CEL environment: %w", err)
	}

	return &Validator{
		engine:    options.engine,
		objectEnv: objectEnv,
		fieldEnv:  fieldEnv,
		rules:     rules,
		types:     options.types,
		logger:    options.logger,
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
	if pkgPath := typ.PkgPath(); pkgPath != "" {
		parts := strings.Split(pkgPath, "/")
		_ = parts[len(parts)-1]
		// typ.String() can be "pkg.Type" or "pkg.Type[T]"
		// typ.Name() is just "Type" or "Type[T]"
		// We need to construct "pkgName.Name"
	} else {
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
		parts := strings.Split(pkgPath, "/")
		pkgName := parts[len(parts)-1]
		return fmt.Sprintf("%s.%s", pkgName, typ.Name())
	}
	return typ.String() // Handle anonymous or built-in types
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

	// Dereference pointer to get the actual value and its type.
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
	if genericMarkerPos := strings.LastIndex(typeName, "["); genericMarkerPos != -1 {
		baseName := typeName[:genericMarkerPos]
		v.logger.Debug("detected generic type", "original", typeName, "base", baseName)
		if typeSpecName, ok := v.getGenericTypeName(baseName); ok {
			v.logger.Debug("found matching generic rule", "from", baseName, "to", typeSpecName)
			typeName = typeSpecName
		}
	}

	// If there are rules for this type, validate it.
	if ruleSet, hasRules := v.rules[typeName]; hasRules {
		// The object to be validated. If the original `obj` was a pointer,
		// `val.Interface()` is the struct value. If `obj` was a struct value,
		// we pass it directly.
		var objectToValidate any
		if reflect.ValueOf(obj).Kind() == reflect.Ptr {
			objectToValidate = val.Interface()
		} else {
			objectToValidate = obj
		}
		objectVars := map[string]any{"self": objectToValidate}

		// Apply type rules.
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

		// Apply field rules.
		for fieldName, rules := range ruleSet.FieldRules {
			// Check for context cancellation before each field validation.
			select {
			case <-ctx.Done():
				*allErrors = append(*allErrors, ctx.Err())
				return
			default:
			}

			// Get the field value using reflection.
			fieldVal := val.FieldByName(fieldName)
			if !fieldVal.IsValid() {
				v.logger.Warn("field not found in struct", "field", fieldName, "type", typeName)
				continue
			}

			// For field rules, 'self' is the field's value itself.
			// We need to handle pointers carefully. If the value is a pointer,
			// we must dereference it before passing it to CEL.
			fieldInterface := fieldVal.Interface()
			rv := reflect.ValueOf(fieldInterface)
			if rv.Kind() == reflect.Ptr {
				if rv.IsNil() {
					fieldInterface = nil // Use nil for CEL's 'null'
				} else {
					fieldInterface = rv.Elem().Interface() // Dereference non-nil pointers
				}
			}
			fieldVars := map[string]any{"self": fieldInterface}

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
