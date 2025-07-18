package veritas

import (
	"fmt"
)

// ValidationError represents a failure of a specific validation rule.
type ValidationError struct {
	TypeName  string
	FieldName string
	Rule      string
}

func (e *ValidationError) Error() string {
	if e.FieldName == "" {
		return fmt.Sprintf("%s: validation failed, rule: %s", e.TypeName, e.Rule)
	}
	return fmt.Sprintf("%s.%s: validation failed, rule: %s", e.TypeName, e.FieldName, e.Rule)
}

// NewValidationError creates a new validation error.
func NewValidationError(typeName, fieldName, rule string) error {
	return &ValidationError{
		TypeName:  typeName,
		FieldName: fieldName,
		Rule:      rule,
	}
}

// FatalError represents a critical, non-recoverable error during validation,
// such as a rule compilation failure.
type FatalError struct {
	Message string
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("veritas fatal error: %s", e.Message)
}

// NewFatalError creates a new fatal error.
func NewFatalError(message string) error {
	return &FatalError{Message: message}
}

// ToErrorMap converts a validation error into a map of field names to error messages.
// If the error is not a composition of ValidationErrors, it returns nil.
func ToErrorMap(err error) map[string]string {
	var validationErrs []*ValidationError
	if errs, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range errs.Unwrap() {
			if ve, ok := e.(*ValidationError); ok {
				validationErrs = append(validationErrs, ve)
			}
		}
	} else if ve, ok := err.(*ValidationError); ok {
		validationErrs = append(validationErrs, ve)
	}

	if len(validationErrs) == 0 {
		return nil
	}

	errMap := make(map[string]string)
	for _, ve := range validationErrs {
		// Use FieldName for field-specific errors, and a general key for type-level errors.
		key := ve.FieldName
		if key == "" {
			key = ve.TypeName
		}
		errMap[key] = ve.Rule
	}
	return errMap
}
