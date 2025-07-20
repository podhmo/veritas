# Advanced Topics

## Custom CEL Functions

Veritas provides a set of default custom functions that can be used in your validation rules.

### `strings.ToUpper(string) string`

Converts a string to uppercase.

**Example:**
```go
// @cel: strings.ToUpper(self.CountryCode) == "US"
type Address struct {
    CountryCode string `validate:"nonzero"`
}
```

### `custom.matches(string, string) bool`

Performs a regular expression match. This provides an alternative to the built-in `matches` function and uses Go's `regexp` engine.

**Example:**
```go
type User struct {
    Username string `validate:"cel:custom.matches(self, '^[a-z0-9_]+$')"`
}
```

## Adding Your Own Custom Functions

You can extend Veritas with your own custom functions by creating a `cel.EnvOption`.

1.  **Define your function:**

    ```go
    import (
        "github.com/google/cel-go/cel"
        "github.com/google/cel-go/common/types"
        "github.com/google/cel-go/common/types/ref"
    )

    func isAwesome() cel.EnvOption {
        return cel.Function("isAwesome",
            cel.Overload("is_awesome_string",
                []*cel.Type{cel.StringType},
                cel.BoolType,
                cel.UnaryBinding(func(s ref.Val) ref.Val {
                    return types.Bool(s.(types.String) == "veritas")
                }),
            ),
        )
    }
    ```

2.  **Create your validator with the custom function:**

    ```go
    import (
        "github.com/podhmo/veritas"
        "your-project/models" // Assuming your types are in this package
    )


    // Create a validator, passing the custom function
    // and registering your application's types.
    validator, err := veritas.NewValidator(
        veritas.WithCELHelpers(isAwesome()),
        veritas.WithTypes(models.GetKnownTypes()...),
    )
    if err != nil {
        // handle error
    }

    // Now you can use the isAwesome function in your rules
    ```

## Legacy Patterns: The `TypeAdapter`

Previous versions of Veritas used a `TypeAdapter` pattern to convert Go structs into a `map[string]any` before validation. This was necessary to work around limitations in `cel-go`'s native type support.

**This pattern is now considered legacy.**

The recommended approach is to use `veritas.WithTypes(...)` to register your Go structs directly with the validation engine. This provides better performance and a simpler API.

The `TypeAdapter` is preserved for backward compatibility and for complex scenarios involving generic types where the native `cel-go` support may still have limitations. For most use cases, you should prefer `WithTypes`.
