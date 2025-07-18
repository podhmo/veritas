package lint

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "veritas",
	Doc:  "veritas is a linter for veritas rules",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// TODO: implement lint logic
			return true
		})
	}
	return nil, nil
}
