package main

import (
	"github.com/gostaticanalysis/codegen/singlegenerator"
	"github.com/podhmo/veritas/cmd/veritas/gen"
)

func main() {
	singlegenerator.Main(gen.Generator)
}
