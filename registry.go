package veritas

import "sync"

var (
	globalRegistry     = make(map[string]ValidationRuleSet)
	globalRegistryLock sync.RWMutex
)

// Register adds a validation rule set to the global registry.
// This function is intended to be called from init() functions in generated code.
func Register(typeName string, ruleSet ValidationRuleSet) {
	globalRegistryLock.Lock()
	defer globalRegistryLock.Unlock()
	globalRegistry[typeName] = ruleSet
}

// Unregister removes a validation rule set from the global registry.
// This is mainly useful for testing purposes.
func Unregister(typeName string) {
	globalRegistryLock.Lock()
	defer globalRegistryLock.Unlock()
	delete(globalRegistry, typeName)
}

// UnregisterAll removes all validation rule sets from the global registry.
// This is mainly useful for testing purposes.
func UnregisterAll() {
	globalRegistryLock.Lock()
	defer globalRegistryLock.Unlock()
	globalRegistry = make(map[string]ValidationRuleSet)
}

// NewRuleProviderFromRegistry creates a new RuleProvider that uses the global registry.
func NewRuleProviderFromRegistry() RuleProvider {
	return &registryProvider{}
}

type registryProvider struct{}

// GetRuleSets returns a copy of the global registry.
func (p *registryProvider) GetRuleSets() (map[string]ValidationRuleSet, error) {
	globalRegistryLock.RLock()
	defer globalRegistryLock.RUnlock()

	// Return a copy to prevent race conditions if the caller modifies the map.
	rules := make(map[string]ValidationRuleSet, len(globalRegistry))
	for k, v := range globalRegistry {
		rules[k] = v
	}
	return rules, nil
}
