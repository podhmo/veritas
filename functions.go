package veritas

import (
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// DefaultFunctions returns a list of common, useful custom functions for CEL.
func DefaultFunctions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function("strings.ToUpper",
			cel.Overload("upper_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(s ref.Val) ref.Val {
					return types.String(strings.ToUpper(s.(types.String).Value().(string)))
				}),
			),
		),
		cel.Function("custom.matches",
			cel.Overload("matches_string_pattern",
				[]*cel.Type{cel.StringType, cel.StringType},
				cel.BoolType,
				cel.BinaryBinding(func(s, p ref.Val) ref.Val {
					str := s.(types.String).Value().(string)
					pattern := p.(types.String).Value().(string)
					matched, err := regexp.MatchString(pattern, str)
					if err != nil {
						return types.NewErr("regexp compilation error: %s", err)
					}
					return types.Bool(matched)
				}),
			),
		),
	}
}

func DefaultEnvOptions() []cel.EnvOption {
	return append(DefaultFunctions(), cel.StdLib())
}
