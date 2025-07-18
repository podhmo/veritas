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
			"sources.Base": {
				FieldRules: map[string][]string{
					"ID": {`self != "" && self.size() > 1`},
				},
			},
			"sources.EmbeddedUser": {
				FieldRules: map[string][]string{
					"ID":   {`self != "" && self.size() > 1`},
					"Name": {`self != ""`},
				},
			},
			"sources.MockUser": {
				TypeRules: []string{"self.Age >= 18"},
				FieldRules: map[string][]string{
					"Name":  {`self != ""`},
					"Email": {`self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`},
					"ID":    {`self != null`},
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
						`self.all(v, v != null)`,
					},
					"Users":  {`self.all(x, x != null)`},
					"Matrix": {`self.all(x, x.all(x, x != 0))`},
				},
			},
			"sources.MockMoreComplexData": {
				FieldRules: map[string][]string{
					"ListOfMaps": {
						`self.all(x, x.size() > 0 && x.all(k, k.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')) && x.all(v, v != ""))`,
					},
					"MapOfSlices": {
						`self.all(k, k != "")`,
						`self.all(v, v.all(x, x != ""))`,
					},
				},
			},
		"sources.ComplexUser": {
			FieldRules: map[string][]string{
				"Name":   {`self != ""`},
				"Scores": {`self.all(x, x >= 0)`},
			},
		},
		"sources.Profile": {
			FieldRules: map[string][]string{
				"Platform": {`self != ""`},
				"Handle":   {`self != "" && self.size() > 2`},
			},
		},
		"sources.UserWithProfiles": {
			FieldRules: map[string][]string{
				"Name": {`self != ""`},
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
