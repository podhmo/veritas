package main

//go:generate go run github.com/podhmo/veritas/cmd/veritas -inject=main.go .

import (
	"context"
	"fmt"
	"log"

	"github.com/podhmo/veritas"
)

type User struct {
	Name string `json:"name" validate:"nonzero"`
	Age  int    `json:"age" validate:"nonzero"`
}

func main() {
	ctx := context.Background()
	// TODO: uncomment after running go generate
	// validator, err := veritas.NewValidator(
	validator, err := veritas.NewValidator(
		veritas.WithTypes(GetKnownTypes()...),
	)
	if err != nil {
		log.Fatalf("failed to create validator: %+v", err)
	}

	// valid user
	validUser := User{Name: "Alice", Age: 20}
	if err := validator.Validate(ctx, validUser); err != nil {
		log.Fatalf("validation failed for valid user: %+v", err)
	}
	fmt.Println("Validation successful for valid user")

	// invalid user (name)
	invalidUser_name := User{Name: "", Age: 20}
	if err := validator.Validate(ctx, invalidUser_name); err == nil {
		log.Fatal("validation should have failed for invalid user (name)")
	} else {
		fmt.Printf("Validation failed as expected for invalid user (name): %v\n", err)
	}

	// invalid user (age)
	invalidUser_age := User{Name: "Bob", Age: 0}
	if err := validator.Validate(ctx, invalidUser_age); err == nil {
		log.Fatal("validation should have failed for invalid user (age)")
	} else {
		fmt.Printf("Validation failed as expected for invalid user (age): %v\n", err)
	}
}

func setupValidation() {
	veritas.Register("github.com/podhmo/veritas/examples/codegen-onefile.User", veritas.ValidationRuleSet{
		FieldRules: map[string][]string{
			"Age": {
				`self != 0`,
			},
			"Name": {
				`self != ""`,
			},
		},
	})
}

// GetKnownTypes returns a list of all types that have validation rules.
func GetKnownTypes() []any {
	return []any{
		User{},
	}
}

func init() {
	setupValidation()
}
