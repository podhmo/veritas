package veritas

// This file can be used to define custom error types for more granular
// error handling by the library's users. For example, a specific
// ValidationError type could be defined.

// For now, we will rely on wrapping errors with fmt.Errorf and joining
// them with errors.Join.