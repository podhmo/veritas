package main

import (
	"log/slog"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/veritas"
)

func TestParser(t *testing.T) {
	t.Run("parse struct with tags and comments", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		p := NewParser(logger)

		want := map[string]veritas.ValidationRuleSet{
			"MockUser": {
				TypeRules: []string{
					"self.Age >= 18",
				},
				FieldRules: map[string][]string{
					"Name":  {"size(self.Name) > 0"},
					"Email": {"size(self.Email) > 0", "self.Email.matches('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$')"},
				},
			},
		}

		// Note: The path is relative to the `veritas` package root.
		got, err := p.Parse("../../testdata/sources/user.go")
		if err != nil {
			t.Fatalf("Parse() error = %v, want nil", err)
		}

		// Custom comparer to sort slices before comparing, making tests robust against element order changes.
		opts := cmp.Options{
			cmp.Transformer("Sort", func(s []string) []string {
				sort.Strings(s)
				return s
			}),
		}
		if diff := cmp.Diff(want, got, opts...); diff != "" {
			t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
		}
	})
}
