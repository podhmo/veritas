# Plan to Remove the Adapter Pattern

This document outlines the analysis of the current adapter pattern in `veritas` and the plan to remove it.

## 1. Analysis of the Current Adapter Pattern

The `veritas` library currently uses a `TypeAdapter` pattern to bridge the gap between Go structs and the `cel-go` library. This is necessary because `cel-go`, in its default configuration, does not understand how to access fields on a native Go struct. It primarily operates on `map[string]any`.

### How it Works

- **`TypeAdapterFunc`**: A function type `func(obj any) (map[string]any, error)` that takes a Go object (a struct) and converts it into a map.
- **`validator.adapters`**: The `Validator` stores a map of `reflect.Type` to `TypeAdapterTarget`. This allows it to find the correct adapter for a given Go struct type.
- **`NewValidatorFromJSONFile`**: When using JSON-based rules, the user is required to manually provide these adapters. This is a significant source of boilerplate and friction.
- **`go generate`**: The code generation approach *does not* use adapters. It registers rules globally, and the validation logic implicitly relies on the JSON-based validation path, which currently requires an adapter. This was a point of confusion in the initial analysis.

### The Problem with the Adapter Pattern

1.  **Boilerplate**: Users need to write and maintain `TypeAdapter` functions for each of their types when using the JSON validation method. This is tedious and error-prone.
2.  **Complexity**: The presence of adapters adds a layer of indirection to the validation logic, making the codebase harder to understand and maintain.
3.  **Performance**: The conversion from a struct to a map at runtime introduces a performance overhead that could be avoided.

## 2. The Solution: Native Go Structs in CEL with `ext.NativeTypes`

The initial assumption that `cel.Types()` could be used for arbitrary Go structs was incorrect. `cel.Types()` is primarily designed for Protobuf messages. The correct way to enable `cel-go` to work directly with native Go structs is by using the `ext.NativeTypes()` extension. This extension registers Go structs with the CEL type system, eliminating the need for the manual struct-to-map conversion and, therefore, the `TypeAdapter`.

### How `ext.NativeTypes` Works

The `ext.NativeTypes()` function, when passed to `cel.NewEnv()`, enables CEL to understand and access fields on Go structs directly. It requires the `reflect.TypeOf()` of the struct you want to register.

```go
import (
    "reflect"
    "github.com/google/cel-go/cel"
    "github.com/google/cel-go/ext"
)

// Example of registering a Go struct with CEL
env, err := cel.NewEnv(
    ext.NativeTypes(reflect.TypeOf(&my_package.MyStruct{})),
    cel.Variable("self", cel.ObjectType("my_package.MyStruct")),
)
```

By registering the Go struct types with the CEL environment using `ext.NativeTypes`, we can pass struct instances directly to the `program.Eval()` function. CEL will then be able to access the struct fields by name.

## 3. The Plan to Remove the Adapter

The user has requested that I only create this document. The following is a proposed plan for how a developer could implement this change.

1.  **Modify `NewValidator` to Accept Types**:
    -   Create a new `ValidatorOption` called `WithTypes(types ...any)`.
    -   This option will take a variadic list of Go struct instances (e.g., `User{}`, `Post{}`).
    -   Inside `NewValidator`, collect the `reflect.TypeOf()` for each provided struct instance.
    -   Use these types to create a new `cel.Env` with the `ext.NativeTypes()` option. This new environment will be used for all validations, replacing the adapter-based approach.

2.  **Deprecate and Remove `TypeAdapter`**:
    -   Remove the `TypeAdapterFunc`, `TypeAdapterTarget`, and the `adapters` map from the `Validator` struct.
    -   Remove the `WithTypeAdapters` option.
    -   This is a breaking change, and should be noted in the release.

3.  **Update the Validation Logic**:
    -   In `validator.go`, the `validateRecursive` function currently looks for an adapter and uses it to convert the object to a map.
    -   This logic needs to be changed. Instead of converting to a map, it should pass the raw struct object directly to the `prog.ContextEval()` call.
    -   The `objectVars` map will look like this: `map[string]any{"self": obj}` where `obj` is the user's struct, not a map.

4.  **Simplify `NewValidatorFromJSONFile`**:
    -   Change the signature of `NewValidatorFromJSONFile` to accept the new `WithTypes` option.
    -   Remove the now-defunct `WithTypeAdapters` option from its signature and documentation.

5.  **Update `go generate` Path (If Necessary)**:
    -   The `go generate` mechanism currently relies on the global registry. The `NewValidator()` function (with no options) uses this global registry.
    -   We need to ensure that when `NewValidator` is used, it can create an environment that can handle the types defined in the generated code. This might require a new mechanism to pass the types from the generated code to the validator. A simple approach would be to generate a function that returns the list of types, which the user can then pass to `NewValidator`.

6.  **Update Documentation and Examples**:
    -   Update `README.md` to reflect the new, simpler API.
    -   Update the `examples/http-server/main.go` to remove the `TypeAdapter` and use the new `WithTypes` option.
    -   Update all other relevant documentation.

By following this plan, we can completely remove the `TypeAdapter` pattern, resulting in a simpler, faster, and more user-friendly API for `veritas`.

## 4. Implementation Challenges and Future Work

During the implementation of the plan to remove the `TypeAdapter`, several challenges were encountered, revealing complexities in `cel-go`'s native type support, especially concerning generic types and environment management.

### Key Challenges

1.  **`overlapping identifier` Error**:
    -   **Problem**: When creating a single `cel.Env` for multiple types using `WithTypes`, repeatedly calling `cel.Variable("self", ...)` for each type resulted in an "overlapping identifier" error. The `self` variable was being redefined in the same environment.
    -   **Initial Attempt**: We tried to declare `self` as a `types.DynType` once. However, this caused `no such key` errors during evaluation because the environment lost the specific field information for each struct.
    -   **Solution**: The final approach was to create and cache a separate `cel.Env` for each `reflect.Type`. This ensures that each environment is configured with the correct type information for `self` without causing identifier conflicts. A `getNativeEnv(typ reflect.Type)` method was introduced to manage this cache.

2.  **`unsupported conversion to ref.Val` for Nil Pointers**:
    -   **Problem**: When a native struct field with a pointer type was `nil`, `cel-go`'s `ContextEval()` would return an `unsupported conversion` error. The adapter-based path handled this gracefully, but the native path did not.
    -   **Solution**: A `nil` check was added in the `validateNative` function before attempting to validate pointer fields. If a pointer is `nil`, validation on that field is skipped, with the assumption that a `required` or `nonzero` rule would catch it if the `nil` value is not allowed.

3.  **Inconsistent Rule Keys**:
    -   **Problem**: The `veritas-gen` tool generated fully qualified type names (e.g., `github.com/user/project/def.User`) as keys for the rule registry. However, the `getTypeName` method in the validator was generating a shorter, package-relative name (e.g., `def.User`). This mismatch caused rule lookup failures.
    -   **Solution**: The `getTypeName` method was updated to always generate the full package path, ensuring consistency between generated code and runtime lookup. This required updating handwritten rule files in test cases as well.

4.  **Generic Type Instantiation**:
    -   **Problem**: `cel-go`'s `ext.NativeTypes` requires a concrete `reflect.Type`. For a generic type like `Box[T]`, we need to provide a real instantiation, such as `Box[string]{}`. The `veritas-gen` tool was updated to generate a `GetKnownTypes() []any` function, but it currently uses `any` as a placeholder for generic type parameters (e.g., `Box[any]{}`). This is a simplification and may not cover all use cases.

### Future Work

The `TypeAdapter` pattern has not been fully removed. It remains the fallback for types not explicitly registered via `WithTypes`. The complete removal is postponed and tracked in `TODO.md` under a new, more detailed task. The primary focus of future work will be:

-   **Robust Generic Type Support**: Investigate a more robust way to handle generic type parameters in `veritas-gen` and the runtime validator to ensure correct `cel.Env` setup for any generic instantiation.
-   **Finalize API**: Once the native path is feature-complete and stable for all supported types (including generics, pointers, slices, and maps), the `TypeAdapter` and its related options can be fully deprecated and removed.
-   **Comprehensive Documentation**: Update all documentation to reflect the new `WithTypes`-first approach and provide clear guidance on handling complex and generic types.
