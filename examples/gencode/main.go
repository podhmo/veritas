package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/podhmo/veritas"
	_ "github.com/podhmo/veritas/examples/gencode/def"
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
	v, err := veritas.NewValidator(
		veritas.WithTypeAdapters(
			map[string]veritas.TypeAdapter{
				"def.User": func(ob any) (map[string]any, error) {
					v, ok := ob.(User)
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
		veritas.WithTypeNames(map[string]string{
			"gencode.User": "def.User",
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// valid
	{
		user := User{Name: "foo", Email: "foo@example.com"}
		if err := v.Validate(ctx, user); err != nil {
			return fmt.Errorf("validation failed, unexpectedly: %+v", err)
		}
		slog.Info("validation ok", "user", user)
	}

	// invalid
	{
		user := User{Name: "foo", Email: "foo"}
		if err := v.Validate(ctx, user); err != nil {
			slog.Info("validation failed, as expected", "user", user, "err", err)
		} else {
			return fmt.Errorf("validation succeeded, unexpectedly")
		}
	}
	return nil
}
