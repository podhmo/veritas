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

## Project Name

The name "Veritas" is Latin for "truth." It was chosen to symbolize the library's core function: to verify the "truthfulness" or correctness of data against a defined set of rules.

## Installation

```bash
go get github.com/podhmo/veritas
```

## Quick Start

The recommended approach is to use Go code generation for a type-safe, high-performance validation experience.

1.  **Annotate your structs:**

    Add validation rules using struct tags and special `// @cel:` comments.

    ```go
    // file: models/user.go
    package models

    //go:generate go run github.com/podhmo/veritas/cmd/veritas -gen-type=User -out-name=veritas_gen.go

    // @cel: self.Password == self.PasswordConfirm
    type User struct {
        Name     string `validate:"required,cel:self.size() < 50"`
        Email    string `validate:"required,email"`
        Age      int    `validate:"cel:self >= 18"`
        Password string `validate:"required,cel:self.size() >= 10"`
        PasswordConfirm string `validate:"required"`
    }
    ```

2.  **Generate validation code:**

    Use `go generate` to run the `veritas` tool. It will scan your struct and create a `veritas_gen.go` file in the same package.

    ```bash
    go generate ./...
    ```

    The generated file will contain an `init()` function that automatically registers your validation rules with the veritas library.

3.  **Use the validator in your application:**

    The validator can now be created without any special configuration. The rules are available globally.

    ```go
    package main

    import (
        "context"
        "fmt"
        "log"

        "github.com/podhmo/veritas"
        "your-project/models" // Import the package with your models
    )

    func main() {
        // Create a new validator.
        // Rules are automatically loaded from the generated code.
        validator, err := veritas.NewValidator()
        if err != nil {
            log.Fatalf("Failed to create validator: %v", err)
        }

        // Create a user object to validate
        user := models.User{
            Name:            "Test User",
            Email:           "test@example.com",
            Age:             30,
            Password:        "password123",
            PasswordConfirm: "password123",
        }

        // Validate the object
        if err := validator.Validate(context.Background(), user); err != nil {
            fmt.Printf("Validation failed: %v\n", err)
        } else {
            fmt.Println("Validation successful!")
        }
    }
    ```

For more advanced use cases, such as loading rules from JSON files for dynamic environments, please refer to the documentation in the `docs` directory.