package main

import (
	"flag"

	"github.com/gostaticanalysis/codegen/singlegenerator"
	"github.com/podhmo/veritas/cmd/veritas/gen"
	"github.com/podhmo/veritas/lint"
	"github.com/podhmo/veritas/lint/required"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	gen.Generator.Flags.VisitAll(func(f *flag.Flag) {
		flag.CommandLine.Var(f.Value, f.Name, f.Usage)
	})

	var lintFlag bool
	flag.BoolVar(&lintFlag, "lint", false, "run linter")
	flag.Parse()

	if lintFlag {
		multichecker.Main(
			lint.Analyzer,
			required.Analyzer,
		)
		return
	}

	singlegenerator.Main(gen.Generator)
}
