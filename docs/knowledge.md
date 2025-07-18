# Investigation into `cel-go`'s `unsupported type` Error

This document details the extensive troubleshooting process undertaken to resolve the `unsupported type: <native Go struct>` error when attempting to dynamically register and validate native Go structs with `cel-go`.

## The Problem

The primary goal was to create a validation library that could dynamically validate any given Go struct without requiring manual, boilerplate registration code from the user. The core of this functionality relied on `cel-go`'s ability to understand and operate on native Go types.

The issue manifested as a persistent `unsupported type` error originating from `cel.NewEnv` or `cel.Env.Extend` whenever `cel.Types()` was used with a native Go struct type.

## Approaches Attempted

### 1. Dynamic Environment Caching per Type

-   **Hypothesis:** The `cel.Env` should be immutable after creation. Perhaps creating a new, extended environment for each type during validation would work.
-   **Implementation:**
    -   The `Engine` was designed to cache `cel.Env` instances on a per-type basis.
    -   When `Validator.Validate(obj)` was called, it would request an environment for `typeof(obj)`.
    -   If not cached, a new environment was created by calling `baseEnv.Extend(cel.Types(obj), ...)`
-   **Result:** **Failure.** The `unsupported type` error still occurred within `env.Extend`. This indicated that `cel.Types()` was the source of the problem, regardless of when it was called.

### 2. Pre-registration at Validator Creation

-   **Hypothesis:** `cel.Types()` might only be effective during the initial `cel.NewEnv` call, not in `env.Extend`. The principle of immutability suggested that type information must be known at the very beginning.
-   **Implementation:**
    -   The `Validator` API was changed. `NewValidator` now accepted a list of sample struct instances (`typesToRegister ...any`).
    -   Inside `NewValidator`, a single `cel.Env` was created for that validator instance by calling `cel.NewEnv(..., cel.Types(type1), cel.Types(type2), ...)`
-   **Result:** **Failure.** The exact same `unsupported type` error occurred, this time during the `cel.NewEnv` call. This was a critical finding, as it invalidated the hypothesis that the timing of the call (`NewEnv` vs `Extend`) was the issue.

### 3. Exploring `cel-go` Environment Options

-   **Hypothesis:** Perhaps a missing environment option was needed to enable native type reflection.
-   **Implementation:**
    -   Added `cel.HomogeneousAggregateLiterals()`: This is for list/map literals, but it was worth a try. No effect.
    -   Attempted to use `ext.NativeTypes()`: This seemed promising. However, its API requires passing the types to be registered, which is the same as `cel.Types`, and it resulted in a different set of "unsupported type" errors related to its arguments (`unsupported native type: true (bool)`). This felt like a dead end.

### 4. `cel.Lib` Implementation

-   **Hypothesis:** The idiomatic way to provide custom types and functions to `cel-go` is by implementing the `cel.Library` interface. This encapsulates type registration logic.
-   **Implementation:**
    -   The `Validator` was refactored to implement `cel.Library`.
    -   The `CompileOptions()` method returned a slice of `cel.EnvOption`, including `cel.Types(...)` for all registered types.
    -   `NewValidator` called `engine.env.Extend(cel.Lib(validator))`.
-   **Result:** **Failure.** The `unsupported type` error persisted. This was the most surprising failure, as `cel.Lib` is the primary extension mechanism for `cel-go`. The error still originated from the underlying `cel.Types()` call within the library's `CompileOptions`.

### 5. Manual Field Registration via Reflection (Attempted)

-   **Hypothesis:** If `cel.Types()` is fundamentally broken or unsuitable for this use case, perhaps we can bypass it by manually describing the struct to `cel-go`.
-   **Implementation Idea:**
    -   Use Go's `reflect` package to iterate over the fields of a struct.
    -   For each field, map its Go type (e.g., `string`, `int`) to a CEL type (`cel.StringType`, `cel.IntType`).
    -   Construct a `cel.TypeDescr` with `cel.NewObjectType` and a list of `cel.Field` definitions.
-   **Result:** **Aborted.** This approach proved to be extremely complex. It would require a comprehensive mapping of Go types to CEL types and correctly implementing the `ref.TypeProvider` interface, which is a non-trivial task involving deep `cel-go` internal APIs. The risk of introducing subtle bugs was too high.

## Conclusion

After exhausting all reasonable avenues, the conclusion is that `cel.Types()` does not function as expected for dynamically registering arbitrary Go structs in the context of this project. The root cause is likely a subtle interaction within `cel-go`'s type system, a version-specific bug, or a misunderstanding of a core, undocumented constraint.

Since the primary goal is to deliver a functional validator, clinging to a broken mechanism is counterproductive. The only reliable path forward is to **shift the responsibility of type adaptation to the user**. By requiring the user to provide a `TypeAdapter` function, we work *with* `cel-go`'s type system instead of fighting against it. This sacrifices the "magic" of automatic registration but guarantees correctness and stability.
