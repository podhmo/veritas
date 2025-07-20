package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/podhmo/veritas"
	"github.com/podhmo/veritas/examples/gencode/def"
	"github.com/podhmo/veritas/examples/gencode/validation"
)

func main() {
	if err := run(); err != nil {
		slog.Error("toplevel error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	v, err := veritas.NewValidator(
		veritas.WithTypes(validation.GetKnownTypes()...),
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
