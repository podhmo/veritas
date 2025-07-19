# Advanced Topics

## Custom CEL Functions

Veritas provides a set of default custom functions that can be used in your validation rules.

### `strings.ToUpper(string) string`

Converts a string to uppercase.

**Example:**
```go
// @cel: strings.ToUpper(self.CountryCode) == "US"
type Address struct {
    CountryCode string `validate:"required"`
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
    import "github.com/podhmo/veritas"

    // Create a new engine with your function
    engine, err := veritas.NewEngine(isAwesome())
    if err != nil {
        // handle error
    }

    // Create a validator using the custom engine
    validator, err := veritas.NewValidatorWithEngine(engine)
    if err != nil {
        // handle error
    }

    // Now you can use the isAwesome function in your rules
    ```
