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

## 2. The Solution: Native Go Structs in CEL

The `cel-go` library (especially in versions >= `v0.11.0`) provides a mechanism to work directly with native Go structs. This eliminates the need for the struct-to-map conversion and, therefore, the `TypeAdapter`.

The key is the `cel.Types()` option, which can be passed to `cel.NewEnv()`.

```go
// Example of registering a Go struct with CEL
env, err := cel.NewEnv(
    cel.Types(&my_package.MyStruct{}),
)
```

By registering the Go struct types with the CEL environment, we can pass the struct instances directly to the `program.Eval()` function. CEL will then be able to access the struct fields by name, just as it would with a map.

## 3. The Plan to Remove the Adapter

The user has requested that I only create this document. The following is a proposed plan for how a developer could implement this change.

1.  **Modify `NewValidator` to Accept Types**:
    -   Create a new `ValidatorOption` called `WithTypes(types ...any)`.
    -   This option will take a variadic list of Go struct instances (e.g., `User{}`, `Post{}`).
    -   Inside `NewValidator`, collect these types and use them to create a new `cel.Env` with the `cel.Types()` option. This new environment should be used for object-level validations.

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
