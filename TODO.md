# Veritas Development TODO List

This document outlines the detailed, phased development plan for the "Veritas" validation library.

## Phase 1: Core Engine Construction (v0.1)

**Goal**: Build a minimal, robust engine for CEL-based validation.

-   **[x] Core Components**: Implemented a `Validator` and `Engine` with CEL program caching.
-   **[x] Rule Loading**: Added support for loading validation rules from JSON files via a `JSONRuleProvider`.
-   **[x] Initial API Design**: **[Decision]** To bypass `cel-go`'s `unsupported type` error, the initial design used a `TypeAdapter` function to convert structs to `map[string]any`. (This was later superseded, see Phase 8).

## Phase 2: Static Analysis Tool Development (v0.2 & v0.2.1)

**Goal**: Develop the `veritas` CLI for rule extraction and linting.

-   **[x] Rule Extraction**: The CLI can parse Go source code (`go/ast`) to extract validation rules from `validate` struct tags and `// @cel:` comments.
-   **[x] Shorthand Support**: Implemented a mapping from shorthands (`required`, `email`, etc.) to CEL expressions.
-   **[x] Linter**: Added a `-lint` flag to the CLI to statically check for syntactically valid CEL and correct field names in rules.

## Phase 3: Advanced Data Structures Support (v0.3)

**Goal**: Enhance the library to handle complex data structures.

-   **[x] Nested Structures**: Implemented recursive validation for nested and embedded structs and pointers.
-   **[x] Slices and Maps**: Added support for `dive`, `keys`, and `values` keywords in `validate` tags to apply rules to collection elements.
-   **[x] Error Reporting Note**: **[Limitation]** While validation is recursive for collections, error messages for primitives do not include the specific index or key that failed due to the use of `cel.all()`.

## Phase 4: GA Finalization (v1.0)

**Goal**: Polish the library for a v1.0 release.

-   **[x] Go Generics**: The parser and validator can handle generic struct definitions.
    -   **[Note]** Pointer values within generic types are handled, but full native generic support in CEL is still a challenge.
-   **[x] API Hardening**: Added `context.Context` to `Validate` for cancellation.
-   **[x] Documentation**: Created a documentation website and a `net/http` example project.
-   **[x] Key Decisions**:
    -   **[Decision]** Regular expressions must be RE2-compatible, as CEL does not support PCRE features like lookaheads.
    -   **[Decision]** The CLI uses `pkg.PkgPath` for fully qualified type names to prevent conflicts.

## Phase 5: Go Code Generation (v1.1 / v2.0)

**Goal**: Implement Go code generation as the primary, recommended method for rule management.

-   **[x] Go Code Generation**: The `veritas` CLI can generate a Go source file (`--format=go`) containing all validation rules.
-   **[x] Global Registry**: The generated code uses an `init()` function to populate a global rule registry, simplifying validator setup (`veritas.NewValidator()`).
-   **[x] Tooling**: The generator was refactored to use `singlegenerator` for improved structure and maintainability.

## Phase 6: Additional Tasks

-   **[x] `-o` Flag**: The `veritas` CLI accepts an `-o` flag to specify the output file path.
-   **[x] Example Fixes**: Resolved an issue in the `gencode` example by using `TypeAdapterTarget` for correct rule mapping.

## Phase 7: Refactoring and Performance Tuning

- [x] Refactor the analyzer for performance based on `docs/analyzer-tuning.md`.
  - This involved creating `parser.PackageInfo` to avoid redundant package loading by `veritas-gen`, updating the parser to use it, while keeping the original `Parse` method for backward compatibility.

## Phase 8: Remove Adapter Pattern (Completed)

**Goal**: Removed the `TypeAdapter` pattern, simplifying the API and improving performance by using `cel-go`'s native struct support.

-   [x] **Summary**: Introduced a `WithTypes(...)` option to `NewValidator` to register Go structs directly with the `cel-go` environment. This is now the recommended approach. The `veritas-gen` tool was updated to generate a `GetKnownTypes()` function, making it easy to register all relevant types. The documentation and examples (`http-server`, `gencode`) have been updated to reflect this new, simpler API.
-   **[NOTE]** The `TypeAdapter` pattern is retained as a legacy fallback for complex generic type scenarios where `cel-go`'s native support has limitations. Full native support for generics remains a challenge due to the static nature of CEL's type checking. Rules for generic types must be written to be compatible with all possible type instantiations.

## Phase 9: One-File Code Injection

**Goal**: Add the ability to inject validation logic directly into a single Go file.

-   **[x] `-inject` Flag**: The `veritas` CLI accepts an `-inject=<filename>` flag to enable this mode.
-   **[x] `setupValidation` Function**: The tool finds a function named `setupValidation` in the target file and replaces its body. If the function doesn't exist, it's appended. An `init()` function is also added to call `setupValidation()`.
-   **[x] Documentation**: The `README.md` was updated to reflect this new feature.

## Phase 10: Chicken and Egg Problem Resolution

**Goal**: Resolve the "chicken and egg" problem that occurs during code generation with `-inject`.

- [x] **10.1: Resolve `undefined: GetKnownTypes` error**
  - Implemented a workaround to ignore "undefined: GetKnownTypes" errors during the initial parsing phase of code generation. This allows the process to complete successfully even when the generated functions don't exist yet. A warning is logged to indicate that the error was ignored.

## Phase 11: Parser Enhancement

**Goal**: Improve the capabilities of the `validate` tag parser.

- [ ] Add support for `min=<value>` and `max=<value>` shorthands for numeric types (e.g., `validate:"min=18"`).
- [ ] Add support for `len=<value>`, `min=<value>`, `max=<value>` for string and slice types (e.g., `validate:"min=2,max=20"` for a string).
