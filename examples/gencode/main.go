package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/cel-go/cel"
	"github.com/podhmo/veritas"
	"github.com/podhmo/veritas/examples/gencode/def"
	_ "github.com/podhmo/veritas/examples/gencode/validation"
)

func main() {
	if err := run(); err != nil {
		slog.Error("toplevel error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	engine, err := veritas.NewEngine(slog.Default(), append(veritas.DefaultFunctions(), cel.StdLib())...)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	v, err := veritas.NewValidator(
		veritas.WithEngine(engine),
		veritas.WithTypeAdapters(
			map[string]veritas.TypeAdapter{
				"github.com/podhmo/veritas/examples/gencode/def.User": func(ob any) (map[string]any, error) {
					v, ok := ob.(def.User)
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
