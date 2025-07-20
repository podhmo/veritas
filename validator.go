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

// TypeAdapterFunc is the function signature for converting a Go object.
type TypeAdapterFunc func(obj any) (map[string]any, error)

// TypeAdapterTarget specifies the target rule set for a type adapter.
type TypeAdapterTarget struct {
	TargetName string
	Adapter    TypeAdapterFunc
}

// Validator performs validation on Go objects based on a set of rules.
type Validator struct {
	engine      *Engine
	objectEnv   *cel.Env // For object-level rules (e.g., self.field > 10)
	fieldEnv    *cel.Env // For field-level rules (e.g., self.size() > 0)
	rules       map[string]ValidationRuleSet
	adapters    map[reflect.Type]TypeAdapterTarget
	logger      *slog.Logger
	nativeTypes map[reflect.Type]struct{}
	nativeEnvs  map[reflect.Type]*cel.Env // Cache for type-specific native environments
}

// ValidatorOption is an option for configuring a Validator.
type ValidatorOption func(*validatorOptions)

type validatorOptions struct {
	engine      *Engine
	provider    RuleProvider
	logger      *slog.Logger
	adapters    map[reflect.Type]TypeAdapterTarget
	types       []any
	nativeTypes map[reflect.Type]struct{}
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
func WithTypeAdapters(adapters map[reflect.Type]TypeAdapterTarget) ValidatorOption {
	return func(o *validatorOptions) {
		for k, v := range adapters {
			o.adapters[k] = v
		}
	}
}

// WithTypes enables native struct validation for the given types.
func WithTypes(types ...any) ValidatorOption {
	return func(o *validatorOptions) {
		o.types = append(o.types, types...)
	}
}

// NewValidator creates a new validator with the given options.
// If no rule provider is specified, it defaults to using the global registry.
func NewValidator(opts ...ValidatorOption) (*Validator, error) {
	// Default options
	options := &validatorOptions{
		logger:      slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
		adapters:    make(map[reflect.Type]TypeAdapterTarget),
		types:       []any{},
		nativeTypes: make(map[reflect.Type]struct{}),
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

	var objectEnv *cel.Env
	var fieldEnv *cel.Env

	// Environment for object-level validation where 'self' is a map (adapter path).
	objOpts := make([]cel.EnvOption, len(options.engine.baseOpts))
	copy(objOpts, options.engine.baseOpts)
	objOpts = append(objOpts, cel.Variable("self", types.NewMapType(types.StringType, types.DynType)))
	oenv, err := cel.NewEnv(objOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create object CEL environment: %w", err)
	}
	objectEnv = oenv

	// Environment for field-level validation where 'self' is a dynamic type.
	fieldOpts := make([]cel.EnvOption, len(options.engine.baseOpts))
	copy(fieldOpts, options.engine.baseOpts)
	fieldOpts = append(fieldOpts, cel.Variable("self", types.DynType))
	fieldOpts = append(fieldOpts, cel.StdLib()) // Add standard functions like size() and contains()
	fenv, err := cel.NewEnv(fieldOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create field CEL environment: %w", err)
	}
	fieldEnv = fenv

	v := &Validator{
		engine:      options.engine,
		objectEnv:   objectEnv,
		fieldEnv:    fieldEnv,
		rules:       rules,
		adapters:    options.adapters,
		logger:      options.logger,
		nativeTypes: options.nativeTypes,
		nativeEnvs:  make(map[reflect.Type]*cel.Env),
	}

	// Pre-create native environments for all registered types
	if len(options.types) > 0 {
		for _, t := range options.types {
			typ := reflect.TypeOf(t)
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			options.nativeTypes[typ] = struct{}{} // Add to options, not v
			if _, err := v.getNativeEnv(typ); err != nil {
				return nil, fmt.Errorf("failed to create initial native env for %v: %w", typ, err)
			}
		}
	}

	return v, nil
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

	// Use a helper function to perform the validation recursively.
	v.validateRecursive(ctx, obj, &allErrors)

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}

// getTypeName constructs a full type name string (e.g., "github.com/foo/bar/baz.User") from a reflect.Type.
// For generic types, it attempts to find a matching generic rule definition (e.g., "...Box[T]").
func (v *Validator) getTypeName(typ reflect.Type) string {
	name := typ.Name()
	pkgPath := typ.PkgPath()
	fullName := name
	if pkgPath != "" {
		fullName = fmt.Sprintf("%s.%s", pkgPath, name)
	}

	// If the type is a generic instantiation (e.g., "Box[string]"),
	// try to find the corresponding generic rule key (e.g., "...Box[T]").
	if genericMarkerPos := strings.Index(name, "["); genericMarkerPos != -1 {
		baseName := name[:genericMarkerPos]
		fullBaseName := baseName
		if pkgPath != "" {
			fullBaseName = fmt.Sprintf("%s.%s", pkgPath, baseName)
		}

		// Search for a rule key that starts with the base name and has a generic marker.
		for key := range v.rules {
			if strings.HasPrefix(key, fullBaseName+"[") {
				return key // Found it, e.g., "github.com/podhmo/veritas/testdata/sources.Box[T]"
			}
		}
	}

	// For non-generic types or if no generic rule is found, return the full type name.
	return fullName
}

// dereferenceAndAdapt handles the crucial step of preparing a value for CEL evaluation.
// It dereferences pointers and, if the underlying value is a struct with a registered
// TypeAdapter, it uses the adapter to convert the struct to a map[string]any.
func (v *Validator) dereferenceAndAdapt(value any) any {
	rv := reflect.ValueOf(value)

	// We only need to do something special for pointers.
	if rv.Kind() != reflect.Ptr {
		// If the value is a struct, it might have a direct adapter.
		if rv.Kind() == reflect.Struct {
			if adapterTarget, ok := v.adapters[rv.Type()]; ok {
				adapted, err := adapterTarget.Adapter(value)
				if err == nil {
					return adapted // Success! Return the map.
				}
				v.logger.Warn("TypeAdapter failed for value", "type", rv.Type(), "error", err)
			}
		}
		return value
	}
	if rv.IsNil() {
		return nil // CEL handles nil as 'null'.
	}

	elem := rv.Elem()
	elemInterface := elem.Interface()

	// If the pointer points to a struct, try to adapt it.
	if elem.Kind() == reflect.Struct {
		if adapterTarget, ok := v.adapters[elem.Type()]; ok {
			adapted, err := adapterTarget.Adapter(elemInterface)
			if err == nil {
				return adapted // Success! Return the map.
			}
			v.logger.Warn("TypeAdapter failed for value", "type", elem.Type(), "error", err)
		}
	}

	// For non-struct pointers, or if the adapter fails, return the dereferenced value.
	return elemInterface
}

// getNativeEnv creates and caches a CEL environment for a specific native Go type.
// This is necessary to avoid "overlapping identifier" errors when defining "self"
// for different struct types.
func (v *Validator) getNativeEnv(typ reflect.Type) (*cel.Env, error) {
	if env, ok := v.nativeEnvs[typ]; ok {
		return env, nil
	}

	v.logger.Debug("creating new native CEL environment for type", "type", typ)

	// The environment needs to know about the specific type (`typ`) for `self`,
	// and also needs the standard library functions.
	envOpts := make([]cel.EnvOption, len(v.engine.baseOpts))
	copy(envOpts, v.engine.baseOpts)
	envOpts = append(envOpts, cel.StdLib()) // Add standard functions
	envOpts = append(envOpts, ext.NativeTypes(typ))
	envOpts = append(envOpts, cel.Variable("self", cel.ObjectType(typ.String())))

	env, err := cel.NewEnv(envOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create native CEL environment for type %v: %w", typ, err)
	}

	v.nativeEnvs[typ] = env
	return env, nil
}

// isNativeType checks if a given reflect.Type is configured for native validation.
func (v *Validator) isNativeType(typ reflect.Type) bool {
	// A type is native if it was registered via WithTypes.
	_, ok := v.nativeTypes[typ]
	return ok
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

	// Determine which validation path to take for the current object.
	if v.isNativeType(typ) {
		v.validateNative(ctx, val.Interface(), typ, allErrors)
	} else {
		// Default to adapter-based path if not explicitly native.
		// This handles types with adapters and types with no rules.
		v.validateWithAdapter(ctx, val.Interface(), typ, allErrors)
	}

	// --- Common Recursive Validation Step for Nested Fields ---
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
			if !fieldVal.IsNil() && fieldVal.Type().Elem().Kind() == reflect.Struct {
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

// validateNative handles validation using the native CEL environment.
func (v *Validator) validateNative(ctx context.Context, obj any, typ reflect.Type, allErrors *[]error) {
	typeName := v.getTypeName(typ)
	ruleSet, hasRules := v.rules[typeName]
	if !hasRules {
		return
	}

	nativeEnv, err := v.getNativeEnv(typ)
	if err != nil {
		v.logger.Error("failed to get native env", "type", typ, "error", err)
		*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("env creation error for %s: %s", typeName, err)))
		return
	}

	// Type Rules use the native object directly.
	objectVars := map[string]any{"self": obj}
	for _, rule := range ruleSet.TypeRules {
		prog, err := v.engine.getProgram(nativeEnv, rule)
		if err != nil {
			// This is a workaround. For generic types, a rule like `self.Value != null` can
			// fail with "no matching overload" if `Value` is a non-nullable type like `string`.
			// We choose to ignore this specific compilation error.
			if strings.Contains(err.Error(), "no matching overload") {
				v.logger.Debug("ignored 'no matching overload' compilation error for generic type rule", "rule", rule, "type", typeName, "error", err)
				continue // Skip this rule
			}
			v.logger.Error("failed to compile type rule (native)", "rule", rule, "type", typeName, "error", err)
			*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("type rule compilation error for %s: %s", typeName, err)))
			continue
		}

		out, _, err := prog.ContextEval(ctx, objectVars)
		if err != nil {
			// This is a workaround. For generic types, a rule like `self.Value != null` can
			// fail with "no matching overload" if `Value` is a non-nullable type like `string`.
			// We choose to ignore this specific error, assuming that a field-level rule
			// will correctly validate the non-null aspect (e.g., `self != ""`).
			if strings.Contains(err.Error(), "no matching overload") {
				v.logger.Debug("ignored 'no matching overload' error for generic type rule", "rule", rule, "type", typeName, "error", err)
			} else {
				v.logger.Error("failed to evaluate type rule (native)", "rule", rule, "type", typeName, "error", err)
				*allErrors = append(*allErrors, NewValidationError(typeName, "", fmt.Sprintf("evaluation error: %s", err)))
			}
			continue
		}

		if valid, ok := out.Value().(bool); !ok || !valid {
			*allErrors = append(*allErrors, NewValidationError(typeName, "", rule))
		}
	}

	// Field Rules are evaluated against the field's value.
	val := reflect.ValueOf(obj) // We know obj is a struct here
	for fieldName, rules := range ruleSet.FieldRules {
		fieldVal := val.FieldByName(fieldName)
		if !fieldVal.IsValid() {
			v.logger.Warn("field not found in native struct", "field", fieldName, "type", typeName)
			continue
		}

		var fieldInterface any
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				// Convert nil pointer to CEL null for rules like `self != null`.
				fieldInterface = types.DefaultTypeAdapter.NativeToValue(nil)
			} else {
				// For non-nil pointers, we need to decide what to pass to CEL.
				// If we pass the pointer itself, CEL might not handle it.
				// If we pass the element, we lose pointer context.
				// The `dereferenceAndAdapt` logic is a good reference.
				// For native validation, let's pass the dereferenced element for now.
				// This allows validating fields of the pointed-to struct.
				fieldInterface = fieldVal.Elem().Interface()
			}
		} else {
			fieldInterface = fieldVal.Interface()
		}

		// If the field is a struct that should be validated natively, we skip the CEL rules here
		// because the main recursive validation loop will handle it.
		// We only apply CEL field rules to non-structs or structs handled by adapters.
		fieldTyp := reflect.TypeOf(fieldInterface)
		if fieldTyp != nil && fieldTyp.Kind() == reflect.Struct && v.isNativeType(fieldTyp) {
			// This struct will be handled by a subsequent call to validateRecursive, so we skip field rules here.
		} else {
			fieldVars := map[string]any{"self": fieldInterface}

			for _, rule := range rules {
				// We use the simpler `fieldEnv` for this, as it expects `self` to be dynamic.
				prog, err := v.engine.getProgram(v.fieldEnv, rule)
				if err != nil {
					v.logger.Error("failed to compile field rule (native)", "rule", rule, "type", typeName, "field", fieldName, "error", err)
					*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("field rule compilation error for %s.%s: %s", typeName, fieldName, err)))
					continue
				}

				out, _, err := prog.ContextEval(ctx, fieldVars)
				if err != nil {
					// Check for the specific "unsupported conversion" error and provide a better message.
					if strings.Contains(err.Error(), "unsupported conversion") {
						v.logger.Error("unsupported conversion in native field rule", "rule", rule, "type", typeName, "field", fieldName, "value_type", reflect.TypeOf(fieldInterface), "error", err)
						*allErrors = append(*allErrors, NewValidationError(typeName, fieldName, fmt.Sprintf("unsupported type for native validation: %T", fieldInterface)))
					} else {
						v.logger.Error("failed to evaluate field rule (native)", "rule", rule, "type", typeName, "field", fieldName, "error", err)
						*allErrors = append(*allErrors, NewValidationError(typeName, fieldName, fmt.Sprintf("evaluation error: %s", err)))
					}
					continue
				}

				if valid, ok := out.Value().(bool); !ok || !valid {
					*allErrors = append(*allErrors, NewValidationError(typeName, fieldName, rule))
				}
			}
		}
	}
}

// validateWithAdapter handles validation using the adapter-based CEL environment.
func (v *Validator) validateWithAdapter(ctx context.Context, obj any, typ reflect.Type, allErrors *[]error) {
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

	adapterTarget, hasAdapter := v.adapters[typ]
	ruleSet, hasRules := v.rules[typeName]

	if hasAdapter {
		v.logger.Debug("found TypeAdapter", "source_type", typ, "target_rules", adapterTarget.TargetName)
		ruleSet, hasRules = v.rules[adapterTarget.TargetName]
		typeName = adapterTarget.TargetName
	}

	if !hasRules {
		return
	}

	var objMap map[string]any
	var err error

	if hasAdapter {
		objMap, err = adapterTarget.Adapter(obj)
	} else {
		v.logger.Debug("no TypeAdapter, cannot perform CEL validation, but continuing to recurse", "type", typeName)
		return
	}

	if err != nil {
		v.logger.Error("failed to adapt object", "type", typeName, "error", err)
		*allErrors = append(*allErrors, NewFatalError(fmt.Sprintf("TypeAdapter error for %s: %v", typeName, err)))
		return
	}

	if objMap != nil {
		// Apply type rules using the objectEnv.
		adaptedMapForTypeRules := make(map[string]any, len(objMap))
		for k, val := range objMap {
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
			fieldVal, ok := objMap[fieldName]
			if !ok {
				v.logger.Warn("field not found in adapted map", "field", fieldName, "type", typeName)
				continue
			}

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
