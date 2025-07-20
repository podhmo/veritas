package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/podhmo/veritas"
	"github.com/podhmo/veritas/examples/gencode/def"
	_ "github.com/podhmo/veritas/examples/gencode/validation"
)

type User struct {
	Name  string
	Email string
}

func main() {
	if err := run(); err != nil {
		slog.Error("toplevel error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	// To validate `main.User` with rules for `def.User`, we need a way to map them.
	// The adapter pattern was one way. Without it, the simplest way is to use the
	// original type `def.User` directly, or modify the validation logic to support
	// type aliasing. For this example, we'll just use the `def.User` to show the
	// validator works with generated rules.
	v, err := veritas.NewValidator(
		veritas.WithTypes(def.User{}),
	)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// valid
	{
		user := def.User{Name: "foo", Email: "foo@example.com"}
		if err := v.Validate(ctx, user); err != nil {
			return fmt.Errorf("validation failed, unexpectedly: %+v", err)
		}
		slog.Info("validation ok", "user", user)
	}

	// invalid
	{
		user := def.User{Name: "foo", Email: "foo"}
		if err := v.Validate(ctx, user); err != nil {
			slog.Info("validation failed, as expected", "user", user, "err", err)
		} else {
			return fmt.Errorf("validation succeeded, unexpectedly")
		}
	}
	return nil
}
