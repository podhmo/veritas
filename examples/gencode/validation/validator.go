package def

import (
	"github.com/podhmo/veritas"
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
