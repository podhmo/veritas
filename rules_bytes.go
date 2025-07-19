package veritas

import (
	"encoding/json"
)

// BytesRuleProvider loads validation rules from a byte slice.
// It implements the RuleProvider interface.
type BytesRuleProvider struct {
	bytes []byte
}

// NewBytesRuleProvider creates a new provider that reads from the specified byte slice.
func NewBytesRuleProvider(bytes []byte) *BytesRuleProvider {
	return &BytesRuleProvider{bytes: bytes}
}

// GetRuleSets parses the byte slice into a map of validation rule sets.
func (p *BytesRuleProvider) GetRuleSets() (map[string]ValidationRuleSet, error) {
	var rules map[string]ValidationRuleSet
	if err := json.Unmarshal(p.bytes, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}
