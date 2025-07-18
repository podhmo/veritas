package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/veritas"
)

func TestRun(t *testing.T) {
	// Setup: Create a temporary directory for the output file.
	tmpDir, err := os.MkdirTemp("", "veritas-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Define input and output paths.
	inPath := "../../testdata/sources"
	outFile := filepath.Join(tmpDir, "rules.json")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Execute the run function.
	if err := run(inPath, outFile, logger); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	// Verify the output file was created.
	_, err = os.Stat(outFile)
	if os.IsNotExist(err) {
		t.Fatalf("run() did not create the output file: %s", outFile)
	}

	// Verify the content of the output file.
	gotBytes, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var got map[string]veritas.ValidationRuleSet
	if err := json.Unmarshal(gotBytes, &got); err != nil {
		t.Fatalf("Failed to unmarshal output JSON: %v", err)
	}

	// Define the expected output.
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

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("run() output mismatch (-want +got):\n%s", diff)
	}
}
