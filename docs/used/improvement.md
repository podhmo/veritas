# Veritas Improvement Proposal

This document outlines proposed enhancements for the Veritas validation library. The goal is to improve developer experience, expand functionality, and increase robustness.

## 1. Documentation Enhancement

*   **Problem:** The current documentation is good for getting started but lacks depth for advanced use cases. GoDoc comments are sparse, making the public API harder to understand without reading the source.
*   **Proposal:**
    *   **Expand `README.md` and `docs/`:** Add detailed sections on:
        *   Creating and using custom validation functions.
        *   A deep dive into the generics support, including limitations and best practices.
        *   Advanced usage of dynamic JSON-based rules.
        *   Detailed error handling and customization patterns.
    *   **Enrich GoDoc Comments:** Add comprehensive GoDoc comments for all exported types, functions, and methods (e.g., `Engine`, `Validator`, `ValidatorOption`, `NewValidator`).
    *   **More Examples:** Create new examples in the `examples/` directory for custom functions, complex nested structs, and generic types.

## 2. Automatic TypeAdapter Generation

*   **Problem:** When using JSON-based rules, developers must manually write `TypeAdapter` functions. This is boilerplate, error-prone, and tedious, creating a barrier to adopting dynamic validation.
*   **Proposal:**
    *   Extend the `veritas` CLI tool with a new feature to automatically generate `TypeAdapter` implementations alongside the validation rules.
    *   The generator (`cmd/veritas/gen`) will inspect the struct fields and create a `TypeAdapter` that correctly converts the struct to a `map[string]any`.
    *   This generated adapter would be registered automatically, similar to how rules are registered now, making dynamic validation as seamless as compile-time validation.

## 3. Improved Generics Support

*   **Problem:** The current generics implementation is functional but fragile. It relies on string matching for type names (e.g., `Box[T]`) and has an inefficient lookup mechanism (`getGenericTypeName`). It may not handle complex type constraints correctly.
*   **Proposal:**
    *   **Robust Type Name Parsing:** Enhance `cmd/veritas/parser` to parse generic type definitions more accurately, handling multiple type parameters and complex constraints.
    *   **Efficient Type Lookup:** Instead of linear scanning, create a pre-computed map during code generation that links a base type name (e.g., `"mypackage.Box"`) to its full generic definition (`"mypackage.Box[T]"`). This will make the runtime lookup in the `Validator` an O(1) operation.
    *   **Testing:** Add extensive tests for various generic type definitions to ensure correctness.

## 4. Enhanced Error Reporting

*   **Problem:** The current `ValidationError` provides the type, field, and violated rule, which is good but can be more informative for debugging and user feedback.
*   **Proposal:**
    *   **Add Context to Errors:** Augment `ValidationError` to optionally include the invalid value that caused the error.
    *   **Customizable Error Messages:** Introduce a mechanism to allow custom error message templates. For example, a `messages` tag could be added: `validate:"required" message:"The user name is required."`.
    *   **Structured Error Details:** Provide a helper function that converts a validation error into a structured map (e.g., `map[string]string` or `map[string][]string`), which is more useful for API responses than a single joined error string. The existing `veritas.ToErrorMap` is a good start but could be enhanced.

## 5. Performance Optimization

*   **Problem:** The library relies heavily on reflection, and some parts, like the generic type lookup, are known to be inefficient.
*   **Proposal:**
    *   **Optimize Generic Lookup:** Implement the efficient type lookup mechanism described in "Improved Generics Support."
    *   **Benchmark and Profile:** Introduce a suite of benchmarks to identify performance bottlenecks in the validation process, especially for large and deeply nested structs.
    *   **Cache Reflection Results:** Where possible, cache results from reflection-based operations to reduce overhead on subsequent validations of the same type.

## 6. CLI Feature Expansion

*   **Problem:** The `veritas` CLI is currently focused on code generation and linting. Its utility could be expanded for a better development workflow.
*   **Proposal:**
    *   **`veritas check <path>`:** A new subcommand to parse struct tags and CEL rules in a given file or package and report syntax errors without generating code. This would provide faster feedback during development.
    *   **`veritas test --rule <rule_json> --data <data_json>`:** A subcommand to test a set of validation rules against a given data object from the command line. This would be useful for debugging and testing dynamic rules.
