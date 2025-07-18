package main

import (
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/podhmo/veritas"
)

func TestParser(t *testing.T) {
	t.Run("parse struct with tags and comments", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		p := NewParser(logger)

		want := map[string]veritas.ValidationRuleSet{
			"MockUser": {
				TypeRules: []string{
					"self.Age >= 18",
				},
				FieldRules: map[string][]string{
					"Name":  {"required"},
					"Email": {"required", "email"},
				},
			},
		}

		got, err := p.Parse("../../testdata/sources/user.go")
		require.NoError(t, err)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
		}
	})
}
