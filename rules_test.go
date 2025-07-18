package veritas

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestJSONRuleProvider(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		// Create a temporary JSON file for testing.
		content := `{
			"user.User": {
				"typeRules": ["self.Password == self.PasswordConfirm"],
				"fieldRules": {
					"Email": ["required", "email"]
				}
			}
		}`
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "rules.json")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		provider := NewJSONRuleProvider(filePath)
		got, err := provider.GetRuleSets()
		if err != nil {
			t.Fatalf("GetRuleSets() failed: %v", err)
		}

		want := map[string]ValidationRuleSet{
			"user.User": {
				TypeRules: []string{"self.Password == self.PasswordConfirm"},
				FieldRules: map[string][]string{
					"Email": {"required", "email"},
				},
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("GetRuleSets() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		provider := NewJSONRuleProvider("nonexistent.json")
		_, err := provider.GetRuleSets()
		if err == nil {
			t.Fatal("Expected an error for a nonexistent file, but got nil")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(filePath, []byte("{invalid"), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}
		provider := NewJSONRuleProvider(filePath)
		_, err := provider.GetRuleSets()
		if err == nil {
			t.Fatal("Expected an error for invalid JSON, but got nil")
		}
	})
}
