package main

import (
	"flag"

	"github.com/gostaticanalysis/codegen/singlegenerator"
	"github.com/podhmo/veritas/cmd/veritas/gen"
	lintcmd "github.com/podhmo/veritas/cmd/veritas/lint"
)

func main() {
	var lintFlag bool
	flag.BoolVar(&lintFlag, "lint", false, "run linter")
	flag.Parse()

	if lintFlag {
		lintcmd.Main()
		return
	}

	singlegenerator.Main(gen.Generator)
}
