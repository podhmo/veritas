# CEL Cheatsheet

This page provides a quick reference for common expressions and functions available in the Common Expression Language (CEL). For a complete guide, please refer to the official [CEL language definition](https://github.com/google/cel-spec/blob/master/doc/langdef.md).

## Basic Operators

| Operator | Description        | Example                             |
| -------- | ------------------ | ----------------------------------- |
| `==`     | Equality           | `self.Status == "active"`           |
| `!=`     | Inequality         | `self.Role != "admin"`              |
| `>`      | Greater than       | `self.Age > 18`                     |
| `>=`     | Greater or equal   | `self.Price >= 0`                   |
| `<`      | Less than          | `self.Attempts < 5`                 |
| `<=`     | Less or equal      | `self.Score <= 100`                 |
| `&&`     | Logical AND        | `self.Enabled && self.Visible`      |
| `||`     | Logical OR         | `self.IsAdmin || self.IsOwner`      |
| `!`      | Logical NOT        | `!self.IsExpired`                   |
| `in`     | Membership         | `self.Role in ["editor", "viewer"]` |

## String Functions

| Function      | Description                               | Example                                  |
| ------------- | ----------------------------------------- | ---------------------------------------- |
| `size()`      | Returns the length of the string.         | `self.size() < 20`                       |
| `startsWith()`| Checks if a string starts with a prefix.  | `self.startsWith("user_")`               |
| `endsWith()`  | Checks if a string ends with a suffix.    | `self.endsWith(".jpg")`                  |
| `contains()`  | Checks if a string contains a substring.  | `self.contains("@")`                     |
| `matches()`   | Checks if a string matches a regex.       | `self.matches("^[a-zA-Z0-9]+$")`         |

## List Functions

| Function | Description                               | Example                               |
| -------- | ----------------------------------------- | ------------------------------------- |
| `size()` | Returns the number of elements in a list. | `self.size() > 0`                     |
| `[]`     | Access an element by index.               | `self[0] == "first"`                  |
| `in`     | Check for presence of an element.         | `"admin" in self.Roles`               |

## Map Functions

| Function | Description                        | Example                           |
| -------- | ---------------------------------- | --------------------------------- |
| `size()` | Returns the number of keys in a map. | `size(self.Attributes) > 0`       |
| `in`     | Check for presence of a key.       | `"id" in self.Metadata`           |
| `[]`     | Access a value by key.             | `self.Metadata["version"] == "v1"`|

## Special Keywords

- `self`: Refers to the field or object being validated.

## Type-specific Examples

### Structs

When validating a struct, `self` refers to the struct instance. You can access its fields using dot notation.

```go
// @cel: self.Password == self.PasswordConfirm
type User struct {
    Password        string `validate:"nonzero"`
    PasswordConfirm string `validate:"nonzero"`
}
```

### Fields

When using `validate` tags, `self` refers to the field's value.

```go
type Product struct {
    Name  string `validate:"nonzero,cel:self.size() < 50"`
    Price int    `validate:"cel:self > 0"`
}
```
