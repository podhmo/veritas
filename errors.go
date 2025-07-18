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

// This file can be used to define custom error types for more granular
// error handling by the library's users. For example, a specific
// ValidationError type could be defined.

// For now, we will rely on wrapping errors with fmt.Errorf and joining
// them with errors.Join.
