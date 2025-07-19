# Rules Reference

Veritas uses struct tags to define validation rules. The `validate` tag contains a comma-separated list of rules to be applied to a field.

```go
type User struct {
    Name  string `validate:"nonzero,cel:self.size() < 50"`
    Email string `validate:"nonzero,email"`
    Age   int    `validate:"cel:self >= 18"`
}
```

## Type-Level Rules

You can also define rules that apply to the entire struct using a special `// @cel:` comment. This is useful for validation that involves multiple fields.

```go
// @cel: self.Password == self.PasswordConfirm
type User struct {
    Password        string `validate:"nonzero"`
    PasswordConfirm string `validate:"nonzero"`
}
```

## Field-Level Rules

### Shorthands

Veritas provides several shorthands for common validation scenarios.

| Shorthand  | Description                                                                                             | Applicable Types                        |
| :--------- | :------------------------------------------------------------------------------------------------------ | :-------------------------------------- |
| `required` | Asserts that a pointer is not `nil`.                                                                       | pointers                                |
| `nonzero`  | Asserts that a value is not its zero value (e.g., not `""`, `0`, `false`, `nil`, or an empty slice/map). | `string`, numeric types, pointers, bool, slices, maps |
| `email`    | The string must match a basic email format.                                                             | `string`                                |

### Raw CEL Expressions

For more complex validation, you can use a raw CEL expression with the `cel:` prefix. The field's value is available as the `self` variable.

```go
type Product struct {
    // self refers to the Price field
    Price float64 `validate:"cel:self > 0.0 && self < 1000.0"`

    // self refers to the SKU field
    SKU   string  `validate:"cel:self.startsWith('PROD-')"`
}
```

### Collection Rules

To validate the elements of slices or maps, you can use the `dive`, `keys`, and `values` keywords.

| Keyword  | Description                                        | Example                                               |
| :------- | :------------------------------------------------- | :---------------------------------------------------- |
| `dive`   | Applies rules to each element of a slice.          | `validate:"dive,nonzero"` (each element must not be its zero value) |
| `keys`   | Applies rules to each key of a map.                | `validate:"keys,cel:self.size() > 3"` (each key > 3 chars) |
| `values` | Applies rules to each value of a map.              | `validate:"values,nonzero"` (each value must be non-zero) |

These can be combined with other rules. For example, to validate a slice of emails where each email must also be under 64 characters:

```go
type UserList struct {
    Emails []string `validate:"dive,email,cel:self.size() < 64"`
}
```
