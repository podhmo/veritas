package validation

import (
	veritas "github.com/podhmo/veritas"
	def "github.com/podhmo/veritas/examples/gencode/def"
)

func init() {
	veritas.Register("github.com/podhmo/veritas/examples/gencode/def.User", veritas.ValidationRuleSet{
		FieldRules: map[string][]string{
			"Email": {
				`custom.matches(self, '^[^@]+@[^@]+$')`,
			},
			"Name": {
				`self.size() > 0`,
			},
		},
	})
}

// GetKnownTypes returns a list of all types that have validation rules.
func GetKnownTypes() []any {
	return []any{
		def.User{},
	}
}
