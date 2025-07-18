# Veritas

Veritas is a dynamic, type-safe, and extensible validation library for Go, powered by Google's Common Expression Language (CEL).

It aims to provide a robust validation framework where rules are derived directly from your Go source code as the "Single Source of Truth." This is achieved by a companion CLI tool that performs static analysis on your code, extracting rules from struct tags and special comments.

## Key Features

- **Declarative Rules in Go Code**: Define validation rules directly in your structs using tags (`validate:"..."`) and special comments (`// @cel:`).
- **Static Analysis**: A CLI tool (`veritas`) scans your code and generates a JSON representation of your validation rules, ensuring rules and code are always in sync.
- **Dynamic Execution**: The library loads the generated JSON rules at runtime to perform validation, powered by the high-performance `cel-go` engine.
- **Extensible**: Add your own custom validation functions to the CEL environment.
- **Type-Safe**: Designed to handle complex, nested data structures, including pointers, slices, and maps, with type-safety in mind.
- **Modern Go**: Built with Go 1.24, `log/slog` for structured logging, and `go-cmp` for testing.

## Installation

```bash
go get github.com/podhmo/veritas
```

## Quick Start (Conceptual)

1.  **Annotate your structs:**

    ```go
    package user

    // @cel: self.Password == self.PasswordConfirm
    type User struct {
        Name     string `validate:"required,cel:self.size() < 50"`
        Email    string `validate:"required,email"`
        Age      int    `validate:"cel:self >= 18"`
        Password string `validate:"required,cel:self.size() >= 10"`
        PasswordConfirm string `validate:"required"`
    }
    ```

2.  **Generate rules:**

    ```bash
    go run github.com/podhmo/veritas/cmd/veritas -in ./... -out rules.json
    ```

3.  **Use the validator in your application:**

    ```go
    package main

    import (
        "fmt"
        "github.com/podhmo/veritas"
    )

    func main() {
        // Load rules from the generated JSON file
        provider, err := veritas.NewJSONRuleProvider("rules.json")
        if err != nil {
            // ... handle error
        }

        // Create a new validator
        validator, err := veritas.NewValidator(provider)
        if err != nil {
            // ... handle error
        }

        // Validate an object
        invalidUser := user.User{...}
        if err := validator.Validate(invalidUser); err != nil {
            fmt.Println(err)
        }
    }
    ```