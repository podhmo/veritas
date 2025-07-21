# Veritas Knowledge Base

This document consolidates key architectural decisions, troubleshooting history, and development principles for the Veritas library.

## 1. Native Struct Validation with `WithTypes`

### The Goal: Adapter-Free Validation

The primary goal of Veritas is to provide a seamless validation experience for native Go structs without requiring excessive boilerplate. Initially, this was attempted by dynamically configuring the `cel-go` environment for any given struct, which led to a series of challenges.

### The Problem: `unsupported type` and Environment Management

The main obstacle was a persistent `unsupported type` error from `cel-go` when trying to register arbitrary Go structs using `cel.Types()`. Extensive investigation revealed that this approach was flawed. The correct mechanism is `ext.NativeTypes`, but its usage introduced new complexities:

1.  **Environment Conflicts**: Creating a single `cel.Env` for multiple types by repeatedly calling `cel.Variable("self", ...)` caused an "overlapping identifier" error. `cel-go` environments are not designed to have the same variable (`self`) represent multiple, distinct struct types simultaneously.
2.  **Type Information Loss**: Declaring `self` as a generic `types.DynType` to avoid conflicts resulted in the loss of specific field information, leading to `no such key` errors during rule evaluation.
3.  **Nil Pointer Conversions**: The native validation path in `cel-go` struggled with `nil` pointer fields, throwing `unsupported conversion to ref.Val` errors where the adapter-based path had previously worked.

### The Solution: Type-Specific Environments and `WithTypes`

The robust solution was to pivot from creating a single, universal environment to creating and caching a specific `cel.Env` for each `reflect.Type` being validated.

-   **`WithTypes` Option**: This is now the primary, recommended way to enable native struct validation. The user provides a list of struct instances (e.g., `veritas.WithTypes(User{}, Post{})`).
-   **Dynamic Environment Caching**: Internally, the `Validator` maintains a cache of `cel.Env` instances. When validating a struct of a certain type for the first time, it creates a dedicated environment for that type using `ext.NativeTypes` and caches it. This resolves the "overlapping identifier" issue and ensures full type information is available for validation.
-   **Nil Pointer Handling**: The native validation logic now includes a check to skip `nil` pointer fields, preventing the `unsupported conversion` error. It is assumed that `nil` fields, if disallowed, will be caught by a `required` or `nonzero` rule.

### Current Status and Future Work

The `WithTypes` option is the recommended approach for native struct validation.

The legacy `TypeAdapter` pattern has not been fully removed. It is retained as a fallback mechanism for two main reasons:
1.  **Complex Scenarios**: It can still be useful in situations where `WithTypes` is not a good fit.
2.  **Unsupported Types**: It provides a path for handling types that are not yet fully supported by the native validation path, such as **generic types**.

The complete removal of the `TypeAdapter` is postponed until the native validation path achieves feature parity, particularly with robust support for generics.

## 2. Development Workflow Principles

### Documentation-First Approach for Complex Tasks

When undertaking complex or high-risk tasks, such as major refactoring or implementing features with unclear paths, it is crucial to prioritize documentation.

**Procedure:**
1.  **Initial Analysis and Plan**: Before writing code, document the problem, the proposed solution, and a step-by-step implementation plan in a dedicated markdown file (e.g., `docs/feature-plan.md`).
2.  **Record Progress and Setbacks**: As work progresses, update this document with findings, especially any setbacks, failed attempts, or changes in direction.
3.  **Prioritize Documentation Commits**: If a task is paused or completed, the first action should be to commit the updated documentation. This ensures that valuable knowledge and context are preserved, even if the associated code changes are discarded or postponed.

**Rationale**:
This practice ensures that the "why" behind a decision is never lost. If a developer (including our AI agent, Jules) has to revisit the task later, the document provides a clear record of what was tried and why certain paths were abandoned. This prevents repeating failed experiments and provides a solid foundation for future efforts. For example, the detailed log of attempts to remove the `TypeAdapter` pattern in `docs/remove-adapter-plan.md` is a critical asset for future development.

## 3. Resolving the "Chicken and Egg" Problem in Code Generation

### The Goal: Robust `go:generate` Workflow

A key feature of `veritas` is its ability to automatically generate validation logic from source code using a `//go:generate` directive. A seamless experience requires this process to be robust, even when the code being analyzed is incomplete because it depends on the code that is about to be generated.

### The Problem: `undefined: GetKnownTypes`

This "chicken and egg" problem manifested when running `go generate` in a file that called `GetKnownTypes()`. The `veritas` generator would analyze the source code, find a call to a function (`GetKnownTypes`) that did not yet exist, and fail with a type-checking error (`undefined: GetKnownTypes`). This halted the entire generation process, preventing the very file that would define the function from being created.

### Investigation and Abandoned Solutions

1.  **`go/packages` Overlay**: The initial thought was to use the `overlay` feature of the `go/packages` API. This feature allows providing virtual, in-memory versions of files to the loader. The idea was to create a fake, empty `GetKnownTypes` function in an overlay to satisfy the type checker during the initial analysis.
2.  **The Roadblock**: This approach was quickly abandoned because the higher-level analysis frameworks being used—`golang.org/x/tools/go/analysis` and its runners like `singlegenerator` and `multichecker`—do not expose a configuration option to pass an `overlay` down to the underlying `go/packages` loader. Modifying these frameworks was deemed out of scope.

### The Solution: Targeted Error Suppression

The adopted solution is a pragmatic workaround that directly addresses the symptom:

-   **Error Inspection**: The generator's core analysis function was modified to inspect the error returned by the package loader.
-   **Conditional Suppression**: If the error message contains the string `"undefined:"`, the generator assumes it has encountered the chicken-and-egg problem. It logs a warning message to inform the user but otherwise suppresses the error and continues execution. In this state, no validation rules or types are discovered, which is acceptable because the primary goal is to allow the generator to proceed to the code-writing phase.
-   **Successful Second Pass**: On the *next* run of `go generate`, the file created in the previous run (containing `GetKnownTypes`) now exists, the type checker is satisfied, and the generator can successfully parse the rules and regenerate the file with the correct content.

This solution, while not as elegant as an overlay, is effective, self-contained within the `veritas` codebase, and provides a robust user experience for the `go:generate` workflow.
