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
	const pkgPrefix = "github.com/podhmo/veritas/testdata/sources."
	want := map[string]veritas.ValidationRuleSet{
		pkgPrefix + "Base": {
			FieldRules: map[string][]string{
				"ID": {`self != "" && self.size() > 1`},
			},
		},
		pkgPrefix + "EmbeddedUser": {
			FieldRules: map[string][]string{
				"ID":   {`self != "" && self.size() > 1`},
				"Name": {`self != ""`},
			},
		},
		pkgPrefix + "MockUser": {
			TypeRules: []string{"self.Age >= 18"},
			FieldRules: map[string][]string{
				"Name":  {`self != ""`},
				"Email": {`self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`},
				"ID":    {`self != null`},
			},
		},
		pkgPrefix + "MockVariety": {
			FieldRules: map[string][]string{
				"Count":    {"self != 0"},
				"IsActive": {"self"},
				"Scores":   {"self.size() > 0"},
				"Metadata": {"self.size() > 0"},
			},
		},
		pkgPrefix + "MockComplexData": {
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
		pkgPrefix + "MockMoreComplexData": {
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
		pkgPrefix + "ComplexUser": {
			FieldRules: map[string][]string{
				"Name":   {`self != ""`},
				"Scores": {`self.all(x, x >= 0)`},
			},
		},
		pkgPrefix + "Profile": {
			FieldRules: map[string][]string{
				"Platform": {`self != ""`},
				"Handle":   {`self != "" && self.size() > 2`},
			},
		},
		pkgPrefix + "UserWithProfiles": {
			FieldRules: map[string][]string{
				"Name": {`self != ""`},
			},
		},
		pkgPrefix + "Box[T]": {
			TypeRules: []string{"self.Value != null"},
			FieldRules: map[string][]string{
				"Value": {`self != null`},
			},
		},
		pkgPrefix + "Item": {
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

func TestRun_mainPackage(t *testing.T) {
	// Create a temporary directory for test output.
	tempDir, err := os.MkdirTemp("", "veritas-test-main-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define input and output paths.
	inPath := "./testdata/mainpkg"
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

	// Define the expected output. We expect the package path, not "main".
	// The package path is relative to the module root.
	expectedKey := "github.com/podhmo/veritas/cmd/veritas/testdata/mainpkg.User"
	want := map[string]veritas.ValidationRuleSet{
		expectedKey: {
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
