# Veritas Development TODO List

This document outlines the detailed, phased development plan for the "Veritas" validation library.

## Phase 1: Core Engine Construction (v0.1)

**Goal**: Build a minimal, robust engine capable of evaluating CEL expressions and performing basic type and field validation.

-   **[x] 1.1: `Engine` Implementation**
    -   [x] 1.1.1: Define an `Engine` struct that encapsulates `cel.Env` and an LRU cache (`lru.Cache`) for compiled `cel.Program`s.
    -   [x] 1.1.2: Use `log/slog` for internal tracing. Log cache misses and program generations at the `Debug` level.
    -   [x] 1.1.3: Implement a framework to register custom CEL functions (e.g., `strings.ToUpper`, `matches`).

-   **[~] 1.2: `Validator` Implementation**
    -   [x] 1.2.1: Define a `Validator` struct holding a reference to the `Engine` and a `map[string]ValidationRuleSet` for rule lookup by type name.
    -   [~] 1.2.2: Implement the primary `Validate(obj any) error` method. Use `log/slog.Error` for internal errors like reflection failures. (Note: Implementation is in progress, facing issues with dynamic type registration in cel-go).
    -   [x] 1.2.3: Implement logic to apply `TypeRules` and `FieldRules` separately, aggregating all validation failures using `errors.Join`.
    -   [x] 1.2.4: Ensure error messages are contextual, including the type and field name (e.g., `User.Email: validation failed...`).
    -   [ ] 1.2.5: **[TODO]** Resolve the `unsupported type` error in `cel-go`. Investigate the correct way to register native Go structs with the CEL environment to allow validation of struct fields. This is currently blocking all validation tests.

-   **[x] 1.3: Rule Provider Implementation**
    -   [x] 1.3.1: Define the `ValidationRuleSet` struct (containing `TypeRules`, `FieldRules`).
    -   [x] 1.3.2: Implement `JSONRuleProvider` to load rule sets from a JSON file. Log I/O or parsing errors with `slog.Error`.

-   **[~] 1.4: Unit Testing Foundation**
    -   [~] 1.4.1: Create a test suite for the `Validator` covering success, single failure, and multiple failure scenarios. (Note: Tests are written but currently failing due to the `Validate` method's issues).
    -   [x] 1.4.2: Use `github.com/google/go-cmp/cmp` for all assertions, especially `cmpopts.EquateErrors` for error comparison.
    -   [x] 1.4.3: Adhere to the `want` and `got` variable naming convention for test comparisons.

## Phase 2: Static Analysis Tool Development (v0.2)

**Goal**: Develop the `veritas` CLI tool to automatically extract validation rules from Go source code.

-   **[ ] 2.1: Go Source Code Parser**
    -   [ ] 2.1.1: Implement a parser using `go/ast` to recursively scan directories and find `struct` definitions.

-   **[ ] 2.2: Rule Extraction Logic**
    -   [ ] 2.2.1: Extract field-level rules from `validate:"..."` struct tags.
    -   [ ] 2.2.2: Extract type-level rules from special `// @cel: ...` comments preceding a `struct` definition.
    -   [ ] 2.2.3: Implement a mapping from common shorthands (`required`, `email`, etc.) to their corresponding CEL expressions.

-   **[ ] 2.3: `veritas` CLI Implementation**
    -   [ ] 2.3.1: Implement logic to output the extracted rules as a structured JSON file.
    -   [ ] 2.3.2: Use `slog.Info` for progress reporting and `slog.Debug` for detailed parsing steps.

-   **[ ] 2.4: Static Analysis Tool Testing**
    -   [ ] 2.4.1: Prepare sample Go source files and their expected JSON output as test data.
    -   [ ] 2.4.2: Write tests that run the generator and compare the actual output against the expected JSON using `go-cmp/cmp`.

## Phase 3: Advanced Data Structures Support (v0.3)

**Goal**: Enhance the library to handle complex data structures common in modern Go applications.

-   **[ ] 3.1: Pointer and Nested Struct Handling**
    -   [ ] 3.1.1: Implement recursive validation for nested structs.
    -   [ ] 3.1.2: Implement logic to safely dereference and validate pointer fields.

-   **[ ] 3.2: Slice (`[]T`) Support**
    -   [ ] 3.2.1: Support a `dive` keyword in the `validate` tag to apply rules to each element of a slice.
    -   [ ] 3.2.2: Include the array index in error messages (e.g., `User.Scores[2]: is invalid`).

-   **[ ] 3.3: Map (`map[K]V`) Support**
    -   [ ] 3.3.1: Support `keys` and `values` keywords to apply rules to a map's keys and values, respectively.
    -   [ ] 3.3.2: Include the map key in error messages (e.g., `User.Metadata['user_id']: is invalid`).

-   **[ ] 3.4: Advanced Structure Testing**
    -   [ ] 3.4.1: Expand the test suite to include complex structs with slices, maps, and nested pointers.

## Phase 4: General Availability (GA) Finalization (v1.0)

**Goal**: Support modern Go features, create comprehensive documentation, and polish the library for its official v1.0 release.

-   **[ ] 4.1: Go Generics Support**
    -   [ ] 4.1.1: Update the `veritas` tool to correctly parse generic `struct` definitions.
    -   [ ] 4.1.2: Ensure the runtime `Validator` can correctly handle instantiated generic types via reflection.

-   **[ ] 4.2: Performance and Stabilization**
    -   [ ] 4.2.1: Establish a benchmark suite to identify and optimize performance bottlenecks.
    -   [ ] 4.2.2: Add `context.Context` to the `Validate` method signature to support timeouts and cancellation.

-   **[ ] 4.3: Documentation and Ecosystem**
    -   [ ] 4.3.1: Create a comprehensive documentation website detailing installation, usage, CLI commands, and all supported rules/shorthands.
    -   [ ] 4.3.2: Develop an example project demonstrating integration with a standard `net/http` server.
        -   [ ] 4.3.2.1: Show how to decode a JSON request, run validation, and return a structured HTTP 400 error response.
        -   [ ] 4.3.2.2: Use `slog` for structured request logging.

-   **[ ] 4.4: Final API Review and Testing**
    -   [ ] 4.4.1: Implement end-to-end tests for the `net/http` example.
    -   [ ] 4.4.2: Conduct a final review of all public APIs to ensure stability for the v1.0 release.