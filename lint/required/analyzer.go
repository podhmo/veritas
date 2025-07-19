package required

import (
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var Analyzer = &analysis.Analyzer{
	Name:     "required",
	Doc:      "check for invalid usage of 'required' tag",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			structType, ok := n.(*ast.StructType)
			if !ok {
				return true
			}

			for _, field := range structType.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
				validateTag, ok := tag.Lookup("validate")
				if !ok {
					continue
				}

				hasRequired := false
				rules := strings.Split(validateTag, ",")
				for _, rule := range rules {
					if rule == "required" {
						hasRequired = true
						break
					}
				}

				if hasRequired {
					tv := pass.TypesInfo.TypeOf(field.Type)
					if _, ok := tv.Underlying().(*types.Pointer); !ok {
						pass.Reportf(field.Pos(), "'required' tag can only be used with pointer types")
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
