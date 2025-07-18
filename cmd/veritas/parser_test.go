package main

import (
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/veritas"
)

func TestParser(t *testing.T) {
	t.Run("parse struct with tags and comments", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		p := NewParser(logger)

		// The key is now the fully qualified type name
		want := map[string]veritas.ValidationRuleSet{
			"sources.MockUser": {
				TypeRules: []string{"self.Age >= 18"},
				FieldRules: map[string][]string{
					"Name":  {`self != ""`},
					"Email": {`self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`},
					"ID":    {`self != nil`},
				},
			},
			"sources.MockVariety": {
				FieldRules: map[string][]string{
					"Count":    {"self != 0"},
					"IsActive": {"self"},
					"Scores":   {"self.size() > 0"},
					"Metadata": {"self.size() > 0"},
				},
			},
			"sources.MockComplexData": {
				FieldRules: map[string][]string{
					"UserEmails": {`self.all(x, x.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$'))`},
					"ResourceMap": {
						`self.all(k, k.startsWith('id_'))`,
						`self.all(v, v != nil)`,
					},
					"Users":  {`self.all(x, x != nil)`},
					"Matrix": {`self.all(x, x.all(x, x != 0))`},
				},
			},
		}

		// Parse the directory containing the test file.
		got, err := p.Parse("../../testdata/sources")
		if err != nil {
			t.Fatalf("Parse() error = %v, want nil", err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
		}
	})
}
