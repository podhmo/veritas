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

// TestRun is an integration-style test for the main `run` function.
func TestRun(t *testing.T) {
	// Create a temporary directory for test output.
	tempDir, err := os.MkdirTemp("", "veritas-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define input and output paths.
	inPath := "../../testdata/sources"
	outFile := filepath.Join(tempDir, "rules.json")

	// Logger for the test.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Execute the main logic.
	if err := run(inPath, outFile, logger); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Read the generated file.
	gotBytes, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Unmarshal the result into the expected structure.
	var got map[string]veritas.ValidationRuleSet
	if err := json.Unmarshal(gotBytes, &got); err != nil {
		t.Fatalf("Failed to unmarshal generated JSON: %v", err)
	}

	// Define the expected output.
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

	// Compare the actual result with the expected result.
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("run() output mismatch (-want +got):\n%s", diff)
	}
}
