package main

import (
	veritas "github.com/podhmo/veritas"
)

func setupValidation() {
	veritas.Register("github.com/podhmo/veritas/examples/codegen-onefile.User", veritas.ValidationRuleSet{
		FieldRules: map[string][]string{
			"Age": {
				`self != 0`,
			},
			"Name": {
				`self != ""`,
			},
		},
	})
}

func init() {
	setupValidation()
}

// GetKnownTypes returns a list of all types that have validation rules.
func GetKnownTypes() []any {
	return []any{
		User{},
	}
}
