package main

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/podhmo/veritas"
)

func TestGoCodeGenerator(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	generator := NewGoCodeGenerator(logger)

	ruleSets := map[string]veritas.ValidationRuleSet{
		"main.User": {
			TypeRules: []string{"self.Name != ''"},
			FieldRules: map[string][]string{
				"Email": {"self.matches('^[^@]+@[^@]+$')"},
				"Age":   {"self > 18"},
			},
		},
		"main.Post": {
			FieldRules: map[string][]string{
				"Title":   {"self != ''"},
				"Content": {"self.size() > 10"},
			},
		},
	}

	var buf bytes.Buffer
	err := generator.Generate("main", ruleSets, &buf)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// For now, we'll just check that the output is not empty.
	// A more robust test would parse the generated Go code and verify its structure.
	if buf.Len() == 0 {
		t.Errorf("Generate() output is empty")
	}

	// You can uncomment this to see the generated code during testing.
	// t.Log(buf.String())
}
