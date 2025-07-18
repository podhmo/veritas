package veritas

import (
	"encoding/json"
	"os"
)

// ValidationRuleSet holds all validation rules for a single Go type.
type ValidationRuleSet struct {
	TypeRules  []string            `json:"typeRules"`
	FieldRules map[string][]string `json:"fieldRules"`
}

// RuleProvider is the interface for any component that can supply validation rules.
type RuleProvider interface {
	GetRuleSets() (map[string]ValidationRuleSet, error)
}

// JSONRuleProvider loads validation rules from a JSON file.
// It implements the RuleProvider interface.
type JSONRuleProvider struct {
	filePath string
}

// NewJSONRuleProvider creates a new provider that reads from the specified file path.
func NewJSONRuleProvider(filePath string) *JSONRuleProvider {
	return &JSONRuleProvider{filePath: filePath}
}

// GetRuleSets reads and parses the JSON file into a map of validation rule sets.
func (p *JSONRuleProvider) GetRuleSets() (map[string]ValidationRuleSet, error) {
	bytes, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, err
	}
	var rules map[string]ValidationRuleSet
	if err := json.Unmarshal(bytes, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}