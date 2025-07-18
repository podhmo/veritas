package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/veritas"
)

// TestEndToEnd runs the veritas CLI tool and verifies its output.
func TestEndToEnd(t *testing.T) {
	// Define paths
	tempDir := t.TempDir()
	outputJSONPath := filepath.Join(tempDir, "test_rules.json")
	inputPath := "../../testdata/sources/..." // Use recursive path

	// Build the veritas binary
	buildCmd := exec.Command("go", "build", "-o", "veritas_test_binary")
	buildCmd.Dir = "." // Run in the main package directory
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build veritas binary: %v\nOutput: %s", err, string(output))
	}
	defer os.Remove("veritas_test_binary")

	// Run the generated binary
	runCmd := exec.Command("./veritas_test_binary", "-in", inputPath, "-out", outputJSONPath)
	if output, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("Veritas CLI failed: %v\nOutput: %s", err, string(output))
	}

	// Read the actual output
	actualJSON, err := os.ReadFile(outputJSONPath)
	if err != nil {
		t.Fatalf("Failed to read output JSON: %v", err)
	}

	// Define the expected output
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

	// Unmarshal actual JSON to compare with the 'want' map
	var got map[string]veritas.ValidationRuleSet
	if err := json.Unmarshal(actualJSON, &got); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	// The double backslash in the Go string literal for the regex (`\\.`) is preserved
	// as `\\` by the JSON marshaller. The test's `want` map also has `\\.` in its
	// string literal, so no transformation is needed for the comparison.

	// Custom comparer to sort slices before comparing.
	opts := cmp.Options{
		cmp.Transformer("Sort", func(s []string) []string {
			sort.Strings(s)
			return s
		}),
	}
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("Generated JSON mismatch (-want +got):\n%s", diff)
	}
}
