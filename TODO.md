# Veritas Development TODO List

This document outlines the detailed, phased development plan for the "Veritas" validation library.

## Phase 1: Core Engine Construction (v0.1)

**Goal**: Build a minimal, robust engine capable of evaluating CEL expressions and performing basic type and field validation.

-   **[x] 1.1: `Engine` Implementation**
    -   [x] 1.1.1: Define an `Engine` struct that encapsulates `cel.Env` and an LRU cache (`lru.Cache`) for compiled `cel.Program`s.
    -   [x] 1.1.2: Use `log/slog` for internal tracing. Log cache misses and program generations at the `Debug` level.
    -   [x] 1.1.3: Implement a framework to register custom CEL functions (e.g., `strings.ToUpper`, `matches`).

-   **[x] 1.2: `Validator` Implementation**
    -   [x] 1.2.1: Define a `Validator` struct holding a reference to the `Engine` and a `map[string]ValidationRuleSet` for rule lookup by type name.
    -   [x] 1.2.2: Implement the primary `Validate(obj any) error` method. Use `log/slog.Error` for internal errors like reflection failures.
    -   [x] 1.2.3: Implement logic to apply `TypeRules` and `FieldRules` separately, aggregating all validation failures using `errors.Join`.
    -   [x] 1.2.4: Ensure error messages are contextual, including the type and field name (e.g., `User.Email: validation failed...`).
    -   [x] 1.2.5: **[RESOLVED]** The `unsupported type` error in `cel-go` was bypassed by redesigning the API. Instead of attempting to register Go structs directly, the validator now requires a `TypeAdapter` function for each type. This function converts the struct to a `map[string]any`, which `cel-go` can handle reliably.

-   **[x] 1.3: Rule Provider Implementation**
    -   [x] 1.3.1: Define the `ValidationRuleSet` struct (containing `TypeRules`, `FieldRules`).
    -   [x] 1.3.2: Implement `JSONRuleProvider` to load rule sets from a JSON file. Log I/O or parsing errors with `slog.Error`.

-   **[x] 1.4: Unit Testing Foundation**
    -   [x] 1.4.1: Create a test suite for the `Validator` covering success, single failure, and multiple failure scenarios. The test suite has been updated to reflect the new `TypeAdapter`-based API.
    -   [x] 1.4.2: Use `errors.Is`, `errors.As`, and direct error string comparison for assertions, moving away from `go-cmp` for `error` types due to complexities with `errors.Join`.
    -   [x] 1.4.3: Adhere to the `want` and `got` variable naming convention for test comparisons.

## Phase 2: Static Analysis Tool Development (v0.2)

**Goal**: Develop the `veritas` CLI tool to automatically extract validation rules from Go source code.

-   **[x] 2.1: Go Source Code Parser**
    -   [x] 2.1.1: Implement a parser using `go/ast` to recursively scan directories and find `struct` definitions.

-   **[x] 2.2: Rule Extraction Logic**
    -   [x] 2.2.1: Extract field-level rules from `validate:"..."` struct tags.
    -   [x] 2.2.2: Extract type-level rules from special `// @cel: ...` comments preceding a `struct` definition.
    -   [x] 2.2.3: Implement a mapping from common shorthands (`required`, `nonzero`, `email`, etc.) to their corresponding CEL expressions using type-aware analysis.

-   **[x] 2.3: `veritas` CLI Implementation**
    -   [x] 2.3.1: Implement logic to output the extracted rules as a structured JSON file.
    -   [x] 2.3.2: Use `slog.Info` for progress reporting and `slog.Debug` for detailed parsing steps.

-   **[x] 2.4: Static Analysis Tool Testing**
    -   [x] 2.4.1: Prepare sample Go source files and their expected JSON output as test data.
    -   [x] 2.4.2: Write tests that run the generator and compare the actual output against the expected JSON using `go-cmp/cmp`.

## Phase 3: Advanced Data Structures Support (v0.3)

**Goal**: Enhance the library to handle complex data structures common in modern Go applications.

-   **[x] 3.1: Pointer and Nested Struct Handling**
    -   [x] 3.1.1: Implement recursive validation for nested structs.
    -   [x] 3.1.2: Implement logic to safely dereference and validate pointer fields (achieved by generating `!= nil` checks from `required` tag).

-   **[x] 3.2: Slice (`[]T`) Support**
    -   [x] 3.2.1: Support a `dive` keyword in the `validate` tag to apply rules to each element of a slice.
    -   [ ] 3.2.2: Include the array index in error messages (e.g., `User.Scores[2]: is invalid`).

-   **[x] 3.3: Map (`map[K]V`) Support**
    -   [x] 3.3.1: Support `keys` and `values` keywords to apply rules to a map's keys and values, respectively.
    -   [ ] 3.3.2: Include the map key in error messages (e.g., `User.Metadata['user_id']: is invalid`).

-   **[x] 3.4: Advanced Structure Testing**
    -   [x] 3.4.1: Expand the test suite to include complex structs with slices, maps, and nested pointers.

-   **[x] 3.5: Embedded Struct Support**
    -   [x] 3.5.1: Update parser to correctly handle embedded structs, inheriting validation rules from the embedded type.
    -   [x] 3.5.2: Ensure validator can correctly validate fields within embedded structs.

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

## Phase 5: Go Code Generation (v1.1 / v2.0)

**Goal**: Implement Go code generation as the primary, recommended method for rule management, improving type-safety, performance, and developer experience.

-   **[ ] 5.1: Enhance CLI for Go Code Generation**
    -   [ ] 5.1.1: Add a `--format=go` flag to the `veritas` CLI to enable Go source code output. The default will remain `--format=json` for backward compatibility in v1.x.
    -   [ ] 5.1.2: Implement the core logic to generate a Go source file (e.g., `generated_rules.go`) containing `veritas.ValidationRuleSet` definitions.
    -   [ ] 5.1.3: The generated file will use an `init()` function to register नियम sets into a global registry.

-   **[ ] 5.2: Implement Static Rule Provider**
    -   [ ] 5.2.1: Create a global rule registry within the `veritas` library that can be populated by the `init()` functions of generated code.
    -   [ ] 5.2.2: Update `veritas.NewValidator()` to be able to use this global registry by default, removing the need to pass a `RuleProvider` for the common use-case.
    -   [ ] 5.2.3: The existing `JSONRuleProvider` will be kept for users who need dynamic rule loading.

-   **[ ] 5.3: Update Documentation and Tooling**
    -   [ ] 5.3.1: Thoroughly document the new Go code generation workflow, including `go:generate` examples.
    -   [ ] 5.3.2: Update the main `README.md` and example projects to reflect Go code generation as the recommended approach.
    -   [ ] 5.3.3: Ensure tests cover the end-to-end code generation and validation process.