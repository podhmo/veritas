>[!NOTE]
>This library was created entirely with Jules (AI agent) for experimentation purposes.

# Veritas

Veritas is a dynamic, type-safe, and extensible validation library for Go, powered by Google's Common Expression Language (CEL).

## Features

- **Declarative Rules in Go Code**: Define validation rules directly in your structs using tags (`validate:"..."`) and special comments (`// @cel:`).
- **Two Modes of Operation**:
    - **Code Generation**: A CLI tool (`veritas`) scans your code and generates a Go file with your validation rules, ensuring that your rules and code are always in sync. This is the recommended approach for most use cases.
    - **JSON-based Rules**: The library can also load rules from a JSON file at runtime, which is useful for dynamic environments.
- **Powered by CEL**: Veritas uses Google's Common Expression Language (CEL) for high-performance, dynamic validation.
- **Extensible**: Add your own custom validation functions to the CEL environment.
- **Type-Safe**: Designed to handle complex, nested data structures, including pointers, slices, and maps, with type-safety in mind.
- **Modern Go**: Built with Go 1.24 and `log/slog` for structured logging.

## Getting Started

### Option 1: Using Go Generate (Recommended)

The recommended approach is to use Go code generation for a type-safe, high-performance validation experience.

1.  **Annotate your structs:**

    Add validation rules using struct tags and special `// @cel:` comments.

    ```go
    // file: models/user.go
    package models

    //go:generate go run github.com/podhmo/veritas/cmd/veritas -gen-type=User -out-name=veritas_gen.go

    // @cel: self.Password == self.PasswordConfirm
    type User struct {
        Name     string `validate:"nonzero,cel:self.size() < 50"`
        Email    string `validate:"nonzero,email"`
        Age      int    `validate:"cel:self >= 18"`
        Password string `validate:"nonzero,cel:self.size() >= 10"`
        PasswordConfirm string `validate:"nonzero"`
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

### Option 2: Using JSON Rules

For more dynamic environments, you can load validation rules from a JSON file.

1.  **Define your rules in a JSON file:**

    Create a `rules.json` file that defines the validation rules for your types. The key is the type name (e.g., `main.User`), and the value contains the validation rules.

    ```json
    // file: rules.json
    {
      "main.User": {
        "fieldRules": {
          "Name": [
            "size(self) > 0"
          ],
          "Email": [
            "self.contains('@')"
          ]
        }
      }
    }
    ```

2.  **Load the rules and use the validator:**

    In your application, use `veritas.NewValidatorFromJSONFile()` to create a validator from your JSON file. You'll also need to provide a `TypeAdapter` to convert your Go types into a map that the validator can understand.

    ```go
    // file: main.go
    package main

    import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/podhmo/veritas"
    )

    type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
    }

    func main() {
        // setup validator
        v, err := veritas.NewValidatorFromJSONFile("./rules.json", veritas.WithTypeAdapters(
            map[string]veritas.TypeAdapter{
                "main.User": func(ob any) (map[string]any, error) {
                    v, ok := ob.(User)
                    if !ok {
                        return nil, errors.New("unexpected type")
                    }
                    return map[string]any{
                        "Name":  v.Name,
                        "Email": v.Email,
                    }, nil
                },
            },
        ))
        if err != nil {
            log.Fatalf("Failed to create validator: %v", err)
        }

        // ... (rest of the http server example)
    }
    ```

## Documentation

For more detailed information, please see the [documentation website](docs/website/index.md).

For a quick reference of common CEL expressions, see the [**CEL Cheatsheet**](docs/website/cel-cheatsheet.md). For a complete guide, please refer to the official [**CEL language definition**](https://github.com/google/cel-spec/blob/master/doc/langdef.md).

## Contributing

We welcome contributions! Please follow these steps to contribute:

1.  **Fork the repository.**
2.  **Create a new branch.**
3.  **Make your changes.**
4.  **Format your code:**
    ```bash
    goimports -w ./...
    ```
5.  **Run the tests:**
    ```bash
    go test ./...
    go test -C ./examples/http-server ./...
    ```
6.  **Submit a pull request.**

## Project Name

The name "Veritas" is Latin for "truth." It was chosen to symbolize the library's core function: to verify the "truthfulness" or correctness of data against a defined set of rules.
