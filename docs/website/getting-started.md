# Getting Started

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
