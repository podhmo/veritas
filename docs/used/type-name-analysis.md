# Analysis of Type Name Handling in Validator

## 1. The Problem

The current implementation of the `Validator` has an issue with how it handles `TypeAdapter`s when the type of the object being validated is different from the type for which the validation rules are registered.

Specifically, the `Validator` uses the `reflect.Type` of the input object to derive a type name (e.g., `main.User`), and then uses this name to look up both `TypeAdapter`s and `ValidationRuleSet`s.

This creates a problem in scenarios like the `gencode` example, where:

- Validation rules are defined for `def.User`.
- The code to be validated uses a different, local `main.User` struct.
- A `TypeAdapter` is provided to convert `main.User` to the structure expected by `def.User`'s rules.

The current implementation fails because when `v.Validate(ctx, user)` is called with a `main.User` object, the validator:
1. Derives the type name `main.User`.
2. Looks for a `TypeAdapter` with the key `main.User` and doesn't find one (it's registered as `def.User`).
3. Fails with a "no TypeAdapter registered" error.

It never gets to the point of using the provided adapter because the key doesn't match.

### Code Example (Current State)

**`examples/gencode/main.go`:**
```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/podhmo/veritas"
	_ "github.com/podhmo/veritas/examples/gencode/def"
)

type User struct {
	Name  string
	Email string
}

func main() {
	// ...
	v, err := veritas.NewValidator(
		veritas.WithTypeAdapters(
			map[string]veritas.TypeAdapter{
				"def.User": func(ob any) (map[string]any, error) { // KEY is "def.User"
					v, ok := ob.(User) // INPUT is main.User
					if !ok {
						return nil, fmt.Errorf("unexpected type %T", ob)
					}
					return map[string]any{
						"Name":  v.Name,
						"Email": v.Email,
					}, nil
				},
			},
		),
	)
	// ...
	user := User{Name: "foo", Email: "foo@example.com"}
	if err := v.Validate(ctx, user); err != nil { // user is of type main.User
		return fmt.Errorf("validation failed, unexpectedly: %+v", err)
	}
}
```

## 2. Desired Behavior

The validator should be able to map a source type (`main.User`) to a target type's validation rules (`def.User`) using an adapter.

The developer should be able to configure the validator like this:
"When you encounter an object of type `main.User`, use this specific adapter function, and then apply the validation rules registered for `def.User`."

## 3. Proposed Solution

To achieve the desired behavior, I propose changing the `TypeAdapter` registration mechanism.

### 3.1. New `TypeAdapter` and `TypeAdapterTarget` structs

Instead of `map[string]TypeAdapter`, we will introduce a new structure that explicitly defines the mapping from a source type to a target type's rules.

```go
// in validator.go

// TypeAdapterFunc is the function signature for converting a Go object.
type TypeAdapterFunc func(obj any) (map[string]any, error)

// TypeAdapterTarget specifies the target rule set for a type adapter.
type TypeAdapterTarget struct {
    TargetName string
    Adapter    TypeAdapterFunc
}
```

### 3.2. Modified `WithTypeAdapters`

The `WithTypeAdapters` option will now accept a map of the new structure: `map[reflect.Type]TypeAdapterTarget`. Using `reflect.Type` as the key is more robust and type-safe than using strings.

```go
// in validator.go

// WithTypeAdapters sets the type adapters for the validator.
func WithTypeAdapters(adapters map[reflect.Type]TypeAdapterTarget) ValidatorOption {
	return func(o *validatorOptions) {
		for k, v := range adapters {
			o.adapters[k] = v
		}
	}
}
```

The `validatorOptions` and `Validator` structs will be updated to store this new map.

### 3.3. Updated `Validator.Validate` Logic

The `Validator.Validate` and `Validator.validateRecursive` methods will be updated to use this new mapping.

The new logic will be:
1. Get the `reflect.Type` of the input object.
2. Look up this `reflect.Type` in the `v.adapters` map.
3. If a `TypeAdapterTarget` is found:
    a. Use its `Adapter` function to convert the object to a `map[string]any`.
    b. Use its `TargetName` string to look up the `ValidationRuleSet`.
    c. Apply the rules.
4. If no adapter is found, proceed with the existing logic (which is useful for nested objects that are already in the correct domain).

### 3.4. Example of New Usage

**`examples/gencode/main.go` (updated):**
```go
package main

// ... imports
import "reflect"

// ... User struct

func main() {
	// ...
	v, err := veritas.NewValidator(
		veritas.WithTypeAdapters(
			map[reflect.Type]veritas.TypeAdapterTarget{
				reflect.TypeOf(User{}): { // KEY is reflect.Type
					TargetName: "github.com/podhmo/veritas/examples/gencode/def.User", // TARGET rule set
					Adapter: func(ob any) (map[string]any, error) {
						v, ok := ob.(User)
						if !ok {
							return nil, fmt.Errorf("unexpected type %T", ob)
						}
						return map[string]any{
							"Name":  v.Name,
							"Email": v.Email,
						}, nil
					},
				},
			},
		),
	)
	// ...
}
```

This new approach is more explicit, type-safe, and directly solves the problem of mapping between different but structurally-compatible types.

## 4. Record of Attempts

This section will be updated with the results of implementing the proposed solution.

### Attempt 1: Implementing the Proposed Solution

- **Action**: Modified `validator.go` to introduce `TypeAdapterFunc` and `TypeAdapterTarget`. Changed `WithTypeAdapters` to accept `map[reflect.Type]TypeAdapterTarget`.
- **Result**: This seems promising. The code is more explicit.

- **Action**: Updated `Validator.Validate` to use the new adapter map.
- **Result**: The logic is more complex, but it correctly separates the concerns of "what type am I?" from "what rules should I use?".

- **Action**: Updated `examples/gencode/main.go` to use the new API.
- **Result**: The example code is now more verbose but also much clearer about its intent.

- **Action**: Ran `go test ./...`.
- **Result**: Some tests failed because they were still using the old `WithTypeAdapters` API. I will need to update them.

- **Action**: Updated `validator_test.go` to use the new API and added a specific test case for the type-mapping scenario.
- **Result**: All tests now pass, including the new test case. The `gencode` example also works as expected.

### Conclusion

The proposed solution was successful. The key was to decouple the type of the object being validated from the name of the rule set to apply. The new `TypeAdapterTarget` struct and the use of `reflect.Type` as the map key provide a robust and explicit way to configure this mapping.
